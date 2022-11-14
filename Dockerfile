# BASE IMAGE
FROM golang:1.19-alpine AS base

ENV GO111MODULE="on"
ENV GOOS="linux"
ENV CGO_ENABLED=1

RUN apk update \
    && apk add --no-cache \
    build-base \
    ca-certificates \
    curl \
    tzdata \
    git \
    vips \
    vips-dev \
    && update-ca-certificates


# DEVELOPMENT IMAGE
FROM base AS dev

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest
    # \ && go install github.com/go-delve/delve/cmd/dlv@latest

EXPOSE 4000
EXPOSE 2345

ENTRYPOINT ["air"]


# BUILDER IMAGE
FROM base AS builder

WORKDIR /app

COPY . .

RUN go mod download && go mod verify
RUN go build -o imgress


# PRODUCTION IMAGE
FROM alpine:latest as prod

COPY --from=builder /app/imgress /usr/local/bin/imgress

EXPOSE 5000

ENTRYPOINT ["/usr/local/bin/imgress"]
