# Condo Cluster Deployment Troubleshooting
While Kubernetes combined with the [Git-ops](https://www.weave.works/technologies/gitops/) approach provides a multitude of advantages in terms of scalability and maintainability, can be complicated to set up and diagnose. Below are common commands used to gather more information about your deployments.


### [Command Set 1] Pods within your cluster
A [pod](https://kubernetes.io/docs/concepts/workloads/pods/) is an instance of your running application that has been deployed (although it may still be experiencing issues) 
```sh
kubectl get pods {pod-name} -n {namespace} #Get pod within a particular namespace with a particular name
kubectl get pods -n {namespace} #Get pods within a particular namespace
kubectl get pods --all-namespaces #Get all pods within the entire cluster

kubectl describe {pod-name} -n {namespace} #Get more details of a particular pod

kubectl logs {pod-name} #Get log dump of a particular pod
```

### [Command Set 2] Deployments within your cluster
A [deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) defines your desired state (configuration) of your application in the cluster.
```sh
kubectl get deployments {deployment-name} -n {namespace} #Get pod within a particular namespace with a particular name
kubectl get deployments -n {namespace} #Get deployments within a particular namespace
kubectl get deployments --all-namespaces #Get all deployments within the entire cluster

kubectl describe {deployment-name} -n {namespace} #Get more details of a particular deployment
```

## Common Issues
### Where is my pod?
After performing a git add and commit to your deployment yaml file within the deploy folder (`UserRoot/.am/clusters/{your-cluster-name}/deploy/`), it usually takes a few minutes for flux to detect and role out the change. The easiest way to check the status of your pod(s) is by finding the pod by the namespace it was supposed to be in (see command 1 above).

---

### I see my pod(s) but they are not ready/show an error
Pods do take time to build, however, they can fail to start or crash during runtime due to a number of reasons. To get more details as to the state of the pod(s) use the `describe` command (see command 1 above). For more in-depth information, consider getting the logs for that pod (see command 1 above).

---

### I have waited 10 minutes, I still do not see my pod
Check to see if flux created a deployment for your application (see command 1 above).

If the deployment exists but pods are not being created, try scaling the pods manually with the following command:
```sh
kubectl scale deployment {deployment-name} -n {namespace} --replicas=2
```
A deployment may have been created but pods were not automatically scaled due to default helm configurations.

---

### I do not see the deployment either
If the deployment does not exist, investigate the flux logs to determine why flux did not create the deployment by using the following commands:

```sh
kubectl get pods -n weave #List all of the pods in the weave namespace

#Identify the flux pod, usually a pod like: flux-7785dfc54d-cx9w2

kubectl logs {flux pod} -n weave #Get logs from flux

#(OPTIONAL)
kubectl logs flux-7785dfc54d-cx9w2 -n weave > flux-logs.txt #pipe flux logs to a text file
```

