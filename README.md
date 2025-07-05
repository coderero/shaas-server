# IoT SmaAS Server

A smart home automation system server built with Go, PocketBase, and MQTT for managing IoT devices, sensors, and home automation controls.

## 🏠 Overview

The IoT SmaAS (Smart as a Service) Server is a comprehensive backend solution for smart home automation that provides:

- **Device Management**: Register and manage various IoT devices
- **Sensor Data Collection**: Real-time data from climate, motion, LDR (light), and other sensors
- **Relay Control**: Manage electrical relays for home automation
- **Security System**: RFID-based access control and security logging
- **Configuration Management**: Dynamic sensor and device configuration
- **User Authentication**: Secure user management with PocketBase
- **MQTT Communication**: Real-time bidirectional communication with IoT devices

## 🏗️ Architecture

The system consists of several key components:

- **PocketBase Backend**: Database and authentication layer
- **MQTT Server**: Real-time communication with IoT devices
- **Collection Management**: Structured data models for devices and sensors
- **Topic Handlers**: MQTT message processing for different device types
- **Security Layer**: RFID-based access control and logging

## 📋 Features

### Device Management

- Register and manage multiple IoT devices
- WiFi credential management for devices
- Device status tracking and monitoring

### Sensor Support

- **Climate Sensors**: Temperature, humidity, and air quality monitoring
- **Motion Sensors**: PIR motion detection with configurable settings
- **LDR Sensors**: Light-dependent resistor for ambient light sensing
- **Security Sensors**: RFID card readers for access control

### Automation & Control

- **Relay Control**: Manage electrical relays (low-duty and heavy-duty)
- **Port Management**: Configure and control individual relay ports
- **State Synchronization**: Real-time state sync between server and devices
- **Configuration Management**: Dynamic sensor configuration updates

### Security Features

- RFID-based access control
- Security event logging
- User-based device access control
- Admin-only administrative functions

## 🚀 Quick Start

### Prerequisites

- Go 1.19 or higher
- Docker (for containerized deployment)
- Make (for build automation)

### Local Development

1. **Clone the repository**

   ```bash
   git clone https://github.com/coderero/shaas-server.git
   cd smaas-server
   ```

2. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Build the application**

   ```bash
   make build
   ```

4. **Run the server**
   ```bash
   make run
   ```

The server will start with:

- HTTP API on port 8090
- MQTT broker on port 1883
- PocketBase admin UI available at `http://localhost:8090/_/`

### Docker Deployment

1. **Build the Docker image**

   ```bash
   docker build -t iot-server .
   ```

2. **Run with Docker**
   ```bash
   docker run -d \
     --name iot-server \
     -e ADMIN_EMAIL="your-admin@email.com" \
     -e ADMIN_PASSWORD="your-secure-password" \
     -e MQTT_USERNAME="mqtt-user" \
     -e MQTT_PASSWORD="mqtt-password" \
     -p 8090:8090 \
     -p 1883:1883 \
     iot-server
   ```

## ⚙️ Configuration

### Environment Variables

| Variable         | Description                     | Example             |
| ---------------- | ------------------------------- | ------------------- |
| `ADMIN_EMAIL`    | Admin user email for PocketBase | `admin@example.com` |
| `ADMIN_PASSWORD` | Admin user password             | `somthingsecure`    |
| `MQTT_USERNAME`  | MQTT broker username            | `mqtt_user`         |
| `MQTT_PASSWORD`  | MQTT broker password            | `mqtt_pass`         |

### MQTT Topics

The server listens to the following MQTT topic patterns:

| Topic Pattern          | Description            | Payload Type              |
| ---------------------- | ---------------------- | ------------------------- |
| `arduino/+/climate`    | Climate sensor data    | ClimateData (protobuf)    |
| `arduino/+/ldr`        | Light sensor data      | LDRData (protobuf)        |
| `arduino/+/motion`     | Motion sensor data     | MotionData (protobuf)     |
| `arduino/+/relay`      | Relay control commands | RelayState (protobuf)     |
| `arduino/+/relay/full` | Full relay state sync  | RelayStateSync (protobuf) |
| `arduino/+/rfid`       | RFID security events   | RfidEnvelope (protobuf)   |

### Published Topics

The server publishes to these topics:

| Topic Pattern                       | Description           | Purpose                    |
| ----------------------------------- | --------------------- | -------------------------- |
| `arduino/{device_id}/config`        | Device configuration  | Send sensor configs        |
| `arduino/{device_id}/config/remove` | Configuration removal | Remove sensor configs      |
| `arduino/{device_id}/relay`         | Relay commands        | Control relay states       |
| `arduino/{device_id}/rfid`          | RFID commands         | Register/revoke RFID cards |

## 📊 Database Collections

### Core Collections

#### Devices

- **Purpose**: Main device registry
- **Fields**: `user`, `device_name`, `device_status`, `timestamp`
- **Access**: User-specific (users can only see their own devices)

#### Sensor Data Collections

