
.PHONY: clean all

all: dev

dev: build

build: clean swag deps gobuild

deps:
	dep ensure

swag:
	cd ../ && swag init && cp -rv docs front/docs && cd -

gobuild: *.go **/*.go
	install -d -m0700 ./cache
	go build -o front

docker-deps:
	dep ensure --vendor-only

release: *.go **/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o front

test:
	go test ./...

run:
	echo "Start the api"
	mkdir -p ./cache
	GOOGLE_APPLICATION_CREDENTIALS=../keys/google_api_keys.json GIN_MODE=debug ./front

clean:
	rm -rf ./docs