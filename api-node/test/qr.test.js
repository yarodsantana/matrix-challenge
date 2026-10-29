import test from 'node:test';
import assert from 'node:assert/strict';
import { calculateStatistics, calculateQRStatistics } from '../src/statistics.js';

test('QR agrega valores de las dos matrices y verifica si alguna es diagonal', () => {
  const result = calculateQRStatistics({ q: [[1,0],[0,1]], r: [[2,3],[0,4]] });
  assert.equal(result.max, 4);
  assert.equal(result.min, 0);
  assert.equal(result.sum, 11);
  assert.equal(result.average, 11/8);
  assert.equal(result.valueCount, 8);
  assert.equal(result.matrixCount, 2);
  assert.deepEqual(result.diagonalMatrices, ['q']);
  assert.equal(result.isDiagonal, true);
  assert.equal(result.perMatrix.r.isDiagonal, false);
});

test('QR usa promedio ponderado por cantidad de elementos', () => {
  const result = calculateQRStatistics({ q: [[1,0,0],[0,1,0],[0,0,1]], r: [[1,2],[3,4],[5,6]] });
  assert.equal(result.valueCount, 15);
  assert.equal(result.sum, 24);
  assert.equal(result.average, 1.6);
});

test('QR aplica tolerancia relativa sin cambiar la comparacion exacta legacy', () => {
  const q = [[1,5e-12],[0,1]];
  assert.equal(calculateStatistics(q).isDiagonal, false);
  assert.equal(calculateQRStatistics({ q, r: [[2,0],[0,3]] }).isDiagonal, true);
  assert.deepEqual(calculateQRStatistics({ q, r: [[2,0],[0,3]] }).diagonalMatrices, ['q','r']);
  assert.equal(calculateQRStatistics({q:[[1,0.01],[0,1]],r:[[2,3],[0,4]]}).isDiagonal, false);
});

test('QR rechaza matrices ausentes o incompatibles', () => {
  for (const input of [null, [], {}, {q:[[1]]}, {q:[[1,2]],r:[[1]]}, {q:[[1]],r:[[1],[2]]}, {q:[[1]],r:[[null]]}]) {
    assert.throws(() => calculateQRStatistics(input), TypeError);
  }
});

test('QR rechaza desbordamiento agregado aun si cada suma es finita', () => {
  assert.throws(() => calculateQRStatistics({q:[[Number.MAX_VALUE]],r:[[Number.MAX_VALUE]]}), RangeError);
});