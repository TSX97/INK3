FROM golang:latest AS builder

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 go build main.go -o app

FROM scratch AS runner

WORKDIR /run

COPY --from=builder . .

CMD ["./app"]
