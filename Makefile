test:
	golint -set_exit_status ./pkg/... ./bin/...
	ineffassign ./
	misspell -error README.md ./pkg/**/* ./bin/**/*
	gofmt -d -s -e ./bin/ ./pkg/ ./lib/
	go test -covermode=atomic ./pkg/...

format:
	gofmt -d -w -s -e ./bin/ ./pkg/ ./lib/

coverage:
	golint -set_exit_status ./pkg/... ./bin/...
	go list -f '"go test -covermode=atomic -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"' ./pkg/... | xargs -I % sh -c %
	gover ./ ./coverage.out

build:
	GOOS=linux GOARCH=amd64 go build -o ./fuse ./bin/main.go

.PHONY: test build lint
