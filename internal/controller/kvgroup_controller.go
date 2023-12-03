/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/secretengine"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"sync"
)

// KVGroupReconciler reconciles a KVGroup object
type KVGroupReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	SecretEngineClient secretengine.Client
	lock               *sync.Mutex
}

//+kubebuilder:rbac:groups=sas.com.sas.com,resources=kvgroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sas.com.sas.com,resources=kvgroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sas.com.sas.com,resources=kvgroups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KVGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *KVGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	_ = log.FromContext(ctx)

	var kvGroup sascomv1.KVGroup
	if err := r.Get(ctx, req.NamespacedName, &kvGroup); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	consulKvClient := utils.NewConsulKV(kvGroup.Spec.ConsulUrl)

	pathToWeights := map[string]int{}

	unvalidatedConfigMapPayload := map[string]string{}
	for _, pathSpec := range kvGroup.Spec.Paths {
		consulKvResponse, err := consulKvClient.GetPath(pathSpec.Path)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error occurred while GET-ing the consul key at the path %s: %w", pathSpec.Path, err)
		}

		for _, elem := range consulKvResponse {
			key := elem.Key
			base64EncodedValue := elem.Value
			decodedValueBytes, err := base64.StdEncoding.DecodeString(base64EncodedValue)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to decode the base64 value corresponding to the Key '%s': %w", key, err)
			}
			value := string(decodedValueBytes)
			key = strings.ReplaceAll(key, "/", ".")
			unvalidatedConfigMapPayload[key] = value

			pathToWeights[key] = pathSpec.CriticalityWeight
		}
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	sanitizedConfigMapPayload, err := r.SecretEngineClient.Run(&kvGroup, unvalidatedConfigMapPayload, pathToWeights)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to track any invalidations after the new reconciliation: %w", err)
	}

	configMapName, configMapNamespace := kvGroup.Name, kvGroup.Namespace
	desiredConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: configMapNamespace,
		},
		Data: sanitizedConfigMapPayload,
	}

	if err := controllerutil.SetControllerReference(&kvGroup, desiredConfigMap, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to setup controller reference on the ")
	}

	if err := r.reconcileConfigMap(ctx, desiredConfigMap); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile the ConfigMap: %w", err)
	}

	return ctrl.Result{}, r.updateStatus(req.NamespacedName, kvGroup.Status.DeepCopy())
}

func (r *KVGroupReconciler) updateStatus(kvGroupKey client.ObjectKey, newStatus *sascomv1.KVGroupStatus) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var obj sascomv1.KVGroup
		err := r.Get(context.Background(), kvGroupKey, &obj)
		if err != nil {
			return err
		}
		obj.Status = *newStatus
		return r.Status().Update(context.Background(), &obj)
	})
}

func (r *KVGroupReconciler) reconcileConfigMap(ctx context.Context, desiredConfigMap *v1.ConfigMap) error {

	var currentConfigMap v1.ConfigMap
	objectKey := client.ObjectKeyFromObject(desiredConfigMap)
	if err := r.Get(ctx, objectKey, &currentConfigMap); err != nil {
		if errors.IsNotFound(err) {
			if len(desiredConfigMap.Data) == 0 {
				return nil
			}
			return r.Create(ctx, desiredConfigMap)
		}
		return fmt.Errorf("error occurred while getting the configmap with the key %s: %w", objectKey, err)
	}
	if len(desiredConfigMap.Data) == 0 {
		return r.Delete(ctx, &currentConfigMap)
	}
	currentData, desiredData := currentConfigMap.Data, desiredConfigMap.Data
	currentOwnerRef, desiredOwnerRef := currentConfigMap.OwnerReferences, desiredConfigMap.OwnerReferences
	if reflect.DeepEqual(currentData, desiredData) && reflect.DeepEqual(currentOwnerRef, desiredOwnerRef) {
		return nil
	}

	if err := r.Update(ctx, desiredConfigMap); err != nil {
		return fmt.Errorf("error occurred while updating the configmap with the key %s: %w", objectKey, err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KVGroupReconciler) SetupWithManager(mgr ctrl.Manager, periodicConfigMapReconcilerChan chan event.GenericEvent) error {
	r.lock = &sync.Mutex{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&sascomv1.KVGroup{}).
		Owns(&v1.ConfigMap{}).
		WatchesRawSource(&source.Channel{Source: periodicConfigMapReconcilerChan}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
