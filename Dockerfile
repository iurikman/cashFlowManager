FROM golang:1.22 as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install github.com/vektra/mockery/v2@latest
RUN go generate ./...
RUN go build -v -o cashFlowManager ./...

FROM debian:stable-slim
WORKDIR /bin
COPY --from=builder /usr/src/app/cashFlowManager ./
ENTRYPOINT ["./cashFlowManager"]
