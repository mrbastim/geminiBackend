
FROM golang:1.25 AS builder

WORKDIR /src


COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -trimpath -ldflags="-s -w" -o /bin/server ./cmd/app


FROM debian:bookworm-slim


RUN apt-get update && \
	apt-get install -y --no-install-recommends ca-certificates tzdata wget && \
	rm -rf /var/lib/apt/lists/* && \
	groupadd --system app && useradd --system --gid app app

WORKDIR /app


COPY --from=builder /bin/server /app/server

ENV ENV=dev \
	LOG_LEVEL=info \
	LOG_FILE=server.log \
	PORT=5000 \
	DB_PATH=/app/data/data.db \
	TRUSTED_PROXIES=127.0.0.1


RUN mkdir -p /app/data && chown -R app:app /app

EXPOSE 5000

USER app

ENTRYPOINT ["/app/server"]
