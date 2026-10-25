
FROM node:22-alpine AS dashboard-builder
WORKDIR /dashboard
COPY dashboard/package*.json ./
RUN npm ci
COPY dashboard/ ./
RUN npm run build


FROM golang:1.26.6-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o zee-mirror ./cmd/zee-mirror


# Local Bot API server, built on the same Alpine release as the runtime image
# so its libstdc++/openssl links match. The bot starts it itself with the
# api_id/api_hash stored in the database.
FROM alpine:3.20 AS botapi-builder
ARG TELEGRAM_BOT_API_REF=master
RUN apk add --no-cache alpine-sdk linux-headers git zlib-dev openssl-dev gperf cmake
RUN git clone --recursive --shallow-submodules --depth 1 --branch "${TELEGRAM_BOT_API_REF}" \
        https://github.com/tdlib/telegram-bot-api.git /src
WORKDIR /src/build
RUN cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local .. && \
    cmake --build . --target install -j"$(nproc)" && \
    strip /usr/local/bin/telegram-bot-api


FROM alpine:3.20
LABEL maintainer="Zee-Mirror Bot"
LABEL description="Telegram Mirror/Leech Bot"

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    aria2 \
    ffmpeg \
    python3 \
    py3-pip \
    p7zip \
    curl \
    unzip \
    nodejs \
    py3-cryptography \
    py3-pycryptodomex \
    coreutils \
    expect \
    libstdc++ \
    libssl3 \
    libcrypto3 \
    zlib && \
    apk add --no-cache --virtual .build-deps gcc musl-dev python3-dev && \
    pip3 install --break-system-packages --no-cache-dir -U yt-dlp speedtest-cli && \
    apk del .build-deps

# gcc/musl-dev/python3-dev dipasang sebagai .build-deps sementara (untuk wheel builds),
# lalu dihapus via apk del agar runtime image tetap ramping.
# Menggunakan --no-cache-dir dan memastikan download versi terbaru setiap build
# Cache bust for yt-dlp: 2026-06-16


COPY --from=rclone/rclone:1.75.1 /usr/local/bin/rclone /usr/bin/rclone
RUN chmod 755 /usr/bin/rclone

COPY --from=botapi-builder /usr/local/bin/telegram-bot-api /usr/local/bin/telegram-bot-api

RUN addgroup -S botgroup && adduser -D -G botgroup botuser
RUN mkdir -p /app/downloads /app/config /app/data /home/botuser/.cache/yt-dlp /var/lib/telegram-bot-api && \
    chown -R botuser:botgroup /app /home/botuser /var/lib/telegram-bot-api

WORKDIR /app
COPY --from=builder /build/zee-mirror /app/zee-mirror
COPY --from=builder /build/migrations /app/migrations
COPY --from=dashboard-builder /dist /app/dist
RUN chmod +x /app/zee-mirror && \
    chown -R botuser:botgroup /app

USER botuser

# Only DATABASE_URL (and ENCRYPTION_KEY when values are encrypted) is needed
# at runtime; every other setting is read from the database settings table.
# Don't add app settings here: a non-empty environment variable overrides
# the database value.
ENV HOME="/home/botuser" \
    PATH="/usr/local/bin:/usr/bin:/bin:/home/botuser/.local/bin"

EXPOSE 5248

VOLUME ["/app/downloads", "/app/config", "/var/lib/telegram-bot-api"]

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:5248/api/health || exit 1

CMD ["/app/zee-mirror"]