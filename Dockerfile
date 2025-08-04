# ---- Build Stage ----
# Use the official Golang image to build the application.
FROM golang:1.23-alpine AS builder

# ARG declares a build-time variable. The CI will pass the service name here.
ARG SERVICE_NAME

# Set the working directory inside the container
WORKDIR /app

# Copy the entire repository to get all dependencies
COPY . .

# Change to the service directory and build
WORKDIR /app/${SERVICE_NAME}

# Download dependencies for this specific service
RUN go mod download

# Build the Go application.
# This is a static build, which is great for containers.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./cmd/main.go

# ---- Final Stage ----
# Start from a minimal base image for the final container.
FROM alpine:latest

# Create a non-root user and group for security.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory for the final image
WORKDIR /app

# Copy ONLY the compiled binary from the 'builder' stage.
COPY --from=builder /app/main .

# Switch to the non-root user
USER appuser

# The ENTRYPOINT is the command that will be run when the container starts.
ENTRYPOINT ["./main"]
