package matrix

import (
	"errors"
	"math"
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestQRProperties(t *testing.T) {
	cases := []struct {
		name  string
		input [][]float64
	}{
		{"clasica", [][]float64{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}}},
		{"alta", [][]float64{{1, 2}, {3, 4}, {5, 6}}},
		{"ancha", [][]float64{{1, 2, 3}, {4, 5, 6}}},
		{"rango deficiente", [][]float64{{1, 2}, {2, 4}, {3, 6}}},
		{"nula", [][]float64{{0, 0}, {0, 0}, {0, 0}}},
		{"diagonal", [][]float64{{2, 0}, {0, -3}}},
		{"una fila", [][]float64{{-1, 0, 2}}},
		{"una columna", [][]float64{{1}, {2}, {3}}},
		{"un elemento", [][]float64{{-7}}},
		{"columna cero", [][]float64{{0, 2}, {0, 3}, {0, 4}}},
		{"muy pequena", [][]float64{{1e-300, 2e-300}, {3e-300, 4e-300}}},
		{"muy grande", [][]float64{{1e150, 2e150}, {3e150, 4e150}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkQR(t, tc.input) })
	}
}

func TestQRDeterministicRandomMatrices(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for m := 1; m <= 8; m++ {
		for n := 1; n <= 8; n++ {
			a := makeMatrix(m, n)
			for i := range a {
				for j := range a[i] {
					a[i][j] = rng.Float64()*20 - 10
				}
			}
			checkQR(t, a)
		}
	}
}

func checkQR(t *testing.T, a [][]float64) {
	t.Helper()
	before := makeMatrix(len(a), len(a[0]))
	for i := range a {
		copy(before[i], a[i])
	}
	qr, err := FactorQR(a)
	if err != nil {
		t.Fatal(err)
	}
	m, n := len(a), len(a[0])
	if len(qr.Q) != m || len(qr.Q[0]) != m || len(qr.R) != m || len(qr.R[0]) != n {
		t.Fatal("dimensiones QR incorrectas")
	}
	if !reflect.DeepEqual(a, before) {
		t.Fatal("se modifico la entrada")
	}
	scale := 0.0
	for _, row := range a {
		for _, v := range row {
			scale = math.Max(scale, math.Abs(v))
		}
	}
	if scale == 0 {
		scale = 1
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			reconstructed := 0.0
			for k := 0; k < m; k++ {
				reconstructed += qr.Q[i][k] * (qr.R[k][j] / scale)
			}
			if math.Abs(reconstructed-a[i][j]/scale) > 1e-10 {
				t.Fatalf("QR no reconstruye A en (%d,%d)", i, j)
			}
			if i > j && qr.R[i][j] != 0 {
				t.Fatal("R no es triangular superior")
			}
		}
		for j := 0; j < m; j++ {
			dot := 0.0
			for k := 0; k < m; k++ {
				dot += qr.Q[k][i] * qr.Q[k][j]
			}
			want := 0.0
			if i == j {
				want = 1
			}
			if math.Abs(dot-want) > 1e-10 {
				t.Fatal("Q no es ortogonal")
			}
		}
	}
	qr.R[0][0] = 99
	if !reflect.DeepEqual(a, before) {
		t.Fatal("salida y entrada comparten memoria")
	}
}

func TestQRRejectsInvalidInput(t *testing.T) {
	for _, a := range [][][]float64{nil, {{}}, {{1}, {1, 2}}, {{math.NaN()}}, {{math.Inf(1)}}, makeMatrix(101, 1)} {
		if _, err := FactorQR(a); err == nil {
			t.Fatal("se acepto una matriz invalida")
		}
	}
	_, err := FactorQR([][]float64{{math.MaxFloat64}, {math.MaxFloat64}})
	if !errors.Is(err, ErrNumericRange) {
		t.Fatalf("se esperaba error de rango: %v", err)
	}
}
