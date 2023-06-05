FROM node:alpine as node
RUN mkdir -p /build
WORKDIR /build
COPY . .
RUN yarn &&\
    mkdir -p /build/assets/static/css/ &&\
    mkdir -p /build/assets/static/js/ &&\
    yarn tailwind &&\
    yarn alpine

FROM golang:1.20-alpine AS builder

WORKDIR /build

# only copy mod file for better caching
COPY go.mod go.mod
RUN go mod download

COPY --from=node ["/build/assets/static/css/theme.css", "/build/assets/static/css/theme.css"]
COPY --from=node ["/build/assets/static/js/bundle.js", "/build/assets/static/js/bundle.js"]

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY . .
RUN apk add git &&\
    go build  \
    -ldflags="-s -w" \
    -o app github.com/danielmichaels/onpicket/cmd/app

FROM instrumentisto/nmap

COPY --from=builder ["/build/app", "/usr/bin/app"]

ENTRYPOINT ["app"]
