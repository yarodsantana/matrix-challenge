'use strict';

const $ = (id) => document.getElementById(id);
const examples = {
  classic: [[12,-51,4],[6,167,-68],[-4,24,-41]],
  tall: [[1,2],[3,4],[5,6]],
  wide: [[1,2,3],[4,5,6]],
  diagonal: [[2,0],[0,3]],
  zero: [[0,0],[0,0]],
};
let lastResult = null;
let busy = false;

function setStatus(message, type = '') {
  $('status').textContent = message;
  $('status').className = `status ${type}`;
}

function readMatrix() {
  let matrix;
  try { matrix = JSON.parse($('matrix-input').value); }
  catch { throw new Error('El JSON no es válido. Usa arrays, comas y números sin comillas.'); }
  if (!Array.isArray(matrix) || matrix.length === 0 || !Array.isArray(matrix[0]) || matrix[0].length === 0) {
    throw new Error('Introduce al menos una fila y una columna.');
  }
  const columns = matrix[0].length;
  for (const row of matrix) {
    if (!Array.isArray(row) || row.length !== columns || row.some((value) => !Number.isFinite(value))) {
      throw new Error('Todas las filas deben tener igual longitud y contener solo números finitos.');
    }
  }
  if ($('operation').value === 'qr' && (matrix.length > 100 || columns > 100)) {
    throw new Error('Para QR, el límite es de 100 filas y 100 columnas.');
  }
  return matrix;
}

function updateDimensions() {
  try { const matrix = readMatrix(); $('dimensions').textContent = `${matrix.length} × ${matrix[0].length}`; }
  catch { $('dimensions').textContent = 'Revisar JSON'; }
}

function formatNumber(number) {
  if (number === 0) return '0';
  if (Math.abs(number) < 0.00001 || Math.abs(number) >= 1e8) return number.toExponential(4);
  return Number(number.toPrecision(7)).toLocaleString('es-PE', { maximumFractionDigits: 7 });
}

function matrixCard(name, matrix) {
  const card = document.createElement('article');
  card.className = 'matrix-card';
  const heading = document.createElement('div');
  heading.className = 'matrix-title';
  const title = document.createElement('h3');
  title.textContent = name;
  const dimensions = document.createElement('span');
  dimensions.textContent = `${matrix.length} × ${matrix[0].length}`;
  heading.append(title, dimensions);
  const scroll = document.createElement('div');
  scroll.className = 'matrix-scroll';
  const table = document.createElement('table');
  table.className = 'matrix-table';
  table.setAttribute('aria-label', `${name}, ${dimensions.textContent}`);
  const body = document.createElement('tbody');
  for (const row of matrix) {
    const tr = document.createElement('tr');
    for (const value of row) {
      const td = document.createElement('td');
      td.textContent = formatNumber(value);
      td.title = String(value);
      if (value === 0) td.className = 'zero';
      tr.append(td);
    }
    body.append(tr);
  }
  table.append(body);
  scroll.append(table);
  card.append(heading, scroll);
  return card;
}

// Comprobacion independiente de reconstruccion; no sustituye las estadisticas de Node.
function reconstructionError(original, q, r) {
  let maxError = 0;
  let scale = 0;
  original.forEach((row) => row.forEach((v) => { scale = Math.max(scale, Math.abs(v)); }));
  if (scale === 0) scale = 1;
  for (let i = 0; i < original.length; i += 1) {
    for (let j = 0; j < original[0].length; j += 1) {
      let value = 0;
      for (let k = 0; k < q[0].length; k += 1) value += q[i][k] * (r[k][j] / scale);
      maxError = Math.max(maxError, Math.abs(value - original[i][j] / scale));
    }
  }
  return maxError;
}

