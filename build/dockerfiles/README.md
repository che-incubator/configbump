# Dockerfile Clarification

**Dockerfile** is the Eclipse Che build file, used in this repo to publish to [quay.io/che-incubator/configbump](https://quay.io/repository/che-incubator/configbump?tab=tags).

**rhel.Dockerfile** is the build file for the [Red Hat OpenShift Devspaces](https://github.com/redhat-developer/devspaces-images/tree/devspaces-3-rhel-8/devspaces-configbump) image, which can be run locally.

**brew.Dockerfile** is a variation on `rhel.Dockerfile` specifically for Red Hat builds, and cannot be run locally as is.