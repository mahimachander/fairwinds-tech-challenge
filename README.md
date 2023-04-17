# Kubernetes Controller

This is a golang Kubernetes Controller which:

- Listens for existing and new pods
- Annotates those pods with a timestamp
- Logs the pod name and timestamp annotation to stdout
- Is deployed via a Helm chart
- Only respond to pods with a particular annotation (`managed`)
- Only respond to pods in namespaces with a particular annotation (`managed`)
- Implement leader election

## Versioning
Helm v3.10.3 and Go 1.19.4 are utilized.

## Deploying
From the parent directory, run `helm upgrade pod-controller ./k8s`. Resources will be in the `default` namespace.

## Further Work
The controller performs actions on existing and new pods; however, the desire is to only listen to new pods. More googling/testing is required to integrate this functionality. My hunch is that an option or query parameter of some sort exists so the controller could differentiate new pods based comparing the pod's created timestamp to the controller's own creation time.