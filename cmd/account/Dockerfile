FROM golang:1.26-alpine AS builder

WORKDIR /build


COPY go.work go.work.sum ./

COPY db/go.mod db/go.sum ./db/

COPY proto/go.mod proto/go.sum ./proto/

COPY cmd/account/go.mod cmd/account/go.sum ./cmd/account/


RUN go mod download

COPY db/ ./db/
COPY proto/ ./proto/
COPY cmd/account/ ./cmd/account/

WORKDIR /build/cmd/account

RUN CGO_ENABLED=0 go build -o account .

#==-==-==-==-==-==-==-==-==-==-==-==-==-==

FROM scratch

WORKDIR /app

COPY --from=builder /build/cmd/account/account .

CMD ["./account"]
