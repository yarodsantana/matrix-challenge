import jwt from 'jsonwebtoken';

const CLOCK_TOLERANCE_SECONDS = 5;

export function requireServiceToken(secret) {
  if (typeof secret !== 'string' || Buffer.byteLength(secret) < 32) {
    throw new Error('SERVICE_JWT_SECRET requiere al menos 32 bytes');
  }

  return (req, res, next) => {
    const match = /^Bearer\s+(\S+)$/i.exec(req.get('Authorization') ?? '');
    try {
      if (!match) throw new Error('Token ausente');
      const claims = jwt.verify(match[1], secret, {
        algorithms: ['HS256'],
        clockTolerance: CLOCK_TOLERANCE_SECONDS,
        issuer: 'matrix-api-go',
        audience: 'matrix-api-node',
        subject: 'matrix-api-go',
      });
      // jsonwebtoken comprueba exp si existe; nuestro contrato lo exige.
      if (!Number.isFinite(claims.exp)) throw new Error('Falta expiracion');
      if (claims.iat !== undefined && (!Number.isFinite(claims.iat) || claims.iat > Date.now() / 1000 + CLOCK_TOLERANCE_SECONDS)) {
        throw new Error('Fecha de emision invalida');
      }
    } catch {
      res.set('WWW-Authenticate', 'Bearer');
      return res.status(401).json({ error: 'token ausente o invalido' });
    }
    return next();
  };
}