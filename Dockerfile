# === Build Stage ===
FROM golang:1.24.2-alpine AS builder

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
FROM alpine:3.19

# Working directory inside the container
WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/bin/iot-server ./iot-server

# Set executable permissions (optional but safe)
RUN chmod +x ./iot-server

# Expose the default HTTP port
EXPOSE 8090

# Define environment variables (for local dev defaults, can be overridden)
ENV MQTT_USERNAME=defaultuser
ENV MQTT_PASSWORD=defaultpass

# Start the server
CMD ["./iot-server", "serve"]
