FROM golang:1.16 AS builder

WORKDIR /go/src/app
COPY . .

RUN go get
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch

ADD ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/app/liquibase-lock-guard /

CMD ["/liquibase-lock-guard"]
