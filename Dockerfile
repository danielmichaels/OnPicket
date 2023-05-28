FROM golang:1.20-bullseye AS builder

# Move to working directory (/build).
WORKDIR /build

# only copy mod file for better caching
COPY go.mod go.mod
RUN go mod download

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
# copy files
COPY . .
RUN go build  \
    -ldflags="-s -w" \
    -o app github.com/danielmichaels/onpicket/cmd/app

FROM gcr.io/distroless/base-debian11:debug

COPY --from=builder ["/build/app", "/usr/bin/app"]

## Command to run when starting the container.
ENTRYPOINT ["app"]
