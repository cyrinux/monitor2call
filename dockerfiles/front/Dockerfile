# STEP 1 build executable binary
FROM golang:alpine as builder
RUN apk update
RUN apk add git && apk add ca-certificates && apk add make && apk add gcc && apk add musl-dev
# Create user monitor2call
RUN adduser -D -g '' monitor2call
# Copy source
COPY . $GOPATH/src/monitor2call/
# Make front
WORKDIR $GOPATH/src/monitor2call/front/
# Create swagger doc
RUN make clean
RUN make swag
# Build the front binary
RUN make release
# Move front files
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