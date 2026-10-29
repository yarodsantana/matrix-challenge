import { createRequire } from 'node:module';
const require = createRequire(new URL('../api-node/package.json', import.meta.url));
const jwt = require('jsonwebtoken');

const secret = process.env.JWT_SECRET;
if (typeof secret !== 'string' || Buffer.byteLength(secret) < 32) {
  throw new Error('JWT_SECRET requiere al menos 32 bytes; ejecuta con --env-file=.env');
}
// Herramienta local del desarrollador. No se publica un endpoint que emita tokens.
const token = jwt.sign({}, secret, {
  algorithm: 'HS256', issuer: 'matrix-challenge', audience: 'matrix-api-go',
  subject: 'technical-test-user', expiresIn: '15m',
});
console.log(token);