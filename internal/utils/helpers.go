package utils

import "golang.org/x/exp/constraints"

func MaxInSlice[K constraints.Ordered](slice []K) (K, bool) {
	if len(slice) == 0 {
		var zeroVal K
		return zeroVal, false
	}
	curMax := slice[0]
	for _, item := range slice {
		if item > curMax {
			curMax = item
		}
	}
	return curMax, true
}

func ValueInSlice[K comparable](value K, slice []K) bool {
	for _, item := range slice {
		if value == item {
			return true
		}
	}
	return false
}

func ZeroValue[K any](value *K) K {
	var zero K
	return zero
}

func DeepCopyMap[K comparable, V any](src map[K]V) map[K]V {
	new := map[K]V{}
	for k, v := range src {
		new[k] = v
	}
	return new
}

func FromPtr[K any](value *K) K {
	if value == nil {
		var zero K
		return zero
	}
	return *value
}

func ToPtr[K any](value K) *K {
	return &value
}

func RemoveKeysFromMap[K comparable, V any](src map[K]V, keys []K) map[K]V {
	deepCopy := DeepCopyMap(src)
	for k := range deepCopy {
		if ValueInSlice(k, keys) {
			delete(deepCopy, k)
		}
	}
	return deepCopy
}
