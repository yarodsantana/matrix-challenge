import express from 'express';
import { requireServiceToken } from './auth.js';
import { calculateStatistics, calculateQRStatistics } from './statistics.js';

// El calculo se inyecta por ruta para compartir HTTP sin duplicar reglas matematicas.
function statisticsHandler(calculate) {
  return (req, res, next) => {
    try {
      return res.json(calculate(req.body));
    } catch (error) {
      if (error instanceof TypeError) return res.status(400).json({ error: error.message });
      if (error instanceof RangeError) return res.status(422).json({ error: error.message });
      return next(error);
    }
  };
}

function requireJSON(req, res, next) {
  if (!req.is('application/json')) {
    return res.status(415).json({ error: 'Content-Type debe ser application/json' });
  }
  return next();
}

export function createApp({ serviceSecret = process.env.SERVICE_JWT_SECRET } = {}) {
  const authenticate = requireServiceToken(serviceSecret);
  const app = express();
  app.disable('x-powered-by');
  app.get('/health', (req, res) => res.json({ status: 'ok' }));
  const parseJSON = express.json({ limit: '1mb' });
  // Mantener /statistics conserva el contrato del bloque de rotacion.
  app.post('/statistics', authenticate, requireJSON, parseJSON, statisticsHandler(calculateStatistics));
  app.post('/statistics/qr', authenticate, requireJSON, parseJSON, statisticsHandler(calculateQRStatistics));
  app.use((req, res) => res.status(404).json({ error: 'Ruta no encontrada' }));
  app.use((error, req, res, next) => {
    if (res.headersSent) return next(error);
    if (error.type === 'entity.parse.failed') return res.status(400).json({ error: 'JSON invalido' });
    if (error.type === 'entity.too.large') return res.status(413).json({ error: 'El cuerpo supera el limite de 1 MiB' });
    if (error.status === 415) return res.status(415).json({ error: 'Codificacion o charset no admitido' });
    console.error(error);
    return res.status(500).json({ error: 'Error interno del servidor' });
  });
  return app;
}