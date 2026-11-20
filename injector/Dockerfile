# Stage 1: Build binary
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o manup-injector .

# Stage 2: Minimal runtime image
FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/manup-injector /app/manup-injector
USER 1000:1000
EXPOSE 8443
ENTRYPOINT ["/app/manup-injector"]
