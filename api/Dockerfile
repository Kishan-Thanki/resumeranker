FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/fixtures ./fixtures

EXPOSE 8080

CMD ["./api"]
