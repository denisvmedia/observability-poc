# syntax=docker/dockerfile:1
# Multi-stage Dockerfile for observability-poc

# Stage 1: Build frontend
FROM node:25-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./

RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY frontend/ ./

RUN npm run build

# Stage 2: Base Go environment
FROM golang:1.26.0-alpine AS go-base

RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

WORKDIR /app

COPY go/go.mod go/go.sum ./go/
COPY frontend/go.mod frontend/frontend.go ./frontend/

WORKDIR /app/go
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY go/ ./
COPY --from=frontend-builder /app/frontend/dist ../frontend/dist/

# Stage 3: Production builder
FROM go-base AS backend-builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /app/go/cmd/observability
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -tags with_frontend \
    -ldflags "-X github.com/denisvmedia/observability-poc/internal/version.Version=${VERSION} \
              -X github.com/denisvmedia/observability-poc/internal/version.Commit=${COMMIT} \
              -X github.com/denisvmedia/observability-poc/internal/version.Date=${BUILD_DATE}" \
    -a -installsuffix cgo \
    -o observability .

# Stage 4: Production runtime
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata curl

RUN addgroup -g 1001 -S observability && \
    adduser -u 1001 -S observability -G observability

RUN mkdir -p /app && chown -R observability:observability /app

WORKDIR /app

COPY --from=backend-builder /app/go/cmd/observability/observability .

RUN chown observability:observability observability

USER observability

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["./observability"]
CMD ["run"]

