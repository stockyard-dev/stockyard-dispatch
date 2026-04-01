FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" -o /bin/dispatch ./cmd/dispatch/
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=builder /bin/dispatch /usr/local/bin/dispatch
ENV PORT="8900" DATA_DIR="/data" RETENTION_DAYS="30" DISPATCH_LICENSE_KEY="" SMTP_HOST="" SMTP_PORT="587" SMTP_USER="" SMTP_PASS="" SMTP_FROM=""
EXPOSE 8900
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD curl -sf http://localhost:8900/health || exit 1
ENTRYPOINT ["dispatch"]
