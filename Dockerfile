# syntax=docker/dockerfile:1

FROM golang:1.19.3-alpine as builder
LABEL stage=builder

WORKDIR /app

# Copy necessary files
COPY go.mod ./
COPY go.sum ./

# Download necessary Go modules
RUN go mod download

# Copy files to workdir
COPY *.go ./
COPY static/index.html ./static/index.html

RUN go build

# Generate clean, final image for deployment
FROM alpine:3.16.2
LABEL stage=deploy

COPY --from=builder ./app/local-network-overview .
COPY --from=builder ./app/static/index.html ./static/index.html

# Executable
ENTRYPOINT [ "./local-network-overview" ]