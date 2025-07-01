
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app/main.go

FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S seacatering && \
    adduser -u 1001 -S seacatering -G seacatering

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/database/migrations ./database/migrations
COPY --from=builder /app/database/seeds ./database/seeds

COPY --from=builder /app/tools/migration ./tools/migration

RUN mkdir -p /app/logs && chown -R seacatering:seacatering /app

USER seacatering

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENV GIN_MODE=release
ENV APP_ENV=production

CMD ["./main"]