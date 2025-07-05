# === Build Stage ===
FROM golang:1.24-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go module files first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Install build tools
RUN apk update && apk add --no-cache make

# Build the application using Makefile
RUN make


# === Runtime Stage ===
FROM alpine:3.21.3

# Working directory inside the container
WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/bin/iot-server ./iot-server

# Set executable permissions (optional but safe)
RUN chmod +x ./iot-server

# Expose the default HTTP port and MQTT port
EXPOSE 8090
EXPOSE 1883

# Start the server
CMD ["./iot-server", "serve", "--http=0.0.0.0:8090"]
