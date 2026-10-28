FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY . .

RUN go work sync

RUN CGO_ENABLED=0 go build -o /app ./cmd/account

FROM alpine:latest AS runner

RUN apk --no-cache add wget

WORKDIR /app

COPY --from=builder /app .

CMD ["./app"]
