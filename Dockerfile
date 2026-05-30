# syntax=docker/dockerfile:1.7

FROM golang:1.24.2-alpine AS builder

ARG VERSION=0.1.0
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /src

RUN apk --no-cache add ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /out/kextant-agent ./cmd/agent

FROM alpine:3.21

ARG VERSION=0.1.0
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="Kextant Agent" \
      org.opencontainers.image.description="Read-only Kubernetes health scanner and inventory collector for Kextant Version Intelligence" \
      org.opencontainers.image.source="https://github.com/kextant/kextant-agent" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}"

RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -u 1000 -h /app kextant

WORKDIR /app
COPY --from=builder /out/kextant-agent /app/kextant-agent

USER 1000:1000
EXPOSE 8080

ENTRYPOINT ["/app/kextant-agent"]
CMD ["serve"]
