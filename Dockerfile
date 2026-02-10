FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

RUN go install github.com/benbjohnson/ego/cmd/ego@latest

WORKDIR /go/src/github.com/hunter-io/faktory

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate github.com/hunter-io/faktory/webui
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /faktory cmd/faktory/daemon.go

FROM alpine:3.21
RUN apk add --no-cache redis ca-certificates socat

COPY --from=builder /faktory /faktory

RUN mkdir -p /root/.faktory/db
RUN mkdir -p /var/lib/faktory/db
RUN mkdir -p /etc/faktory

EXPOSE 7419 7420
CMD ["/faktory", "-w", "0.0.0.0:7420", "-b", "0.0.0.0:7419", "-e", "development"]
