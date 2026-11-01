FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY proto/ /proto/
COPY cmd/authentication/go.mod cmd/authentication/go.sum ./

RUN go mod download

COPY cmd/authentication/ ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /authentication \
    .

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /authentication ./authentication

USER app

EXPOSE 8082

ENTRYPOINT ["./authentication"]

