# &#9179; — Kubernetes deploy tool

[![Build Status](https://travis-ci.org/Dalee/fuse.svg?branch=master)](https://travis-ci.org/Dalee/fuse)
[![Coverage](https://codecov.io/gh/Dalee/fuse/branch/master/graph/badge.svg)](https://codecov.io/gh/Dalee/fuse)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dalee/fuse)](https://goreportcard.com/report/github.com/Dalee/fuse)

Simple, but powerful tool, build around `kubectl` command, great for CI/CD environments.

Key features:
 * `apply` — update cluster configuration with automated undo in case of error
 * `garbage-collect` — detect unused images for each deployment, remove unused images from 
 Docker Distribution (Registry). 
 
## Configuration

Environment variables:
 * `CLUSTER_CONTEXT` cluster context to use (default, no context)
 * `CLUSTER_RELEASE_TIMEOUT` timeout before release will be marked as failed (default, 120 seconds)

Flags:
 * Cluster context variable can be overridden via global flag: `-c` or `--context`
 * Cluster release timeout can be overridden via `-t` or `--release-timeout` for `apply` command
 * For a `garbage-collect` command, cluster namespace can be changed via `-n, --namespace`, default is `"default"`.

## Apply

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
  -f, --configuration string   Release configuration yaml file, mandatory
  -t, --release-timeout int    Deploy timeout in seconds, override CLUSTER_RELEASE_TIMEOUT environment

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

## Garbage Collect

Remove tags from registry not registered within any Kubernetes ReplicaSet

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
  -n, --namespace string      Kubernetes namespace to use (default "default")
  -r, --registry-url string   Registry URL (e.g. "https://registry.example.com:5000/")

Global Flags:
  -c, --context string   Override CLUSTER_CONTEXT defined in environment (default "")
```

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

Test and Coverage
 * `make test` — linting and testing
 * `make coverage` — display coverage information
 * `make format` — gofmt sources
 * `make coverage && go tool cover -html=coverage.txt` — see coverage
