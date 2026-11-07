# syntax=docker/dockerfile:1

# Etapa 1: compila Tailwind CSS y copia HTMX.
FROM node:22-alpine AS activos
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY web/plantillas ./web/plantillas
COPY web/estaticos/css/entrada.css ./web/estaticos/css/entrada.css
RUN mkdir -p web/estaticos/js \
    && npm run css \
    && npm run htmx \
    && npm run qr

# Etapa 2: compila el binario de Go sin dependencias del sistema.
FROM golang:1.27-alpine AS compilacion
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /salida/servidor ./cmd/servidor

# Etapa 3: imagen final mínima, sin herramientas de compilación.
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S asistencia \
    && adduser -S -G asistencia asistencia

WORKDIR /app
COPY --from=compilacion /salida/servidor /app/servidor
COPY --from=activos /app/web/estaticos /app/web/estaticos
RUN mkdir -p /app/datos/fotos \
    && chown -R asistencia:asistencia /app

USER asistencia
ENV PUERTO=8080 \
    ALMACENAMIENTO_TIPO=disco \
    ALMACENAMIENTO_RUTA=/app/datos/fotos

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PUERTO}/salud" > /dev/null || exit 1

ENTRYPOINT ["/app/servidor"]
