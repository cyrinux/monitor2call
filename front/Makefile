
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

test:
	go test ./...

run:
	echo "Start the api"
	mkdir -p ./cache
	GOOGLE_APPLICATION_CREDENTIALS=../keys/google_api_keys.json GIN_MODE=debug ./front

clean:
	rm -rf ./docs