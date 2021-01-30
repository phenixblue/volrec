# Steps to get up and running

Just tracking this as Kubebuilder gets a bit cantankerous when trying to work with core Kubernetes resources.

## Kubebuilder Init

```shell
$ kubebuilder init --domain storage.k8s.twr.dev --repo twr.dev/volrec --license apache2 --owner "The WebRoot"
```

## Add Controllers

```shell
$ kubebuilder create api --group "core" --version v1 --kind PersistentVolume --resource false
$ kubebuilder create api --group "core" --version v1 --kind PersistentVolumeClaim --resource false
$ kubebuilder create api --group "core" --version v1 --kind Namespace --resource false
```

## Generate manifests

This is sort of stupid to do, but there are a few files that don't get scaffolded by default and make it harder to test `kustomize build` for the additional cleanup steps.

```shell
$ make manifests
```

## Cleanup things for core K8s resources

You need to edit a handful of things after running the above comamnds to make things work for the core K8s API resources.

I'm basically running `make manifests` and `kustomize build config/default/` while doing the below steps to make sure things don't come back once cleaned up.

1. Remove custom API's/Types

    ```shell
    $ rm -rf ./api
    ```

1. Cleanup config scaffolding

    Some files I've left in place to try and not break all of the kuztomize scaffolding/make targets

    - Comment out resources within `./config/crd/kustomization.yaml` that are associated with the core resources above
    - Remove all `cainjection_*` and `webhook_*` patch files in the `./config/crd/patches` directory
    - Remove all editor/viewer RBAC roles within `config/rbac` that are associated with the core resources above
    - Remove all sample manifests in the `./config/samples` directory

1. Replace the references to the local package/api to reference the actual k8s core API/package

    Old: `corev1 "twr.dev/volrec/api/v1"`
    New: `corev1 "k8s.io/api/core/v1"` 

    In the following files:

    - ./main.go
    - ./controllers/namespace_controller.go
    - ./controllers/persistentvolume_controller.go
    - ./controllers/persistentvolumeclaim_controller.go
    - ./controllers/suite_test.go

1. Set Markers for core k8s resource RBAC in controller files

    By default you should see 2 lines like this in each controller file (`./controllers`):

    ```go
    // +kubebuilder:rbac:groups=core.storage.k8s.twr.dev,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
    // +kubebuilder:rbac:groups=core.storage.k8s.twr.dev,resources=namespaces/status,verbs=get;update;patch
    ```

    The above lines should be replaces with something like this:

    ```go
    // +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
    ```

