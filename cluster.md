# Condo Cluster
The **condo_cli** build system now has the ability to create a local Kubernetes cluster environment directly in docker.

### Requirements
The following packages need to be installed and globally accessible:
 - kind
 - git
 - kubectl
 - helm
 - docker

The **Condo Cluster** tool also require the **docker engine** to be running. 

---

## Create Cluster

**Creates a kubernetes cluster on your local docker instance.**
It also installs and auto-configures the following [Git-ops](https://www.weave.works/technologies/gitops/) tools:
1. Flux
2. Helm-Operator
3. git-server docker container (specific to flux)

```
condo_cli create cluster
```

The cluster configuration files are located at: `UserRoot/.am/clusters/{your-cluster-name}`


Flags:
```
---name {cluster-name} 
```
Used to specify the name for the cluster to be created, written in kebab-case. If not set, the name of the cluster defaults to "local"
```
---image {image-url} 
```
Used to specify the image to use for the cluster
```
---version {image-version} 
```
Used to specify the version of the image to use for the cluster

---

## Destroy Cluster
**Gracefully removes the cluster instance from docker, deleting the containers and images.** 

```
condo_cli destroy cluster
```

The configuration files for the cluster are presistent at `UserRoot/.am/clusters/{your-cluster-name}` and can be used again to rebuild the cluster. To do this, run the Create Cluster command with the same cluster name. 

To clear the configuration, delete the cluster folder  `UserRoot/.am/clusters/{your-cluster-name}`

Flags:
```
---name {cluster-name} 
```
Used to specify the name for the cluster to destroy, written in kebab-case. If not set, the name of the cluster defaults to "local"

---
# Deploy to your cluster
Your built cluster using the [Git-ops](https://www.weave.works/technologies/gitops/) model for building and managing deployments and services within your cluster. On Cluster creation, a local git repository was setup for this purpose at `UserRoot/.am/clusters/{your-cluster-name}/deploy/`.

The cluster also supports the use of helm and helmRelease manifests, the values for which can be configured at `UserRoot/.am/clusters/{your-cluster-name}/helm/`

### Adding Deployments and Services
1. Verify that your cluster is running in your docker instance
2. Add your deployment/service manifests (*.yaml files) into the deploy folder (`UserRoot/.am/clusters/{your-cluster-name}/deploy/`).
3. Within the context of the deploy folder, preform a git add and commit, eg.
```sh
git add -A
git commit -m "added new podinfo deployment"
```
4. Your changes will take effect in the cluster within a few minutes


### Modifying Deployments and Services
1. Verify that your cluster is running in your docker instance.
2. Edit your deployment/service manifests (*.yaml files) within the deploy folder  (`UserRoot/.am/clusters/{your-cluster-name}/deploy/`) with your preferred text editor.
3. Within the context of the deploy folder, preform a git add and commit, eg.
```sh
git add -A
git commit -m "modified podinfo deployment"
```
4. Your changes will take effect in the cluster within a few minutes

### Removing Deployments and Services
1. Verify that your cluster is running in your docker instance.
2. Remove your deployment/service manifests (*.yaml files) from the deploy folder  (`UserRoot/.am/clusters/{your-cluster-name}/deploy/`).
3. Within the context of the deploy folder, preform a git add and commit, eg.
```sh
git add -A
git commit -m "modified podinfo deployment"
```
4. Your changes will take effect in the cluster within a few minutes

---

# Deploy Docker Images
The cluster pulls docker images defined in your deployment/helmRelease from a local Image Registry Instance. An example helmRelease that pulls the image from the local Image Registry Instance is shown below, note the value of `repository` and `tag`.

```sh
apiVersion: helm.fluxcd.io/v1
kind: HelmRelease
metadata:
  name: demo-kebab
  namespace: stg
  annotations:
    fluxcd.io/automated: "true"
    fluxcd.io/tag.chart-image: semver:*
spec:
  chart:
    git: git@git.default:/git-server/repos/helm
    ref: local
    path: am-micro-service
  releaseName: demo-kebab
  helmVersion: v3
  values:
    environment: local
    image:
      repository: localhost:5000/kebab-demo
      tag: 1.20.59
    privileged: true
    repository:
      name: kebab
    ingress:
      external:
        enabled: true

```
### Viewing the local Image Registry Instance catalog

1. Verify that the `docker-image-reg` docker container is running
2. access [http://localhost:5000/v2/_catalog](http://localhost:5000/v2/_catalog)


### Pushing Images to the local Image Registry Instance
1. Verify that the `docker-image-reg` docker container is running
2. Verify docker can access your image. Use command:
```sh
docker images
```
3. Tag the image for the Image Registry Instance. use the command:
```sh
docker tag <imageID> localhost:5000/<imageName>:<taggedVersion>
```
4. Push the image to the Image Registry Instance. use the command:
```sh
docker push localhost:5000/<imageName>:<taggedVersion>
```
