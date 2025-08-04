# Copy pre-built artifacts
FROM scratch AS build-stage

WORKDIR /app

COPY bin/main bin/main
COPY .env .env

# Runtime stage
FROM alpine:3.20 AS build-release-stage

WORKDIR /

RUN addgroup -S executorgroup && adduser -S executor -G executorgroup
USER executor

COPY --from=build-stage /app/bin/main /main
COPY --from=build-stage /app/.env /.env

# Notification service listens on port 8080
EXPOSE 8080 

ENTRYPOINT [ "/main" ]
