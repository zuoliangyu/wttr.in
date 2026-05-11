# ========================
# Builder stage (Ubuntu/Debian-based)
# ========================
FROM golang:1.26-bookworm AS builder

WORKDIR /app

# Install build dependencies + fonts (Debian packages)
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    rsync \
    gcc \
    libc6-dev \
    fontconfig \
    fonts-dejavu-core \
    fonts-noto-core \
    fonts-noto-cjk \
    fonts-wqy-zenhei \
    fonts-symbola \
    fonts-motoya-l-cedar \
    fonts-lexi-gulim \
    && rm -rf /var/lib/apt/lists/*

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Copy source code
COPY . .

# Run the official build process
RUN bash build.sh build

# ========================
# Runtime stage (keep Alpine for small final image)
# ========================
FROM alpine:3.23

WORKDIR /app

# Minimal runtime dependencies
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1000 wttr && \
    mkdir -p /app/cache && \
    chown -R wttr:wttr /app

# Copy the built binary
COPY --from=builder /app/srv /app/bin/srv

# Environment variables
ENV WTTR_MYDIR="/app"
ENV WTTR_GEOLITE="/app/GeoLite2-City.mmdb"
ENV WTTR_LISTEN_HOST="0.0.0.0"
ENV WTTR_LISTEN_PORT="8002"

USER wttr

EXPOSE 8002

CMD ["/app/bin/srv"]
