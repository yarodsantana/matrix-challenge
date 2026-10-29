// La entrada debe ser una matriz rectangular no vacia de numeros finitos.
export function calculateStatistics(matrix, { tolerance = 0 } = {}) {
  if (!Array.isArray(matrix) || matrix.length === 0) {
    throw new TypeError('La matriz debe tener al menos una fila');
  }

  if (!Array.isArray(matrix[0]) || matrix[0].length === 0) {
    throw new TypeError('La matriz debe tener al menos una columna');
  }

  const rows = matrix.length;
  const columns = matrix[0].length;
  let max = -Infinity;
  let min = Infinity;
  let sum = 0;
  let offDiagonalMax = 0;

  for (let r = 0; r < rows; r += 1) {
    const row = matrix[r];
    if (!Array.isArray(row) || row.length !== columns) {
      throw new TypeError('Todas las filas deben tener la misma longitud');
    }

    for (let c = 0; c < columns; c += 1) {
      const value = row[c];
      if (!Number.isFinite(value)) {
        throw new TypeError('Cada elemento debe ser un numero finito');
      }

      max = Math.max(max, value);
      min = Math.min(min, value);
      sum += value;

      if (!Number.isFinite(sum)) {
        throw new RangeError('La suma excede el rango numerico admitido');
      }

      if (r !== c) offDiagonalMax = Math.max(offDiagonalMax, Math.abs(value));
    }
  }

  return {
    max,
    min,
    average: sum / (rows * columns),
    sum,
    isDiagonal: rows === columns && offDiagonalMax <= tolerance * Math.max(1, Math.abs(max), Math.abs(min)),
  };
}
// Para QR, el contrato agrega todos los valores de Q y R, sin promediar promedios.
// La tolerancia relativa por matriz evita tratar el ruido numerico como no-cero.
export const QR_DIAGONAL_TOLERANCE = 1e-10;

export function calculateQRStatistics(input) {
  if (input === null || typeof input !== 'object' || Array.isArray(input)) {
    throw new TypeError('Se espera un objeto con matrices q y r');
  }
  const options = { tolerance: QR_DIAGONAL_TOLERANCE };
  const q = calculateStatistics(input.q, options);
  const r = calculateStatistics(input.r, options);
  const rows = input.q.length;
  if (input.q[0].length !== rows || input.r.length !== rows) {
    throw new TypeError('Q debe ser cuadrada y R debe tener tantas filas como Q');
  }
  const valueCount = rows * rows + input.r.length * input.r[0].length;
  const sum = q.sum + r.sum;
  if (!Number.isFinite(sum)) throw new RangeError('La suma excede el rango numerico admitido');
  const perMatrix = { q, r };
  const diagonalMatrices = ['q', 'r'].filter((name) => perMatrix[name].isDiagonal);
  return {
    max: Math.max(q.max, r.max),
    min: Math.min(q.min, r.min),
    average: sum / valueCount,
    sum,
    isDiagonal: diagonalMatrices.length > 0,
    matrixCount: 2,
    valueCount,
    diagonalMatrices,
    diagonalTolerance: QR_DIAGONAL_TOLERANCE,
    perMatrix,
  };
}