FROM golang:1.27.3-alpine AS build
WORKDIR /app
COPY api-go/go.mod api-go/go.sum ./
RUN go mod download
COPY api-go/cmd ./cmd
COPY api-go/internal ./internal
COPY frontend /frontend
RUN go test ./... && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.23 AS runtime
LABEL org.opencontainers.image.title="api-go matrix challenge"
RUN addgroup -S app && adduser -S -G app app && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api /app/api
COPY frontend /app/frontend
ENV FRONTEND_DIR=/app/frontend
ENV PORT=8080
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/app/api"]