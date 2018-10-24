
.PHONY: clean all

all: dev

dev: build

build: clean swag deps gobuild

deps:
	GO111MODULE=on go get -d -v
	GO111MODULE=on go mod tidy

swag:
	GO111MODULE=on go get -v github.com/swaggo/swag/cmd/swag
	cd ../ && GO111MODULE=on swag init && cp -rv docs front/docs && cd -

gobuild: *.go **/*.go
	install -d -m0700 ./cache
	GO111MODULE=on go build -o front

docker:
	sudo docker build ./dockerfiles/front/Dockerfile

release: *.go **/*.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -a -installsuffix cgo -ldflags="-w -s" -o front

test:
	GO111MODULE=on go test ./...

run:
	echo "Start the api"
	mkdir -p ./cache
	GOOGLE_APPLICATION_CREDENTIALS=../keys/google_api_keys.json GIN_MODE=debug ./front

clean:
	rm -rf ./docs