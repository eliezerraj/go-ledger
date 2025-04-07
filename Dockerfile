#docker build -t go-oauth .
#docker run -dit --name go-card -p 6100:6100 go-card sleep infinity

FROM golang:1.23.3 As builder

RUN apt-get update && apt-get install bash && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY . .
RUN go mod tidy

WORKDIR /app/cmd
RUN go build -o go-card -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-card .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/go-card"]