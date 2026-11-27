# syntax=docker/dockerfile:1
FROM golang:1.24.13-alpine AS api-build
WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/casino ./cmd/casino && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bootstrap-admin ./cmd/bootstrap-admin && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed-h5 ./cmd/seed-h5 && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/check-config ./cmd/check-config

FROM alpine:3.22 AS api
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S casino && adduser -S -G casino -u 10001 casino
WORKDIR /app
COPY --from=api-build /out/ /app/
COPY server/configs/casino.yml /app/configs/casino.yml
COPY --chown=casino:casino server/static/ /app/static/
# Serve bundled H5 media from both portals through the API's /static/ route.
COPY --chown=casino:casino apps/h5/public/static/ /app/static/
RUN mkdir -p /app/backups && chown -R casino:casino /app/static /app/backups
USER casino
EXPOSE 1639
CMD ["/app/casino", "-c", "/app/configs/casino.yml"]

FROM node:22-alpine AS h5-build
WORKDIR /src
COPY apps/h5/package*.json ./
RUN npm ci --no-audit --no-fund
COPY apps/h5/ ./
RUN npm run build

FROM node:22-alpine AS admin-build
WORKDIR /src
COPY apps/admin/package*.json ./
RUN npm ci --no-audit --no-fund
COPY apps/admin/ ./
RUN npm run build


FROM nginx:alpine AS web
COPY deploy/nginx/ /etc/nginx/
COPY --from=h5-build /src/dist/ /usr/share/nginx/html/h5/
COPY --from=admin-build /src/dist/ /usr/share/nginx/html/admin/
EXPOSE 8080 8081
