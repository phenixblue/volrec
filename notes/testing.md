# Notes on Testing the volrec Controller

## Create Test Namespace

```shell
$ kubectl create ns test1
```

## Label Namespace

```shell
$ kubectl label ns test1 k8s.twr.dev/owner="user1"
```

## Deploy Test Statefulset

```shell
$ kubectl apply -f ./testing/kubernetes/crdb-sts.yaml -n test1
```

## Check the logs of the controller

```shell
$ kubectl logs -n volrec-system -l app=volrec -c volrec -f
```

## Apply labels to PVC to change Reclaim Policy

```shell
$ kubectl label pvc -n test1 datadir-crdb1-0 storage.k8s.twr.dev/reclaim-policy=Retain --overwrite
```

NOTE: The default Reclaim Policy could differ, so update the label value to something pertinent for your environment

## Check logs and Persistent Volumes

```shell
$ kubectl logs -n volrec-system -l app=volrec -c volrec -f
$ k get pv --show-labels
```
