# Use the Go 1.23.2 bookworm image for building the application
FROM golang:1.23.2-bookworm AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the current directory contents into the container
COPY . ./

# Download Go module dependencies
RUN go mod download

# Build the Go application, setting CGO_ENABLED=0 to build a static binary
RUN CGO_ENABLED=0 go build -o /bin/app

# Use a minimal distroless image for the final build
FROM gcr.io/distroless/static-debian11

# Copy the compiled binary from the build stage to the final image
COPY --from=build /bin/app /bin
COPY .env /app/.env
# Set the command to run the application
ENTRYPOINT [ "/bin/app" ]
