FROM golang:alpine AS builder

WORKDIR /app

# For using layer caching of docker, we first copy the go.mod and go.sum files.
COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/server

RUN go build -o /server

FROM alpine:latest

WORKDIR /app/

COPY --from=builder /server .

EXPOSE 8000

ENTRYPOINT ["./server"]
