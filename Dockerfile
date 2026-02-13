# =============================================================================
# Stage 1: Build frontend
# =============================================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

# Install dependencies first (layer caching)
COPY web/package.json web/package-lock.json* ./
RUN npm install

# Copy frontend source and build
COPY web/ ./
RUN npm run build

# =============================================================================
# Stage 2: Build backend
# =============================================================================
FROM golang:1.22-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Download Go modules first (layer caching)
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=1 go build -o bin/agentsocial ./cmd/server/

# =============================================================================
# Stage 3: Runtime
# =============================================================================
FROM alpine:3.20

WORKDIR /opt/agentsocial

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S agentsocial && adduser -S agentsocial -G agentsocial

# Create data directory
RUN mkdir -p /opt/agentsocial/data && chown -R agentsocial:agentsocial /opt/agentsocial/data

# Copy backend binary
COPY --from=backend-builder /app/bin/agentsocial .

# Copy frontend build output
COPY --from=frontend-builder /app/web/dist ./web/dist

# Switch to non-root user
USER agentsocial

EXPOSE 8080

ENTRYPOINT ["./agentsocial"]
