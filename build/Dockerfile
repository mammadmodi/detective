FROM golang:alpine AS builder

WORKDIR /app

# For using layer caching in docker, we first copy the go.mod and go.sum files.
COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/server

RUN go build -o /detective-server

FROM alpine:latest

WORKDIR /app/

# Copy web assets and files to image.
COPY --from=builder /app/web ./web

# Copy app binary file.
COPY --from=builder /detective-server .

EXPOSE 8000

RUN addgroup -g 1001 appuser && \
    adduser -S -u 1001 -G appuser appuser

RUN chown -R appuser:appuser /var/log/

USER appuser

ENTRYPOINT ["./detective-server"]
