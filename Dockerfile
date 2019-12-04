FROM golang:alpine as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
# gcc and musl-dev required by go-sqlite3
RUN apk update && apk add --no-cache git ca-certificates tzdata gcc musl-dev && update-ca-certificates

WORKDIR /app

# Fetch dependencies:
# https://medium.com/@petomalina/using-go-mod-download-to-speed-up-golang-docker-builds-707591336888
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ENV GOOS=linux

# Multi-stage build based off:
# https://github.com/chemidy/smallest-secured-golang-docker-image

# Run tests
RUN go test -cover ./...

# Build the binary
RUN go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/app .

# Create app user
RUN adduser -D -g '' app

FROM alpine

# iap_curl calls out to curl after establishing the token
RUN apk update && apk add --no-cache curl

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /go/bin/app /bin/iap_curl

# Use an unprivileged user.
USER app

ENTRYPOINT ["/bin/iap_curl"]