function showResult(result, input, elapsed, mode) {
  lastResult = result;
  $('empty-state').hidden = true;
  $('results').hidden = false;
  $('elapsed').textContent = `${elapsed.toFixed(0)} ms · HTTP 200`;
  $('stats').replaceChildren();
  const stats = result.statistics;
  for (const [label, value] of [['Máximo',stats.max],['Mínimo',stats.min],['Promedio',stats.average],['Suma total',stats.sum],['Valores',stats.valueCount ?? input.length * input[0].length]]) {
    const card = document.createElement('div'); card.className = 'stat';
    const name = document.createElement('div'); name.className = 'stat-label'; name.textContent = label;
    const number = document.createElement('div'); number.className = 'stat-value'; number.textContent = formatNumber(value); number.title = String(value);
    card.append(name, number); $('stats').append(card);
  }
  $('matrices').replaceChildren();
  if (mode === 'qr') {
    $('diagonal-result').textContent = stats.isDiagonal ? `Matrices diagonales: ${stats.diagonalMatrices.map((name) => name.toUpperCase()).join(', ')}` : 'Ninguna de Q y R es diagonal';
    const error = reconstructionError(result.original, result.q, result.r);
    $('verification').textContent = `Error relativo de reconstrucción: ${error.toExponential(2)}`;
    for (const [name, matrix] of [['Original · A',result.original],['Rotación · 90° horario',result.rotated],['Q · ortogonal',result.q],['R · triangular superior',result.r]]) {
      $('matrices').append(matrixCard(name,matrix));
    }
    $('output-title').textContent = 'Factorización y análisis de Q + R';
  } else {
    $('diagonal-result').textContent = stats.isDiagonal ? 'La matriz rotada es diagonal' : 'La matriz rotada no es diagonal';
    $('verification').textContent = 'Contrato original conservado';
    $('matrices').append(matrixCard('Original · A',input),matrixCard('Rotación · 90° horario',result.matrix));
    $('output-title').textContent = 'Rotación y estadísticas';
  }
  $('raw-result').textContent = JSON.stringify(result,null,2);
}

$('matrix-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  if (busy) return;
  let matrix;
  const token = $('token').value.trim();
  try { matrix = readMatrix(); if (!token) throw new Error('Pega primero un JWT de cliente válido para Go.'); }
  catch (error) { setStatus(error.message,'error'); return; }
  busy = true; $('submit').disabled = true; $('submit').textContent = 'Procesando…';
  setStatus('Go está procesando la matriz y consultando Node. En Render gratuito, el primer arranque puede demorar.','loading');
  const mode = $('operation').value;
  const start = performance.now();
  try {
    const response = await fetch(`/${mode}`, {
      method: 'POST', headers: { 'Content-Type':'application/json', Authorization:`Bearer ${token}` },
      body: JSON.stringify(matrix), signal: AbortSignal.timeout(120000),
    });
    const result = await response.json();
    if (!response.ok) throw new Error(response.status === 401 ? 'HTTP 401: token inválido o vencido. Genera uno nuevo y vuelve a pegarlo.' : `HTTP ${response.status}: ${result.error ?? 'No se pudo completar la operación'}`);
    showResult(result,matrix,performance.now()-start,mode);
    setStatus('Consulta completada: transformación en Go y estadísticas recibidas desde Node.');
  } catch (error) {
    $('results').hidden = true; $('empty-state').hidden = false; $('elapsed').textContent = 'Consulta fallida';
    setStatus(error.name === 'TimeoutError' ? 'La consulta superó 120 segundos. Espera a que ambos servicios estén activos y reintenta.' : error.message,'error');
  } finally {
    busy = false; $('submit').disabled = false; $('submit').textContent = 'Calcular y analizar ↗';
  }
});

for (const button of document.querySelectorAll('[data-example]')) {
  button.addEventListener('click', () => { $('matrix-input').value = JSON.stringify(examples[button.dataset.example],null,2); updateDimensions(); });
}
$('matrix-input').addEventListener('input',updateDimensions);
$('operation').addEventListener('change',updateDimensions);
$('matrix-input').addEventListener('keydown',(event) => { if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') $('matrix-form').requestSubmit(); });
$('toggle-token').addEventListener('click',() => {
  const show = $('token').type === 'password'; $('token').type = show ? 'text':'password';
  $('toggle-token').textContent = show ? 'Ocultar':'Mostrar'; $('toggle-token').setAttribute('aria-pressed',String(show));
});
$('copy-result').addEventListener('click',async () => {
  if (!lastResult) return;
  try { await navigator.clipboard.writeText(JSON.stringify(lastResult,null,2)); setStatus('Resultado JSON copiado.'); }
  catch { setStatus('El navegador no permitió copiar. Puedes seleccionar el JSON manualmente.','error'); }
});
$('download-result').addEventListener('click',() => {
  if (!lastResult) return;
  const url = URL.createObjectURL(new Blob([JSON.stringify(lastResult,null,2)],{type:'application/json'}));
  const link = document.createElement('a'); link.href = url; link.download = 'matrix-result.json'; link.click();
  setTimeout(() => URL.revokeObjectURL(url),1000);
});
fetch('/health',{signal:AbortSignal.timeout(5000)}).then((response) => {
  $('health').textContent = response.ok ? '● Go activo':'Go no disponible';
  $('health').classList.toggle('ok',response.ok);
}).catch(() => { $('health').textContent = 'Sin conexión'; });
updateDimensions();