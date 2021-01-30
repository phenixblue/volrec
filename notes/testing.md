# Notes on Testing the volrec Controller

## Create Test Namespace

```shell
$ kubectl create ns test1
```

## Label Namespace

```shell
$ kubectl label ns test1 k8s.twr.dev/owner="jmsearcy"
```

## Deploy Test Statefulset

```shell
$ kubectl apply -f ./testing/kubernetes/crdb-sts.yaml -n test1