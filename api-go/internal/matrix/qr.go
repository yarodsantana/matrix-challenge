package matrix

import (
	"errors"
	"fmt"
	"math"
)

// MaxQRDimension limita CPU, memoria y el tamano de la respuesta Q completa.
const MaxQRDimension = 100

var ErrNumericRange = errors.New("la factorizacion excede el rango numerico admitido")

// QRResult usa QR completa: Q es m x m y R es m x n, incluso si m < n.
// Esta convencion admite matrices altas, anchas y de rango deficiente.
type QRResult struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

// FactorQR calcula A=QR mediante reflexiones de Householder.
// Escala primero A para evitar desbordamientos al calcular normas.
// Nunca modifica la matriz de entrada ni supone que sus columnas son independientes.
func FactorQR(input [][]float64) (QRResult, error) {
	if err := Validate(input); err != nil {
		return QRResult{}, err
	}
	rows, cols := len(input), len(input[0])
	if rows > MaxQRDimension || cols > MaxQRDimension {
		return QRResult{}, fmt.Errorf("QR admite como maximo %d filas y %d columnas", MaxQRDimension, MaxQRDimension)
	}
	q := makeMatrix(rows, rows)
	r := makeMatrix(rows, cols)
	scale := 0.0
	for i, row := range input {
		q[i][i] = 1
		for _, value := range row {
			scale = math.Max(scale, math.Abs(value))
		}
	}
	if scale == 0 {
		return QRResult{Q: q, R: r}, nil
	}
	for i, row := range input {
		for j, value := range row {
			r[i][j] = value / scale
		}
	}
	steps := min(rows, cols)
	for k := 0; k < steps; k++ {
		norm := 0.0
		for i := k; i < rows; i++ {
			norm = math.Hypot(norm, r[i][k])
		}
		if norm == 0 {
			continue
		}
		v := make([]float64, rows-k)
		for i := k; i < rows; i++ {
			v[i-k] = r[i][k] / norm
		}
		// Elegir el signo contrario evita cancelar numeros casi iguales.
		v[0] += math.Copysign(1, v[0])
		vnorm := 0.0
		for _, value := range v {
			vnorm = math.Hypot(vnorm, value)
		}
		for i := range v {
			v[i] /= vnorm
		}
		// R <- H R, con H = I - 2 v v^T.
		for j := k; j < cols; j++ {
			dot := 0.0
			for i := k; i < rows; i++ {
				dot += v[i-k] * r[i][j]
			}
			for i := k; i < rows; i++ {
				r[i][j] -= 2 * v[i-k] * dot
			}
		}
		// Q <- Q H. Se acumulan las reflexiones sin construir H completa.
		for i := 0; i < rows; i++ {
			dot := 0.0
			for j := k; j < rows; j++ {
				dot += q[i][j] * v[j-k]
			}
			for j := k; j < rows; j++ {
				q[i][j] -= 2 * dot * v[j-k]
			}
		}
		for i := k + 1; i < rows; i++ {
			r[i][k] = 0
		}
	}
	for i := range r {
		for j := range r[i] {
			r[i][j] *= scale
			if math.IsInf(r[i][j], 0) || math.IsNaN(r[i][j]) {
				return QRResult{}, ErrNumericRange
			}
		}
	}
	return QRResult{Q: q, R: r}, nil
}

func makeMatrix(rows, cols int) [][]float64 {
	result := make([][]float64, rows)
	for i := range result {
		result[i] = make([]float64, cols)
	}
	return result
}

// Validate comparte las reglas estructurales entre rotacion y QR.
func Validate(input [][]float64) error {
	if len(input) == 0 {
		return errors.New("la matriz debe tener al menos una fila")
	}
	cols := len(input[0])
	if cols == 0 {
		return errors.New("la matriz debe tener al menos una columna")
	}
	for i, row := range input {
		if len(row) != cols {
			return fmt.Errorf("la fila %d tiene %d columnas; se esperaban %d", i, len(row), cols)
		}
		for _, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return errors.New("cada elemento debe ser un numero finito")
			}
		}
	}
	return nil
}
