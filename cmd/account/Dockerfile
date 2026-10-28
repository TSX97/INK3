FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY proto/ ./proto/
COPY db/ ./db/
COPY cmd/account/ ./cmd/account/

WORKDIR /src/cmd/account

RUN go mod download

RUN CGO_ENABLED=0 go build -o /app .

FROM alpine:latest AS runner

RUN apk --no-cache add wget

WORKDIR /app

COPY --from=builder /app .

CMD ["./app"]

