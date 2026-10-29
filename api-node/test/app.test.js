import jwt from 'jsonwebtoken';
const serviceSecret = 'test-only-service-secret-32-bytes-minimum';
const serviceToken = () => jwt.sign({}, serviceSecret, { algorithm: 'HS256', issuer: 'matrix-api-go', audience: 'matrix-api-node', subject: 'matrix-api-go', expiresIn: '1m' });
import test from 'node:test';
import assert from 'node:assert/strict';
import { once } from 'node:events';
import { createApp } from '../src/app.js';

// Puerto 0: el sistema asigna uno disponible para evitar conflictos.
test('contrato HTTP de POST /statistics', { timeout: 15000 }, async (t) => {
  const server = createApp({ serviceSecret }).listen(0, '127.0.0.1');
  t.after(() => new Promise((resolve, reject) => {
    server.close((error) => error ? reject(error) : resolve());
  }));
  await once(server, 'listening');
  const baseURL = `http://127.0.0.1:${server.address().port}`;

  const cases = [
    { name: 'matriz rectangular', body: '[[4,1],[5,2],[6,3]]', status: 200,
      expected: { max: 6, min: 1, average: 3.5, sum: 21, isDiagonal: false } },
    { name: 'diagonal', body: '[[-2,0],[0,4]]', status: 200,
      expected: { max: 4, min: -2, average: 0.5, sum: 2, isDiagonal: true } },
    { name: 'vacia', body: '[]', status: 400 },
    { name: 'irregular', body: '[[1,2],[3]]', status: 400 },
    { name: 'elemento null', body: '[[null]]', status: 400 },
    { name: 'texto', body: '[["2"]]', status: 400 },
    { name: 'objeto', body: '{"matrix":[[1]]}', status: 400 },
    { name: 'JSON mal formado', body: '[[1]', status: 400 },
    { name: 'JSON null', body: 'null', status: 400 },
    { name: 'sin cuerpo', body: '', status: 400 },
    { name: 'sin content type', body: '[[1]]', contentType: null, status: 415 },
    { name: 'text/plain', body: '[[1]]', contentType: 'text/plain', status: 415 },
    { name: 'charset UTF-8', body: '[[7]]', contentType: 'application/json; charset=utf-8', status: 200,
      expected: { max: 7, min: 7, average: 7, sum: 7, isDiagonal: true } },
    { name: 'desbordamiento', body: '[[1.7976931348623157e308,1.7976931348623157e308]]', status: 422 },
    { name: 'limite del cuerpo', body: `[[0]]${' '.repeat(1024 * 1024)}`, status: 413 },
    { name: 'ruta inexistente', path: '/missing', body: '[[1]]', status: 404 },
  ];

  for (const tc of cases) {
    await t.test(tc.name, async () => {
      const headers = { Authorization: "Bearer " + serviceToken() };
      const contentType = tc.contentType === undefined ? 'application/json' : tc.contentType;
      if (contentType !== null) headers['Content-Type'] = contentType;
      const response = await fetch(`${baseURL}${tc.path ?? '/statistics'}`, {
        method: 'POST',
        headers,
        // Buffer evita que fetch agregue text/plain cuando queremos omitir el header.
        body: Buffer.from(tc.body),
        signal: AbortSignal.timeout(5000),
      });
      const body = await response.json();
      assert.equal(response.status, tc.status);
      assert.match(response.headers.get('content-type'), /^application\/json\b/);
      if (tc.expected) {
        assert.deepEqual(body, tc.expected);
      } else {
        assert.equal(typeof body.error, 'string');
        assert.ok(body.error.length > 0);
        assert.equal(body.stack, undefined);
      }
    });
  }
});
test('GET /health confirma que Express responde', async (t) => {
  const server = createApp({ serviceSecret }).listen(0, '127.0.0.1');
  t.after(() => new Promise((resolve, reject) => {
    server.close((error) => error ? reject(error) : resolve());
  }));
  await once(server, 'listening');
  const response = await fetch(`http://127.0.0.1:${server.address().port}/health`, {
    signal: AbortSignal.timeout(5000),
  });
  assert.equal(response.status, 200);
  assert.deepEqual(await response.json(), { status: 'ok' });
});

test('POST /statistics/qr mantiene autenticacion y agrega Q y R', async (t) => {
  const server = createApp({ serviceSecret }).listen(0, '127.0.0.1');
  t.after(() => new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve())));
  await once(server, 'listening');
  const url = `http://127.0.0.1:${server.address().port}/statistics/qr`;
  const validBody = JSON.stringify({q:[[1,0],[0,1]],r:[[2,3],[0,4]]});
  for (const [name, body, token, status] of [
    ['valido',validBody,serviceToken(),200],
    ['sin token',validBody,null,401],
    ['matriz faltante','{}',serviceToken(),400],
    ['JSON incorrecto','{',serviceToken(),400],
  ]) {
    await t.test(name, async () => {
      const headers = {'Content-Type':'application/json'};
      if (token) headers.Authorization = `Bearer ${token}`;
      const response = await fetch(url,{method:'POST',headers,body,signal:AbortSignal.timeout(5000)});
      const result = await response.json();
      assert.equal(response.status,status);
      if (status === 200) {
        assert.equal(result.sum,11); assert.equal(result.average,11/8);
        assert.equal(result.isDiagonal,true); assert.deepEqual(result.diagonalMatrices,['q']);
        assert.equal(result.perMatrix.r.isDiagonal,false);
      } else { assert.equal(typeof result.error,'string'); }
    });
  }
});
