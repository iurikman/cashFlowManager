FROM golang:1.22 as builder

WORKDIR /usr/src/app
COPY . .
RUN go mod download
RUN go build -v -o cashFlowManager ./...

FROM debian:stable-slim
WORKDIR /bin
COPY -- from=builder /usr/src/app/cashFlowManager ./
ENTRYPOINT ["./cashFlowManager"]
