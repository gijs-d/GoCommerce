# Fase 1: Frontend compileren
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --prefer-offline || npm install
COPY frontend/ ./
RUN npm run build

# Fase 2: Go Backend compileren
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o shop-server ./cmd/server

# Fase 3: Minimale productie-image (~30MB)
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=backend-builder /app/shop-server /app/shop-server
COPY --from=backend-builder /app/migrations /app/migrations
COPY --from=backend-builder /app/themes /app/themes
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

RUN mkdir -p /app/uploads

ENV PORT=8080
ENV APP_ENV=production
EXPOSE 8080

CMD ["/app/shop-server"]