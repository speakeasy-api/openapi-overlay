FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/overlay-cli ./main.go

FROM alpine:latest

COPY --from=builder /app/overlay-cli /usr/local/bin/overlay-cli

ENTRYPOINT [ "overlay-cli" ]

CMD ["--help"]