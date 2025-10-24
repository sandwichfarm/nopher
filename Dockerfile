# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

# Allow Go to download required toolchain version
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk add --no-cache gcc musl-dev sqlite-dev
RUN CGO_ENABLED=1 GOOS=linux go build -a \
    -ldflags="-s -w" \
    -o nopher cmd/nopher/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 nopher && \
    adduser -D -u 1000 -G nopher nopher

WORKDIR /app

COPY --from=builder /build/nopher /usr/local/bin/nopher
COPY configs/nopher.example.yaml /etc/nopher/nopher.example.yaml

RUN mkdir -p /var/lib/nopher /etc/nopher/certs && \
    chown -R nopher:nopher /var/lib/nopher /etc/nopher

USER nopher

VOLUME ["/var/lib/nopher", "/etc/nopher"]

EXPOSE 70 1965 79

ENTRYPOINT ["/usr/local/bin/nopher"]
CMD ["--config", "/etc/nopher/nopher.yaml"]
