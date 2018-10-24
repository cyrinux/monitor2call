# STEP 1 build executable binary
FROM golang:alpine as builder
RUN apk update
RUN apk add git && apk add ca-certificates && apk add make && apk add gcc && apk add musl-dev
# Create user monitor2call
RUN adduser -D -g '' monitor2call
COPY . $GOPATH/src/monitor2call/
WORKDIR $GOPATH/src/monitor2call/worker/
# Get dependancies
RUN make deps
# Build the worker binary
RUN make release
# Move worker file
RUN mv worker /go/bin/monitor2call_worker

# STEP 3 build the smallest worker image
# start from scratch
FROM scratch as worker
# Install ca-certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Install user
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY --from=builder /go/bin/monitor2call_worker /go/bin/
USER monitor2call
ENTRYPOINT ["/go/bin/monitor2call_worker"]
