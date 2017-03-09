install:
	go get -u github.com/modocache/gover
	go get -u github.com/golang/lint/golint
	go get -u github.com/Masterminds/glide
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/client9/misspell/cmd/misspell
	glide install

test:
	golint -set_exit_status ./pkg/... ./bin/...
	ineffassign ./
	misspell -error README.md ./pkg/**/* ./bin/**/*
	gofmt -d -s -e ./bin/ ./pkg/
	go test -covermode=atomic ./pkg/...

format:
	gofmt -d -w -s -e ./bin/ ./pkg/

coverage:
	go list -f '"go test -covermode=atomic -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"' ./pkg/... | xargs -I % sh -c %
	gover ./ ./coverage.txt

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./fuse ./bin/main.go

.PHONY: test build lint
