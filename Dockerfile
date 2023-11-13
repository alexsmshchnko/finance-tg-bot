FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY internal ./internal
COPY cmd ./cmd

RUN go build -o /app/finance-tg-bot ./cmd/

EXPOSE 8080

CMD ["/app/finance-tg-bot"]