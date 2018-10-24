
.PHONY: all
all: dev

dev: build

deps:
	dep ensure

build: deps gobuild

gobuild: *.go
	go build

release: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o sms_worker

run:
	echo "Starting the SMS worker"
	./sms_worker
