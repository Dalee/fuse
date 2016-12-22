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

## License

Code is unlicensed. Do whatever you want with it. [Set Your Code Free](http://unlicense.org/).
