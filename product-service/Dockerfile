# ---- Build Stage ----
# Use the official Golang image to build the application.
# The `AS builder` part names this stage.
FROM golang:1.22-alpine AS builder

# ARG declares a build-time variable. The CI will pass the service name here.
ARG SERVICE_NAME

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker's build cache.
# This step is only re-run if these files change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code into the container
COPY . .

# Build the Go application.
# This is a static build, which is great for containers.
# The command builds the specific service passed in via SERVICE_NAME.
# Example: go build -o /app/main ./cmd/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./cmd/${SERVICE_NAME}

# ---- Final Stage ----
# Start from a minimal base image for the final container.
# alpine is very small and secure.
FROM alpine:latest

# Create a non-root user and group for security.
# Running containers as a non-root user is a critical security best practice.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory for the final image
WORKDIR /app

# Copy ONLY the compiled binary from the 'builder' stage.
# This is the key to a small final image. No source code is included.
COPY --from=builder /app/main .

# Switch to the non-root user
USER appuser

# The ENTRYPOINT is the command that will be run when the container starts.
ENTRYPOINT ["./main"]
