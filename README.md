# volrec

`Vol`ume `Re`claim `C`ontroller, or `volrec` for short, is a Kubernetes Controller that allows for empowering Developers with the ability to control the Reclaim Policy on their persistent volumes without providing them with cluster scoped permissions.

**NOTE: While this project does work, this was a quick experiment, it's not polished, and you probably shouldn't run this in production!**

## Overview

`volrec` contains a set of controllers operating under a single Controller Manager.

- PersistentVolume Controller
- PersistentVolumeClaim Controller
- Namespace Controller

Each is responsible for handling reconciliation actions for a given target resource.

The mapping relationship between a namespace scoped PVC bound to a cluster scoped PV provides the relationship to enable end-users to control the Reclaim Policy for their volumes through the application of labels on the PVC resource the user has access to.

Add the `storage.k8s.twr.dev/reclaim-policy` label with a valid Reclaim Policy for the value (ie. `Retain`, `Recycle`, or `Delete`) to a PVC within your namespace and `volrec` will follow the mapping to the appropriate PV and set the Reclaim Policy according to the value of the label. ~~A validating Admission Controller is setup to make sure only supported values for the Volume Reclaim policy can be set within the label.~~

## Configuration

`volrec` can be configured via flags/arguments passed at startup.

 |Flag              | Type      |    Default Value     |     Description   |
|---                |---        |---                   |---                |
| --metrics-addr    | string    | ":8081"              | The address the metric endpoint binds to.|
| --enable-leader-election      | bool  | false  | Enable leader election for controller manager to ensure there is only one active controller manager. |
| --reclaim-label   | string    | "storage.k8s.twr.dev/reclaim-policy"  | The label to use for tracking Persistent Volume reclaim policy.|
| --set-owner       | bool      | false | Toggle whether or not owner information from a given namespace is transferred to the Persistent Volume.|
| --owner-label     | string    | "k8s.twr.dev/owner"  | The Label to use to set owner information on a Persistent Volume.|
| --set-ns          | bool      | false | Toggle whether or not to add a label mapping Persistent Volumes back to a namespace.|
| --ns-label        | string    | "k8s.twr.dev/owning-namespace"    | The label to use for identifying an owning namespace on a Persistent Volume.|

## Installation

```shell
$ make deploy
```

## Troubleshooting

## Contributing
