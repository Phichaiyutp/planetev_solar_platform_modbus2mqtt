FROM golang:1.23.2 AS build-stage

# Set environment variables for the build
ENV GOPATH /go
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . ./

# Build the application with specific settings
RUN CGO_ENABLED=0 GOOS=linux go build -o /modbus-mqtt-service

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /app

# Copy the binary from build stage
COPY --from=build-stage /modbus-mqtt-service /app/modbus-mqtt-service

# Copy environment file
# Note: Make sure to create a .env file in your project directory
COPY .env /app/.env

# Optional: Set environment variables directly in Dockerfile
# ENV KEY=value
# ENV DATABASE_URL=your-database-url
# ENV MQTT_BROKER=your-mqtt-broker

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/modbus-mqtt-service"]