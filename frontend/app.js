/* global jQuery */
'use strict';

jQuery(function ($) {
  const $form = $('#matrix-form');
  const $button = $('#submit');
  const $message = $('#message');
  const $result = $('#result');

  function showMessage(text, type) {
    $message
      .removeClass('d-none alert-danger alert-info')
      .addClass(`alert-${type}`)
      .text(text);
  }

  function readMatrix(operation) {
    let matrix;
    try {
      matrix = JSON.parse($('#matrix-input').val());
    } catch {
      throw new Error('La matriz no tiene un formato JSON válido.');
    }

    if (!Array.isArray(matrix) || !matrix.length ||
        !Array.isArray(matrix[0]) || !matrix[0].length) {
      throw new Error('La matriz debe tener al menos una fila y una columna.');
    }

    const columns = matrix[0].length;
    for (const row of matrix) {
      if (!Array.isArray(row) || row.length !== columns ||
          row.some((value) => !Number.isFinite(value))) {
        throw new Error('Todas las filas deben tener igual longitud y contener números finitos.');
      }
    }

    if (operation === 'qr' && (matrix.length > 100 || columns > 100)) {
      throw new Error('La factorización QR admite hasta 100 filas y 100 columnas.');
    }
    return matrix;
  }

  function appendMatrix(title, matrix) {
    const $table = $('<table>').addClass('table table-sm table-bordered matrix-table');
    const $body = $('<tbody>');

    for (const row of matrix) {
      const $row = $('<tr>');
      for (const value of row) {
        // text() evita interpretar los datos como HTML y conserva su precisión.
        $row.append($('<td>').text(String(value)));
      }
      $body.append($row);
    }

    $table.attr('aria-label', title).append($body);
    $('#matrices').append(
      $('<h3>').addClass('h6 mt-3').text(title),
      $('<div>').addClass('table-responsive').append($table)
    );
  }

  function renderResult(response, operation) {
    const stats = response.statistics;
    const $table = $('<table>').addClass('table table-sm');
    const $body = $('<tbody>');

    for (const [label, value] of [
      ['Máximo', stats.max],
      ['Mínimo', stats.min],
      ['Promedio', stats.average],
      ['Suma total', stats.sum]
    ]) {
      $body.append($('<tr>').append(
        $('<th>').attr('scope', 'row').text(label),
        $('<td>').text(String(value))
      ));
    }

    $('#statistics').empty().append($table.append($body));
    $('#matrices').empty();

    if (operation === 'qr') {
      appendMatrix('Matriz rotada', response.rotated);
      appendMatrix('Q', response.q);
      appendMatrix('R', response.r);
      const names = stats.diagonalMatrices.map((name) => name.toUpperCase());
      $('#diagonal-result').text(names.length
        ? `Matrices diagonales: ${names.join(', ')}.`
        : 'Ninguna matriz es diagonal.');
    } else {
      appendMatrix('Matriz rotada', response.matrix);
      $('#diagonal-result').text(stats.isDiagonal
        ? 'La matriz es diagonal.'
        : 'La matriz no es diagonal.');
    }

    $('#raw-result').text(JSON.stringify(response, null, 2));
    $('#empty-result').addClass('d-none');
    $result.removeClass('d-none');
  }

  $form.on('submit', function (event) {
    event.preventDefault();
    if ($button.prop('disabled')) return;

    $result.addClass('d-none');
    $('#empty-result').removeClass('d-none');
    const operation = $('#operation').val();
    const token = $('#token').val().trim();
    let matrix;

    try {
      if (!token) throw new Error('Ingresa el token.');
      matrix = readMatrix(operation);
    } catch (error) {
      showMessage(error.message, 'danger');
      return;
    }

    $button.prop('disabled', true).text('Calculando…');
    showMessage('Procesando…', 'info');

    $.ajax({
      url: `/${operation}`,
      method: 'POST',
      contentType: 'application/json',
      dataType: 'json',
      headers: { Authorization: `Bearer ${token}` },
      data: JSON.stringify(matrix),
      timeout: 120000
    })
      .done(function (response) {
        renderResult(response, operation);
        $message.addClass('d-none');
      })
      .fail(function (xhr, status) {
        let text = xhr.responseJSON?.error || 'No se pudo completar la consulta.';
        if (xhr.status === 401) text = 'El token es inválido o está vencido.';
        if (status === 'timeout') text = 'La consulta tardó demasiado. Intenta nuevamente.';
        showMessage(text, 'danger');
      })
      .always(function () {
        $button.prop('disabled', false).text('Calcular');
      });
  });
});