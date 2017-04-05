# &#9179; — Kubernetes deploy tool

[![Build Status](https://travis-ci.org/Dalee/fuse.svg?branch=master)](https://travis-ci.org/Dalee/fuse)
[![Coverage](https://codecov.io/gh/Dalee/fuse/branch/master/graph/badge.svg)](https://codecov.io/gh/Dalee/fuse)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dalee/fuse)](https://goreportcard.com/report/github.com/Dalee/fuse)
[![codebeat badge](https://codebeat.co/badges/acdf46b7-8265-42b6-96a4-c770969250ef)](https://codebeat.co/projects/github-com-dalee-fuse-master)

Simple, but powerful tool, build around `kubectl` command, great for CI/CD environments.

Key features:
 * `apply` — update cluster configuration with automated undo in case of error
 * `garbage-collect` — detect unused images for each deployment, remove unused images from 
 Docker Distribution (Registry). 
 
## Configuration

Environment variables:
 * `CLUSTER_CONTEXT` cluster context to use (default, no context, so `kubectl` will use default)

Flags:
 * Cluster context variable can be overridden via global flag: `-c` or `--context`
 * Cluster rollout timeout can be set via `-t` or `--release-timeout` for `apply` command
 * For a `garbage-collect` command, cluster namespace can be changed via `-n, --namespace`, default is `"default"`.

## Kubernetes Rollout

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
  -f, --configuration string       Rollout configuration spec file (yaml), mandatory
  -t, --rollout-timeout duration   Rollout timeout (default 2m0s)

Global Flags:
  -c, --context string   Override CLUSTER_CONTEXT defined in environment (default "")
```

### What `apply` command do?

  * `fuse` will get all deployments defined in configuration yml file
  * command `kubectl apply -f deployment.yml` will be executed
  * for each deployment, [deployment rollout status](https://kubernetes.io/docs/user-guide/deployments/#the-status-of-a-deployment) will be monitored 
  * if status is successful:
    * fuse will display logs for each created pod for each deployment
  * if timeout reached: 
    * fuse will display logs from pods attached to each deployment
    * for each deployment `rollout undo` will be executed, but only if deployment undo history is present
    
### Sample output

```
[14:31:04][Step 4/5] Starting: /data/tc-agent/temp/agentTmp/custom_script427977640370952504
[14:31:04][Step 4/5] in directory: /data/tc-agent/work/fa28ed608c5de3ce
[14:31:04][Step 4/5] ===> kubectl apply -f kubernetes.yml -o name
[14:31:04][Step 4/5] service/example-staging
[14:31:04][Step 4/5] deployment/example-staging
[14:31:04][Step 4/5] 
[14:31:04][Step 4/5] ==> Starting rollout monitoring, timeout is 2m0s seconds
[14:31:09][Step 4/5] ===> kubectl --namespace=default get deployment/example-staging -o yaml
[14:31:09][Step 4/5] ===> Deployment: default/example-staging, Ready: false, Generation: meta=84 observed=84, Replicas: s=1, u=1, a=1, na=1
[14:31:14][Step 4/5] ===> kubectl --namespace=default get deployment/example-staging -o yaml
[14:31:15][Step 4/5] ===> Deployment: default/example-staging, Ready: true, Generation: meta=84 observed=84, Replicas: s=1, u=1, a=1, na=0
[14:31:15][Step 4/5] ==> Rollout done!
[14:31:15][Step 4/5] ==> Fetching logs...
[14:31:15][Step 4/5] ===> kubectl --namespace=default get deployment/example-staging -o yaml
[14:31:15][Step 4/5] ===> kubectl --namespace=default get pods --selector=app=example-staging -o yaml
[14:31:15][Step 4/5] ===> kubectl --namespace=default logs --tail=100 example-staging-1567981747-4awvj
[14:31:15][Step 4/5] ===> Deployment: default/example-staging, Pod: default/example-staging-1567981747-4awvj:
[14:31:15][Step 4/5] *** Running /etc/my_init.d/00_regen_ssh_host_keys.sh...
[14:31:15][Step 4/5] *** Running /etc/rc.local...
[14:31:15][Step 4/5] *** Booting runit daemon...
[14:31:15][Step 4/5] *** Runit started as PID 14
[14:31:15][Step 4/5] 2017/03/09 14:31:05 [notice] 23#23: using the "epoll" event method
[14:31:15][Step 4/5] 2017/03/09 14:31:05 [notice] 23#23: nginx/1.11.10
[14:31:15][Step 4/5] 2017/03/09 14:31:05 [notice] 23#23: built by gcc 5.4.0 20160609 (Ubuntu 5.4.0-6ubuntu1~16.04.2) 
[14:31:15][Step 4/5] 2017/03/09 14:31:05 [notice] 23#23: OS: Linux 3.10.0-327.22.2.el7.x86_64
[14:31:15][Step 4/5] 2017/03/09 14:31:05 [notice] 23#23: getrlimit(RLIMIT_NOFILE): 1048576:1048576
[14:31:15][Step 4/5] 2017/03/09 14:31:05 [notice] 23#23: start worker processes
[14:31:15][Step 4/5] 2017/03/09 14:31:05 [notice] 23#23: start worker process 35
[14:31:15][Step 4/5] [Mar 9 14:31:06.712] info: server started: http://example-staging-1567981747-4awvj:3000
[14:31:15][Step 4/5] [Mar 9 14:31:06.722] info: Redis connected
[14:31:15][Step 4/5] [Mar 9 14:31:06.733] info: Redis ready
[14:31:15][Step 4/5] [Mar 9 14:31:10.197] debug: HEAD http://127.0.0.1:3000/healthcheck
[14:31:15][Step 4/5] [Mar 9 14:31:10.225] debug: HEAD http://127.0.0.1:3000/healthcheck
[14:31:15][Step 4/5] [Mar 9 14:31:15.206] debug: HEAD http://127.0.0.1:3000/healthcheck
[14:31:15][Step 4/5] [Mar 9 14:31:15.219] debug: HEAD http://127.0.0.1:3000/healthcheck
[14:31:15][Step 4/5] 
[14:31:15][Step 4/5] Process exited with code 0
```

## Registry Garbage Collection

> Docker Registry access and manipulation is based on our another project [Hitman](https://github.com/Dalee/hitman).

Remove tags from registry not registered within any Kubernetes ReplicaSet.

Usage:
```
$ fuse garbage-collect --registry-url=https://registry.example.com:5000/
```

Help screen:
```
$ fuse help garbage-collect
Remove tags from registry not registered within Kubernetes ReplicaSet

Usage:
  fuse garbage-collect [flags]

Flags:
  -d, --dry-run               Do not execute destructive actions (default "false")
  -i, --ignore-missing        Skip missing images in Registry (default "false")
  -k, --keep-tag stringSlice  Keep tag in Registry, even if it not deployed (default none)
  -n, --namespace string      Kubernetes namespace to use (default "default")
  -r, --registry-url string   Registry URL (e.g. "https://registry.example.com:5000/")

Global Flags:
  -c, --context string   Override CLUSTER_CONTEXT defined in environment (default "")
```

> `-k/--keep-tag` can be provided multiple times, best use case is keep `latest` tag
in order to speed up build image time.


### What `garbage-collect` command do?

  * `fuse` will search all replica sets for given namespace (`default` is by default)
  * For each replica set `Spec.Template.Spec.Containers[].Image` will be analyzed
  * For each image repository, full list of tags and image digests will be fetched from provided `registry-url`
  * If some of repositories absent, error will be thrown, unless `ignore-missing` flag is set
  * All tags of image not registered within any `ReplicaSet` will be marked for deletion
  * If `dry-run` is not set, images digests, marked for deletion, will be marked for deletion 
  in Docker Distribution (beware: Registry itself has own `garbage-collect` command)

> Do not forget to schedule [Registry garbage-collect](https://docs.docker.com/registry/garbage-collection/) command
to perform actual cleanup of deleted images!

### Sample output

```
[14:37:38][Step 2/5] Starting: /data/tc-agent/temp/agentTmp/custom_script5182791551931646628
[14:37:38][Step 2/5] in directory: /data/tc-agent/work/2d12da58bb93314d
[14:37:38][Step 2/5] ==> Fetching repository info...
[14:37:38][Step 2/5] ===> kubectl --namespace=default get replicasets -o yaml
[14:37:41][Step 2/5] ==> Found 1 repositories
[14:37:41][Step 2/5] ===> Repository: acme/example-staging
[14:37:41][Step 2/5] ===> Deployed: [45 40 41 42 43 44]
[14:37:41][Step 2/5] ===> Detected as garbage: [34 35 36 37 38 39]
[14:37:41][Step 2/5]
[14:37:41][Step 2/5] ==> Clearing up...
[14:37:41][Step 2/5] ===> Done: acme/example-staging
[14:37:42][Step 2/5] Process exited with code 0
```

## Stability

Tool currently in pre-release stage, but, it is using heavily to deliver 
releases to our staging/production cluster. Same applied to `garbage-collect`, 
clean up images is used both at production and staging cluster. 

All interactions with kubectl is covered by tests, but scenarios (commands) are not.
Use at your own risk.

Tool is tested with Kubernetes/kubectl `v1.2.0`

## License

Fuse is licensed under the Apache License, Version 2.0. 
See LICENSE for the full license text.

> [fuse v1.0.1](https://github.com/Dalee/fuse/tree/v1.0.1) is released under 
[Unlicense](http://unlicense.org/)

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
go get -u github.com/Masterminds/glide && \
go get -u github.com/gordonklaus/ineffassign && \
go get -u github.com/client9/misspell/cmd/misspell
```

Install project dependencies:
```
$ glide install
```

Test and Coverage:
 * `make test` — linting and testing
 * `make coverage` — display coverage information
 * `make format` — gofmt sources
 * `make coverage && go tool cover -html=coverage.txt` — see coverage


## Useful links, Further reading

* [Kubernetes official site](https://kubernetes.io/)
* [Official way to deploy applications (helm)](https://github.com/kubernetes/helm)
