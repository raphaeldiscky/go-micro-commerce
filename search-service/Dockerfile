# ---- Build Stage ----
# Use the official Golang image to build the application.
FROM golang:1.25.3-alpine AS builder

# Set the working directory to match the repo structure
WORKDIR /workspace

# Copy only the necessary files for this service
COPY pkg/ ./pkg/
COPY search-service/go.mod search-service/go.sum ./search-service/

# Change to the service directory for building
WORKDIR /workspace/search-service
 

RUN go mod download

# Copy the rest of the service files
COPY search-service/ ./

# Build the Go application.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api/main.go

# ---- Final Stage ----
# Start from a minimal base image for the final container.
FROM alpine:latest

# Create a non-root user and group for security.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory for the final image
WORKDIR /app

# Copy ONLY the compiled binary from the 'builder' stage.
# This is the key to a small final image.
COPY --from=builder /workspace/search-service/main .

EXPOSE 8087

# Switch to the non-root user
USER appuser

# The ENTRYPOINT is the command that will be run when the container starts.
ENTRYPOINT [ "/main"]
CMD ["serve-all"]
