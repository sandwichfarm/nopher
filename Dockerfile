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
    -o nophr cmd/nophr/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 nophr && \
    adduser -D -u 1000 -G nophr nophr

WORKDIR /app

COPY --from=builder /build/nophr /usr/local/bin/nophr
COPY configs/nophr.example.yaml /etc/nophr/nophr.example.yaml

RUN mkdir -p /var/lib/nophr /etc/nophr/certs && \
    chown -R nophr:nophr /var/lib/nophr /etc/nophr

USER nophr

VOLUME ["/var/lib/nophr", "/etc/nophr"]

EXPOSE 70 1965 79

ENTRYPOINT ["/usr/local/bin/nophr"]
CMD ["--config", "/etc/nophr/nophr.yaml"]
