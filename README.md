# volrec

`Vol`ume `Re`claim `C`ontroller, or `volrec` for short, is a Kubernetes Controller that allows for empowering Developers with the ability to control the Reclaim Policy on their persistent volumes without providing them with cluster scoped permissions.

**NOTE: While this project does work, this was a quick experiemnt, it's not polished, and you probably shouldn't run this in production!**

## Overview

`volrec` contains a set of controllers operating under a single Controller Manager.

- PersistentVolume Controller
- PersistentVolumeClaim Controller
- Namespace Controller

Each is resposible for handling reconciliation actions for a given target resource.

The mapping relationship between a namespace scoped PVC bound to a cluster scoped PV provides the relationship to enable end-users to control the Reclaim Policy for their volumes through the application of labels on the PVC resource the user has access to.

Add the `to-do: add label name` label to a PVC within your namespace and `volrec` will follow the mapping to the appropriate PV and set the Reclaim Policy according to the value of the label. A validating Admission Controller is setup to make sure only supported values for the Volume Reclaim policy can be set wihtin the label.


## Installation

## Troubleshooting

## Contributing
