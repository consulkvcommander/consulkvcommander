# ConsulKVCommander

## Introduction

Welcome to ConsulKVCommander, a cutting-edge Kubernetes operator designed to seamlessly manage and secure HashiCorp's Consul Key-Value (KV) stores within Kubernetes environments. This tool is crafted to enhance the operational efficiency, security, and compliance of your KV data management, making it an indispensable asset for cloud-native applications.

<div align="center"><img src="./assets/logo-official.png"  width="40%" height="60%"></div>

## Key Features:
- **Self-Adaptive System:** At its core, ConsulKVCommander employs advanced self-adaptive mechanisms, including self-healing and self-protecting functionalities, to manage sensitive data dynamically. It ensures that your KV pairs are consistently monitored, and any potential security risks are promptly addressed.
- **Efficient Data Management:** By synchronizing ConsulKV data with Kubernetes ConfigMaps, ConsulKVCommander provides an intuitive way to handle configuration data, ensuring it is always up-to-date and accessible within your Kubernetes cluster.
- **Enhanced Security:** With its focus on security, ConsulKVCommander vigilantly guards against exposing sensitive data. It employs sophisticated algorithms to detect and manage sensitive information, thereby maintaining the confidentiality and integrity of your data.
- **Resource Optimization:** The integrated Guardian container optimizes resource utilization, ensuring that the operator runs efficiently without overusing or underutilizing cluster resources.
- **Real-Time Auditing and Alerting:** ConsulKVCommander supports real-time auditing through managed CSV sheets in an S3 bucket and integrates with PagerDuty for instant alerting, keeping you informed about the state of your KV pairs at all times.

Whether you're managing large-scale cloud-native applications or looking for a robust solution for your ConsulKV data, ConsulKVCommander offers a reliable and secure way to streamline your data management processes. Join us in exploring the capabilities of ConsulKVCommander and see how it can transform your Kubernetes data management experience.

## Getting Started

### Prerequisites
- go version v1.20.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/consulkv-commander:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. 
And it is required to have access to pull the image from the working environment. 
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/consulkv-commander:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin 
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Contributing

Please refer to [COTRIBUTING.md](./CONTRIBUTING.md) for the contribution guidelines.

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

