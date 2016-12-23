# &#9179;

Simple tool to automate deployments to kubernetes cluster. Use-case is put this tool 
as build step of CI process. 
 
## Usage

```bash
fuse deployment.yml
```

## What will happen?

  * fuse will get all deployments defined in yml file
  * for each deployment defined in yml file fuse will fetch info from k8s cluster
  * command `kubectl apply -f deployment.yml` will be executed
  * fuse will check two things after `apply` for two minutes:
    * deployment generation is changed
    * deployment doesn't have any unavailable replicas
  * if both conditions are met, fuse assumes deployment is successful
  * if time limit is reached, and both conditions for each deployment is not met, command `rollout undo`
  will be executed for each deployment defined in yml file.
  
## Contexts

`kubectl` command support contexts, so, fuse trying to read environment variable
`CLUSTER_CONTEXT` and if it's not empty, argument `--context=${CLUSTER_CONTEXT}`
will be added to every `kubectl` command call, e.g:

```bash
export CLUSTER_CONTEXT=production
fuse deployment.yml
...
kubectl --context=production apply -f deployment.yml
kubectl --context=production rollout undo deployment/sample-deployment
```

## Stability

Tool currently in beta-testing stage. But, it using internally to deliver 
releases to our pre-production cluster.


## Sample output

```
[13:13:46][Step 4/5] Starting: /data/temp/agentTmp/custom_script164335181065595759
[13:13:46][Step 4/5] in directory: /data/work/fa28ed608c5de3ce
[13:13:46][Step 4/5] ==> Using file: /data/work/fa28ed608c5de3ce/kubernetes.yml
[13:13:46][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:13:46][Step 4/5] ==> New is not requested, updating current type
[13:13:46][Step 4/5] ==> Executing: kubectl apply -f /data/work/fa28ed608c5de3ce/kubernetes.yml
[13:13:47][Step 4/5] ==> Response from kubectl:
[13:13:47][Step 4/5] service "cdss-staging" configured
[13:13:47][Step 4/5] deployment "cdss-staging" configured
[13:13:47][Step 4/5]
[13:13:47][Step 4/5] ==> ZzzZzzZzz...
[13:13:52][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:13:52][Step 4/5] ==> Still unavailable: 1
[13:13:52][Step 4/5] ==> ZzzZzzZzz...
[13:13:57][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:13:57][Step 4/5] ==> Still unavailable: 1
[13:13:57][Step 4/5] ==> ZzzZzzZzz...
[13:14:02][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:02][Step 4/5] ==> Still unavailable: 1
[13:14:02][Step 4/5] ==> ZzzZzzZzz...
[13:14:07][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:07][Step 4/5] ==> Still unavailable: 1
[13:14:07][Step 4/5] ==> ZzzZzzZzz...
[13:14:12][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:12][Step 4/5] ==> Still unavailable: 1
[13:14:12][Step 4/5] ==> ZzzZzzZzz...
[13:14:17][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:17][Step 4/5] ==> Still unavailable: 1
[13:14:17][Step 4/5] ==> ZzzZzzZzz...
[13:14:22][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:22][Step 4/5] ==> Still unavailable: 1
[13:14:22][Step 4/5] ==> ZzzZzzZzz...
[13:14:27][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:27][Step 4/5] ==> Still unavailable: 1
[13:14:27][Step 4/5] ==> ZzzZzzZzz...
[13:14:32][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:32][Step 4/5] ==> Notice: no unavailable replicas found, assuming ok
[13:14:32][Step 4/5] ==> Success: All deployments marked as ok..
[13:14:32][Step 4/5] ==> Success: deploy successfull
[13:14:32][Step 4/5] Process exited with code 0
```

## License

Code is unlicensed. Do whatever you want with it. [Set Your Code Free](http://unlicense.org/).
