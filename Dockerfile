FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server-inventory .

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server-inventory /app/server-inventory

USER appuser

ENV PORT=8080
EXPOSE 8080

CMD ["/app/server-inventory"]
