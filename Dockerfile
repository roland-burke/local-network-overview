FROM golang:1.20-alpine as builder
LABEL stage=builder

WORKDIR /app

# Copy necessary files
COPY go.mod ./
COPY go.sum ./

# Download necessary Go modules
RUN go mod download

# Copy files to workdir
COPY cmd/ ./cmd
COPY internal/ ./internal
COPY static/index.html ./static/index.html

RUN go build -o ./network-overview ./cmd/main

# Generate clean, final image for deployment
FROM alpine:3
LABEL stage=deploy

COPY --from=builder ./app/network-overview .
COPY --from=builder ./app/static/index.html ./static/index.html

# Executable
ENTRYPOINT [ "./network-overview" ]