import test from 'node:test';
import assert from 'node:assert/strict';
import { once } from 'node:events';
import jwt from 'jsonwebtoken';
import { createApp } from '../src/app.js';

const serviceSecret = 'test-only-service-secret-32-bytes-minimum';
const wrongSecret = 'test-only-client-secret-32-bytes-minimum';
const sign = (claims = {}, options = {}, key = serviceSecret) => jwt.sign(claims, key, {
  algorithm: 'HS256', issuer: 'matrix-api-go', audience: 'matrix-api-node',
  subject: 'matrix-api-go', expiresIn: '1m', ...options,
});

test('JWT protege las estadisticas antes de leer la matriz', async (t) => {
  const server = createApp({ serviceSecret }).listen(0, '127.0.0.1');
  t.after(() => new Promise((resolve, reject) => server.close((err) => err ? reject(err) : resolve())));
  await once(server, 'listening');
  const url = `http://127.0.0.1:${server.address().port}/statistics`;
  const cases = [
    ['valido', 'Bearer ' + sign(), 200],
    ['desfase de reloj pequeno', 'Bearer ' + sign({ iat: Math.floor(Date.now() / 1000) + 3 }), 200],
    ['ausente', null, 401],
    ['mal formado', 'Bearer nope', 401],
    ['otro esquema', 'Basic abc', 401],
    ['firma incorrecta', 'Bearer ' + sign({}, {}, wrongSecret), 401],
    ['vencido', 'Bearer ' + sign({}, { expiresIn: -60 }), 401],
    ['audiencia incorrecta', 'Bearer ' + sign({}, { audience: 'matrix-api-go' }), 401],
    ['emisor incorrecto', 'Bearer ' + sign({}, { issuer: 'otro' }), 401],
    ['sujeto incorrecto', 'Bearer ' + sign({}, { subject: 'otro' }), 401],
    ['algoritmo no permitido', 'Bearer ' + sign({}, { algorithm: 'HS384' }), 401],
    ['sin expiracion', 'Bearer ' + jwt.sign({}, serviceSecret, { algorithm: 'HS256', issuer: 'matrix-api-go', audience: 'matrix-api-node', subject: 'matrix-api-go' }), 401],
    ['emision futura', 'Bearer ' + sign({ iat: Math.floor(Date.now() / 1000) + 60 }), 401],
    ['aun no valido', 'Bearer ' + sign({}, { notBefore: '1m' }), 401],
  ];
  for (const [name, authorization, status] of cases) {
    await t.test(name, async () => {
      const headers = { 'Content-Type': 'application/json' };
      if (authorization !== null) headers.Authorization = authorization;
      const response = await fetch(url, { method: 'POST', headers, body: '[[1]]', signal: AbortSignal.timeout(5000) });
      assert.equal(response.status, status);
      const body = await response.json();
      if (status === 401) {
        assert.equal(response.headers.get('www-authenticate'), 'Bearer');
        assert.deepEqual(body, { error: 'token ausente o invalido' });
      } else { assert.equal(body.sum, 1); }
    });
  }
});

test('no arranca sin clave de servicio valida', () => {
  assert.throws(() => createApp({ serviceSecret: '' }));
  assert.throws(() => createApp({ serviceSecret: 'short' }));
});