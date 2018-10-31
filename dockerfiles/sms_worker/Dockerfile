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
WORKDIR $GOPATH/src/github.com/cyrinux/monitor2call/sms_worker
# Build the worker binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o sms_worker
# Create user monitor2call
RUN adduser -D -g '' monitor2call
# Move worker file
RUN mv sms_worker /go/bin/monitor2call_sms_worker
# Show version
RUN /go/bin/monitor2call_sms_worker --version

# STEP 3 build the smallest worker image
# start from scratch
FROM scratch as worker
# Install ca-certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Install user
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY --from=builder /go/bin/monitor2call_sms_worker /go/bin/
USER monitor2call
ENTRYPOINT ["/go/bin/monitor2call_sms_worker"]
