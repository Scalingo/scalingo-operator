# scalingo-operator

Scalingo operator for Kubernetes.


# Usage




# Development

## Install environment

```sh
sudo snap install microk8s --classic
sudo snap install kubectl --classic


microk8s.kubectl config view --raw > $HOME/.kube/microk8s.config

# then add in ~/.zshrc
export  KUBECONFIG=$HOME/.kube/config
export  KUBECONFIG=$KUBECONFIG:$HOME/.kube/microk8s.config

# Verification: both commands must return the same informations
microk8s.kubectl config view
kubectl config view
```

## download kubebuilder and install locally.
```sh
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```

## Kubebuilder Commands

### Operator creation example
```sh
kubebuilder init --domain scalingo.com --repo github.com/Scalingo/scalingo-operator
```

### Build commands

```sh
# generate api/v1alpha/zz_generated.deepcopy.go
make generate

# generate the CRD manifests under config/crd/bases and a sample for it under config/samples
make manifests

# execute the CRD
make run
```

### Deployment commands

```sh
# deploy the CRD in the cluster
make install

# build and push your image to the location specified by IMG
make docker-build docker-push IMG=<some-registry>/<project-name>:tag

# deploy the controller to the cluster with image specified by IMG
make deploy IMG=<some-registry>/<project-name>:tag
```
