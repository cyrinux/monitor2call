# STEP 1 build executable binary
FROM golang:alpine as builder
RUN apk update
RUN apk add git && apk add ca-certificates && apk add make && apk add gcc && apk add musl-dev
# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep
WORKDIR $GOPATH/src/github.com/cyrinux/monitor2call/
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
WORKDIR $GOPATH/src/github.com/cyrinux/monitor2call/front/
# Create swagger doc
RUN go get github.com/swaggo/swag/cmd/swag
RUN rm -rf ./docs
RUN cd ../ && swag init && cp -rv docs front/docs && cd -
# Build the front binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o front
# Create user monitor2call
RUN adduser -D -g '' monitor2call
# Move file
RUN mv front /go/bin/monitor2call_front ; \
    install -d -m0755 -o monitor2call /go/cache;  \
    install -d -m0755 -o monitor2call /go/keys; \
    mv assets /go/assets


# STEP 2 build the smallest front image
# start from scratch
FROM scratch as front
# Install ca-certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Install user
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY --from=builder /go/bin/monitor2call_front /go/bin/
# Create directories
COPY --from=builder /go/cache /go/cache
COPY --from=builder /go/assets /go/assets
COPY --from=builder /go/keys /go/keys
USER monitor2call
EXPOSE 3000
ENTRYPOINT ["/go/bin/monitor2call_front"]