FROM node:24.22.1-alpine AS dependencies
WORKDIR /app
COPY api-node/package.json api-node/package-lock.json ./
RUN npm ci --omit=dev --no-audit --no-fund

FROM dependencies AS test
COPY api-node/src ./src
COPY api-node/test ./test
RUN npm test

FROM node:24.22.1-alpine AS runtime
LABEL org.opencontainers.image.title="api-node matrix challenge"
ENV NODE_ENV=production
ENV PORT=3000
WORKDIR /app
# Copiar desde test obliga a ejecutar las pruebas durante la construccion.
COPY --from=test --chown=node:node /app/node_modules ./node_modules
COPY --from=test --chown=node:node /app/package.json ./package.json
COPY --from=test --chown=node:node /app/src ./src
USER node
EXPOSE 3000
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:' + (process.env.PORT || 3000) + '/health').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"
CMD ["node", "src/server.js"]