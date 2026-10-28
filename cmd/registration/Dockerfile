FROM golang:latest AS builder

WORKDIR /src

COPY proto/ ./proto/

COPY cmd/registration/ ./cmd/registration/

WORKDIR /src/cmd/registration

RUN go mod download

RUN CGO_ENABLED=0 go build -o /app .

FROM scratch AS runner

WORKDIR /app

COPY --from=builder /app .

CMD ["./app"]
