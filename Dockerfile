FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /mutating-webhook cmd/main.go

# FROM scratch
FROM ubuntu:22.04
COPY --from=builder /mutating-webhook .
EXPOSE 8443
CMD ["./mutating-webhook"]
