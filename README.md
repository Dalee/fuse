# &#9179;

[![Build Status](https://travis-ci.org/Dalee/fuse.svg?branch=master)](https://travis-ci.org/Dalee/fuse)
[![Coverage](https://codecov.io/gh/Dalee/fuse/branch/master/graph/badge.svg)](https://codecov.io/gh/Dalee/fuse)


Simple tool build around `kubectl` command, great for CI/CD environments.

Key features:
 * Deploy new release to cluster with automated rollback in case of error
 * Automatic image garbage collection for private Docker registry

## Deploy

Apply new configuration to Kubernetes cluster and monitor release delivery.

Usage:
```
$ fuse apply -f deployment.yml
```

Help screen:
```
$ fuse help apply
Apply new configuration to Kubernetes cluster and monitor release delivery

Usage:
  fuse apply [flags]

Flags:
  -f, --configuration string   Release configuration yaml file
```

### What `apply` command do?

  * `fuse` will get all deployments defined in yml file
  * for each deployment defined in yml file fuse will fetch info from k8s cluster
  * command `kubectl apply -f deployment.yml` will be executed
  * `fuse` will periodically check two things after `kubectl apply` for a two minutes:
    * deployment generation is changed
    * deployment doesn't have any unavailable replicas
  * if both conditions are met, fuse assumes deployment is successful
  * if time limit is reached, and both conditions for each deployment is not met, command `rollout undo`
  will be executed for each deployment defined in yml file.
  
### Sample output

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
...
[13:14:27][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:27][Step 4/5] ==> Still unavailable: 1
[13:14:27][Step 4/5] ==> ZzzZzzZzz...
[13:14:32][Step 4/5] ==> Executing: kubectl get deployment/cdss-staging -o yaml
[13:14:32][Step 4/5] ==> Notice: no unavailable replicas found, assuming ok
[13:14:32][Step 4/5] ==> Success: All deployments marked as ok..
[13:14:32][Step 4/5] ==> Success: deploy successfull
[13:14:32][Step 4/5] Process exited with code 0
```

## Garbage Collect

Remove tags from registry not registered within any Kubernetes ReplicaSet

Usage:
```
$ fuse garbage-collect --registry-url=https://registry.example.com:5000/
```

Help screen:
```
$ fuse help garbage-collect
Remove tags from registry not registered within any Kubernetes ReplicaSet

Usage:
  fuse garbage-collect [flags]

Flags:
      --dry-run               Do not try to execute destructive actions (default "false")
      --ignore-missing        Ignore/Skip missing images in Registry (default "false")
      --namespace string      Namespace to fetch ReplicaSet (default "default")
      --registry-url string   Registry URL to use (e.g. https://example.com:5000/)
```

### What `garbage-collect` command do?

  * `fuse` will search all replica sets for given namespace (`default` is by default)
  * For each replica set `Spec.Template.Spec.Containers[].Image` will be analyzed
  * For each image repository, full list of tags and image digests will be fetched from provided `registry-url`
  * If some of repositories absent in provided `registry-url`, error will be thrown, unless `ignore-missing` is set
  * For each founded image repository, all tags not registered in replica set (i.e. not deployed or stale)
    will be marked for deletion
  * If `dry-run` is not set, images digests, marked for deletion, will be marked for deletion in Registry 
  (Beware: Registry itself has own `garbage-collection` command)

> Do not forget to schedule [Registry garbage-collect](https://docs.docker.com/registry/garbage-collection/) command
to perform actual cleanup of deleted images!

### Sample output

```
==> Using namespace: default
==> Executing: kubectl --context=production --namespace=default get replicasets -o yaml
==> Found: 11 ReplicaSets
==> Detecting garbage, dry-run is: true
==> Detection, done
==> acme/project1-live
Deployed: [3 7 4 6 8 5]
Garbage: [1 2]
sha256:ee09ac314a1a79a202b8646538cc9298a8f87da27fb69359f6e7a3f1e7c48e5b
==> heapster
Deployed: [v1.2.0]
Garbage: []
==> kubernetes-dashboard-amd64
Deployed: [1.5.0]
Garbage: []
==> acme/project2-live
Deployed: [4]
Garbage: [1 2 3]
sha256:50989e7a5c59bef81b5f28297f834014fb0903de105cdbff6a57ef263861ecef
sha256:74ba39cfc2ebce16582c1dd9254fa0a566fb05e7f34f149c90382a0085697f08
Done, have a nice day!
```

## Contexts

`kubectl` command support contexts, so, fuse trying to read environment variable
`CLUSTER_CONTEXT` and if it's not empty, argument `--context=${CLUSTER_CONTEXT}`
will be added to every `kubectl` command call, e.g:

Command `apply`:
```
export CLUSTER_CONTEXT=production
fuse apply -f deployment.yml
...
kubectl --context=production apply -f deployment.yml
...
kubectl --context=production rollout undo deployment/sample-deployment
```

Command `garbage-collect`:
```
export CLUSTER_CONTEXT=production
fuse garbage-collect --dry-run --registry-url=https://registry.example.com:5000/
==> Using namespace: default
==> Executing: kubectl --context=production --namespace=default get replicasets -o yaml
...
```

## Stability

Tool currently in pre-release stage. But, it is using heavily to deliver 
releases to our staging/production cluster. So, at least `apply` command 
is mature enough.

Tool is tested with Kubernetes `v1.2.0`

## Known issues / How to avoid problems

Put `build id`, `build number` or any auto incremented value, provided by CI/CD,
as `env` variable parameter or as deployment `label`. This is workaround 
for a situation when Docker image create from `cache` is re-deployed as 
new build (even with new image tag). In this situation Kubernetes will not 
apply any changes to  `deployment`  configuration and will not update 
`Status.ObservedGeneration`, in this case `fuse` will rollout release.

## License

Fuse is licensed under the Apache License, Version 2.0. 
See LICENSE for the full license text.

[fuse v1.0.1](https://github.com/Dalee/fuse/tree/v1.0.1) is released under 
[Unlicense](http://unlicense.org/) license terms, so you can use it, 
if you want.


## Development

 * Golang >= 1.7.x
 * [golint](https://github.com/golang/lint)
 * [glide](https://github.com/Masterminds/glide)
 * [gover](https://github.com/modocache/gover)
 * make


Install developer dependencies:
```
$ go get -u github.com/modocache/gover && \
go get -u github.com/golang/lint/golint && \
go get -u github.com/Masterminds/glide
```

Install project dependencies:
```
$ glide install
```

Test and Coverage
 * `make test` — linting and testing
 * `make coverage && go tool cover -html=coverage.out` — see coverage
