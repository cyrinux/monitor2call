
.PHONY: all
all: dev

dev: build

deps:
	GO111MODULE=on go get -d -v
	GO111MODULE=on go mod tidy

build: deps gobuild

gobuild: *.go
	GO111MODULE=on go build

docker:
	sudo docker build ./dockerfiles/worker/Dockerfile

release: *.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -a -installsuffix cgo -ldflags="-w -s" -o worker

run:
	echo "Starting the worker"
	./worker
