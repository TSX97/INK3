FROM golang:latest AS builder

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 go build -o app ./cmd/registration

FROM scratch AS runner

WORKDIR /app

COPY --from=builder /builder/app .

CMD ["./app"]
