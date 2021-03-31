# Notes on Testing the volrec Controller

## Deploy Volrec

```shell
$ make deploy-prod
```

## Setup Test environment

This creates a test namespace, applies appropriate labels, and deploys a sample Statefulset with dynamic PVC/PV's.

The PVC claim wihtin the Statefulset is configured to set PV reclaim policy from to `Retain`.


```shell
$ make test-setup
```

## Check the logs of the controller

```shell
$ kubectl logs -n volrec-system -l app=volrec -c volrec -f
```

## Apply labels to PVC to change Reclaim Policy

```shell
$ make set-reclaim-delete
$ make set-reclaim-recycle
$ make set-reclaim-retain
```

NOTE: The default Reclaim Policy could differ, so update the label value to something pertinent for your environment

## Check logs and Persistent Volumes

```shell
$ kubectl logs -n volrec-system -l app=volrec -c volrec -f
$ k get pv --show-labels
```
