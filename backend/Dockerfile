FROM golang:1.24-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/dsm-backend .

FROM python:3.11-slim

WORKDIR /app

ENV PYTHONUNBUFFERED=1 \
    PYTHON_BIN=/usr/local/bin/python \
    SCRIPT_STORAGE_DIR=/app/storage/scripts \
    SCRIPT_STATE_DIR=/app/storage/state \
    TZ=Asia/Shanghai

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        curl \
        nodejs \
        tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/dsm-backend /app/dsm-backend

RUN mkdir -p /app/storage/scripts /app/storage/state

EXPOSE 8080

CMD ["/app/dsm-backend"]
