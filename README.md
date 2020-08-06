# configbump

This is a simple Kubernetes controller that is able to quickly synchronize a set of configmaps (selected using labels) to files
on local filesystem.

Additionally, this tool can send a signal to another process (hence the "bump" in the name).

The combination of the two capabilities can be used to enable dynamic reconfiguration or restarts on configuration change for any program that is able to respond to the signals.

This makes it possible to for example use this tool as a sidecar to a container. As long as they share (parts of) the filesystem, this tool can supply the configuration to the program in the other container. If the two containers also share the process namespace, this tool can also be used for the signalling of the other process.

Another approach to use this tool is to create a custom image that would build on top of the original one and start both the original program and configbump.

## Current State

At the moment, only the configmap syncing is implemented. This makes the tool usable only with programs that can respond to changes in the filesystem on their own. We use it to dynamically supply configuration to Traefik configured to watch a directory for additional configuration files. Surprisingly, in our testing this seemed to be faster than using Traefik's own CRDs to achieve the same thing. It also doesn't need to install additional custom resources into the Kubernetes cluster.

## Future plans

When the process signalling is implemented, this tool will be compatible with many more programs to enable dynamic reconfiguration from the configmaps in the cluster.

We originally wrote a prototype of this tool in Rust (https://github.com/metlos/cm-bump) that implements both configmap syncing and process signalling and we successfully used it for dynamic reconfiguration of HAProxy, Nginx and Traefik.

## Configuration

```
$ ./configbump --help
config-bump 0.1.0
Usage: configbump --dir DIR --labels LABELS [--namespace NAMESPACE]

Options:
  --dir DIR, -d DIR      The directory to which persist the files retrieved from config maps. Can also be specified using env var: CONFIG_BUMP_DIR
  --labels LABELS, -l LABELS
                         An expression to match the labels against. Consult the Kubernetes documentation for the syntax required. Can also be specified using env var: CONFIG_BUMP_LABELS
  --namespace NAMESPACE, -n NAMESPACE
                         The namespace in which to look for the config maps to persist. Can also be specified using env var: CONFIG_BUMP_NAMESPACE. If not specified, it is autodetected.
  --help, -h             display this help and exit
  --version              display version and exit
```

## Examples

An example of using Traefik with configbump as a sidecar in a single pod to enable configbump dynamically downloading configuration files to a directory that Traefik watches for configuration changes.

```yaml
# The only thing that our Pod needs is to have access to the cluster API and be able to read
# config maps. The following service account, role and role binding show the minimum perms required:
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa-able-to-access-k8s-api-and-read-configmaps
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: read-configmaps
rules:
- verbs:
  - '*'
  apiGroups:
  - ""
  resources:
  - configmaps
  - pods
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-config-maps-to-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: read-configmaps
subjects:
- kind: ServiceAccount
  name: sa-able-to-access-k8s-api-and-read-configmaps
---
# This is the Pod with Traefik and configbump as a sidecar. The only things required to make
# configbump do its job is to a) assign the proper service account to the Pod and b) connect
# the Traefik container and configbump container using a shared emptydir volume. There is no
# need for the volume to be persistent because configbump syncs its content with all the matching
# configmaps.
kind: Pod
apiVersion: v1
metadata:
  name: traefik
spec:
  serviceAccountName: sa-able-to-access-k8s-api-and-read-configmaps
  containers:
  - name: traefik
    image: traefik
    args: ["--configFile=/dynamic-config/traefik.yml"]
    volumeMounts:
    - name: dynamic-config
      mountPath: "/dynamic-config"
  - name: config-map-sync
    image: quay.io/che-incubator/configbump:latest
    env:
    - name: CONFIG_BUMP_DIR
      value: "/dynamic-config"
    - name: CONFIG_BUMP_LABELS
      value: "config-for=traefik"
    - name: POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: CONFIG_BUMP_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    volumeMounts:
    - name: dynamic-config
      mountPath: "/dynamic-config"
  volumes:
  - name: config
    configMap:
      name: traefik-config
  - name: dynamic-config
    emptyDir: {}
---

# This is the main configuration for Traefik. We configure it to listen
# for changes in the "/dynamic-config" directory - where we put all the
# configuration from the config maps labeled with "config-for" label equal
# "traefik".
kind: ConfigMap
apiVersion: v1
metadata:
  name: traefik-config
  labels:
    config-for: traefik
data:
  traefik.yml: |
    global:
      checkNewVersion: false
      sendAnonymousUsage: false
    entrypoints:
      http:
        address: ":8080"
      https:
        address: ":8443"   
    providers:
      file:
        directory: "/dynamic-config"
        watch: true
```