**Climate**

- **Fields**: `device`, `sensor_id`, `temperature`, `humidity`, `air_quality`, `timestamp`
- **Purpose**: Store climate sensor readings

**LDR (Light Sensors)**

- **Fields**: `device`, `sensor_id`, `ldr_value`, `timestamp`
- **Purpose**: Store light sensor readings

**Motion**

- **Fields**: `device`, `sensor_id`, `motion_detected`, `timestamp`
- **Purpose**: Store motion detection events

#### Configuration Collections

**Climate Config**

- **Fields**: `device`, `sensor_id`, `label`, `dht22_port`, `aqi_port`, `has_buzzer`, `buzzer_port`
- **Purpose**: Configure climate sensors

**LDR Config**

- **Fields**: `device`, `sensor_id`, `label`, `port`
- **Purpose**: Configure light sensors

**Motion Config**

- **Fields**: `device`, `sensor_id`, `label`, `port`, `relay_type`, `relay_port`
- **Purpose**: Configure motion sensors

#### Control Collections

**User Port Labels**

- **Fields**: `device`, `relay`, `port`, `state`, `label`
- **Purpose**: Manage relay port states and labels

**Relay**

- **Fields**: `type`, `switches`
- **Purpose**: Define relay types (low-duty: 4 switches, heavy-duty: 2 switches)

#### Security Collections

**Security**

- **Fields**: `device`, `uuid`
- **Purpose**: Registered RFID cards

**Security Logs**

- **Fields**: `device`, `uuid`, `level`, `details`
- **Purpose**: Security access logs

## 🔌 API Endpoints

### Authentication

All API endpoints require authentication except for the admin setup.

### Device Management

- `GET /api/collections/devices/records` - List user's devices
- `POST /api/collections/devices/records` - Register a new device
- `PATCH /api/collections/devices/records/{id}` - Update device
- `DELETE /api/collections/devices/records/{id}` - Delete device

### Sensor Data

- `GET /api/collections/climate/records` - Get climate data
- `GET /api/collections/ldr/records` - Get light sensor data
- `GET /api/collections/motion/records` - Get motion sensor data

### Configuration

- `POST /api/collections/climate_config/records` - Add climate sensor config
- `POST /api/collections/ldr_config/records` - Add light sensor config
- `POST /api/collections/motion_config/records` - Add motion sensor config

### Relay Control

- `GET /api/collections/user_port_labels/records` - Get relay states
- `PATCH /api/collections/user_port_labels/records/{id}` - Control relay

### Security

- `POST /api/collections/security/records` - Register RFID card
- `DELETE /api/collections/security/records/{id}` - Revoke RFID card
- `GET /api/collections/security_logs/records` - View security logs

## 🔧 Development

### Project Structure

```
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── collections/            # Database schema definitions
│   │   ├── collection.go       # Collection interface
│   │   ├── config.go          # Configuration collections
│   │   ├── device.go          # Device and sensor collections
│   │   └── security.go        # Security collections
│   ├── proto/
│   │   └── transporter/       # Generated protobuf code
│   ├── server/
│   │   ├── mqtt.go           # MQTT server setup
│   │   ├── pocketbase.go     # PocketBase setup
│   │   ├── server.go         # Main server coordination
│   │   └── triggers.go       # Database event triggers
│   └── topics/
│       └── arduino.go        # MQTT topic handlers
├── pkg/
│   └── proto/
│       └── transporter.proto # Protocol buffer definitions
└── Dockerfile               # Container configuration
```

### Building

```bash
# Build binary
make build

# Run locally
make run

# Build Docker image
docker build -t iot-server .
```

### Testing MQTT

You can test MQTT communication using mosquitto clients:

```bash
# Subscribe to all arduino topics
mosquitto_sub -h localhost -p 1883 -t "arduino/#" -u mqtt_user -P mqtt_pass

# Publish climate data (requires protobuf encoding)
mosquitto_pub -h localhost -p 1883 -t "arduino/device123/climate" -f climate_data.bin -u mqtt_user -P mqtt_pass
```

## 🛡️ Security

### Access Control

- **User Isolation**: Users can only access their own devices and data
- **Admin Functions**: Relay control and system configuration require admin privileges
- **MQTT Authentication**: MQTT broker requires username/password authentication

### Data Protection

- Environment variables for sensitive configuration
- Secure password hashing via PocketBase
- Input validation on all API endpoints

## 📈 Monitoring

### Logging

The server provides structured logging for:

- MQTT message processing
- Database operations
- Security events
- Configuration changes
- Error conditions

### Health Checks

- HTTP endpoint availability on port 8090
- MQTT broker status on port 1883
- Database connectivity through PocketBase

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation for API changes
- Use structured logging with appropriate log levels

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For questions or support:

- Create an issue in the GitHub repository
- Check the PocketBase documentation for database-related questions
- Refer to MQTT documentation for messaging protocol details

---

**Built with ❤️ for the IoT community**
