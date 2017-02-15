test:
	golint -set_exit_status ./pkg/... ./bin/...
	go test -covermode=atomic ./pkg/...

coverage:
	golint -set_exit_status ./pkg/... ./bin/...
	go list -f '"go test -v -covermode=atomic -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"' ./pkg/... | xargs -I % sh -c %
	gover ./ ./coverage.out

build:
	GOOS=linux GOARCH=amd64 go build -o ./fuse ./bin/main.go

.PHONY: test build lint
