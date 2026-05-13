# Build stage
FROM golang:1.21 as builder

WORKDIR /app
COPY . ./

RUN go build -o main ./cmd/server

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]