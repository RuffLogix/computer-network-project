# Backend Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary
COPY --from=builder /app/server .

# Copy uploads directory if needed
COPY --from=builder /app/uploads ./uploads

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./server"]