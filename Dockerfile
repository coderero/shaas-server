FROM golang:1.23.5-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

RUN apk add --no-cache make

# Build the application
RUN make

# Final stage
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/iot-server ./iot-server
COPY --from=builder /app/config /app/config

EXPOSE 8090

# Run the application
CMD ["./iot-server","--http", "0.0.0.0:8090"]