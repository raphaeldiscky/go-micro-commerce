# ---- Build Stage ----
# Use the official Golang image to build the application.
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app/auth-service

# Copy go.mod and go.sum from the root of the repository.
# We copy them here so that we can leverage Docker's build cache.
# This assumes the go.mod and go.sum are at the project root.
COPY go.mod go.sum /app/
RUN go mod download

# Copy the rest of the source code into the container
COPY . /app

# Build the Go application.
# The command builds the specific service, which is now hardcoded.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/auth-service/main ./cmd/main.go

# ---- Final Stage ----
# Start from a minimal base image for the final container.
FROM alpine:latest

# Create a non-root user and group for security.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory for the final image
WORKDIR /app/auth-service

# Copy ONLY the compiled binary from the 'builder' stage.
# This is the key to a small final image.
COPY --from=builder /app/auth-service/main .

# Switch to the non-root user
USER appuser

# The ENTRYPOINT is the command that will be run when the container starts.
ENTRYPOINT ["./main"]