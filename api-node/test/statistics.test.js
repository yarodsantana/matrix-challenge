import test from 'node:test';
import assert from 'node:assert/strict';
import { calculateStatistics } from '../src/statistics.js';

const validCases = [
  {
    name: 'matriz rectangular procesada por Go',
    matrix: [[4, 1], [5, 2], [6, 3]],
    expected: { max: 6, min: 1, average: 3.5, sum: 21, isDiagonal: false },
  },
  {
    name: 'todos negativos',
    matrix: [[-8, -2], [-4, -6]],
    expected: { max: -2, min: -8, average: -5, sum: -20, isDiagonal: false },
  },
  {
    name: 'decimales',
    matrix: [[-1.5, 0], [2.5, -3]],
    expected: { max: 2.5, min: -3, average: -0.5, sum: -2, isDiagonal: false },
  },
  {
    name: 'diagonal con elementos negativos',
    matrix: [[-2, 0], [0, 4]],
    expected: { max: 4, min: -2, average: 0.5, sum: 2, isDiagonal: true },
  },
  {
    name: 'matriz cuadrada nula',
    matrix: [[0, 0], [0, 0]],
    expected: { max: 0, min: 0, average: 0, sum: 0, isDiagonal: true },
  },
  {
    name: 'rectangular nula no es diagonal',
    matrix: [[0, 0, 0], [0, 0, 0]],
    expected: { max: 0, min: 0, average: 0, sum: 0, isDiagonal: false },
  },
  {
    name: 'un elemento es diagonal',
    matrix: [[7]],
    expected: { max: 7, min: 7, average: 7, sum: 7, isDiagonal: true },
  },
  {
    name: 'un valor fuera de la diagonal invalida la condicion',
    matrix: [[1, 0.001], [0, 2]],
    expected: { max: 2, min: 0, average: 3.001 / 4, sum: 3.001, isDiagonal: false },
  },
];

for (const { name, matrix, expected } of validCases) {
  test(name, () => {
    assert.deepEqual(calculateStatistics(matrix), expected);
  });
}

const invalidCases = [
  ['null', null],
  ['objeto', {}],
  ['sin filas', []],
  ['sin columnas', [[]]],
  ['fila null', [null]],
  ['fila posterior null', [[1], null]],
  ['filas irregulares', [[1, 2], [3]]],
  ['elemento null', [[null]]],
  ['texto numerico', [['2']]],
  ['booleano', [[true]]],
  ['NaN', [[NaN]]],
  ['infinito', [[Infinity]]],
  ['elemento ausente', [[,]]],
];

for (const [name, matrix] of invalidCases) {
  test(`rechaza ${name}`, () => {
    assert.throws(() => calculateStatistics(matrix), TypeError);
  });
}

test('rechaza desbordamiento de la suma', () => {
  assert.throws(
    () => calculateStatistics([[Number.MAX_VALUE, Number.MAX_VALUE]]),
    RangeError,
  );
});

test('conserva la entrada', () => {
  const matrix = Object.freeze([
    Object.freeze([1, 2]),
    Object.freeze([3, 4]),
  ]);
  calculateStatistics(matrix);
  assert.deepEqual(matrix, [[1, 2], [3, 4]]);
});

test('promedio decimal con tolerancia de punto flotante', () => {
  const result = calculateStatistics([[0.1, 0.2]]);
  assert.ok(Math.abs(result.average - 0.15) < 1e-12);
});