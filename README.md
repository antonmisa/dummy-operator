# dummy-operator
The goal is to use the Operator SDK to write a small Kubernetes Custom Controller in
Go and then deploy it on a Kubernetes cluster.

## Description
Example of Customer Resource:
```yaml
apiVersion: interview.com/v1alpha1
kind: Dummy
metadata:
    name: dummy1
    namespace: default
spec:
    message: "I'm just a dummy"
status:
    specEcho: "I'm just a dummy"
    podStatus: "Pending"
```

Where
1. specEcho  - copy the value of spec.message into status.specEcho by custom controller
2. podStatus - track of the status of the Pod (Phase) associated to the Dummy by custom controller

At a high-level, this is the flow sequence of the operator's functionality:

1. User creates a custom resource (CR) via kubectl command under a Kubernetes namespace.
2. Operator is running on the cluster under the operator's namespace and it watches for these specific custom resources (CR) object.
3. Operator takes action: create or delete Pods (not scaled).

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) or [minikube](https://minikube.sigs.k8s.io/docs/start/) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Developer Software

Install all the following:

- git client - Apple Xcode or any git command line 
- Go (v1.19+) - https://golang.org/dl/
- Docker Desktop - https://www.docker.com/products/docker-desktop
- Kind (Kubernetes in Docker) -  https://kind.sigs.K8s.io/docs/user/quick-start/
- Minikube -  https://minikube.sigs.k8s.io/docs/start/
- Operator SDK - https://sdk.operatorframework.io/docs/installation/
- Kubebuilder - https://go.kubebuilder.io/quick-start.html
- Kustomize - https://kubectl.docs.kubernetes.io/installation/
- [Optional but recommended] - Code editor such as
  [VScode](https://code.visualstudio.com/download) or
  [goland](https://www.jetbrains.com/go/download/)

### Running on the cluster
0. Getting image from [hub.docker.com](https://hub.docker.com/r/savelievant/dummy-operator)
- ```sh
docker pull savelievant/dummy-operator
``` 

or dowload sources from current repo [github.com](https://github.com/antonmisa/dummy-operator)
- ```sh
git clone https://github.com/antonmisa/dummy-operator
```

1. To start a minikube cluster on your local machine, run the following command, setting as an arbitrarily name for your cluster (this name will be used for kubectl context):
```sh
minikube start
```
Or using kind cluster on your local machine, run the following command, setting as an arbitrarily name for your cluster (this name will be used for kubectl context):
```sh
kind create cluster --name operator-dev
```

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/_v1alpha1_dummy2.yaml
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=interview.com/dummy-operator:v0.0.1
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=interview.com/dummy-operator:v0.0.1
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Run unit tests
```sh
make test
```

2. Run e2e tests:
```sh
make test-e2e
```

3. Install the CRDs into the cluster:

```sh
make install
```

4. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)