FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main -ldflags="-s -w" .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .

COPY .env.example .env

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 7007

CMD ["./main"]
