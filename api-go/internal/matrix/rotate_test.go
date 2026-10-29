package matrix

import (
	"reflect"
	"testing"
)

func TestRotateClockwiseRectangular(t *testing.T) {
	input := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	want := [][]float64{
		{4, 1},
		{5, 2},
		{6, 3},
	}

	got, err := RotateClockwise(input)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resultado: %v; esperado: %v", got, want)
	}
}

func TestRotateClockwiseRejectsIrregularMatrix(t *testing.T) {
	input := [][]float64{
		{1, 2},
		{3},
	}

	_, err := RotateClockwise(input)
	if err == nil {
		t.Fatal("se esperaba un error por filas de distinta longitud")
	}
}

func TestRotateClockwiseEdgeCases(t *testing.T) {
	cases := []struct {
		name  string
		input [][]float64
		want  [][]float64
	}{
		{"negativos y decimales", [][]float64{{-1.5, 0}, {2.5, -3}}, [][]float64{{2.5, -1.5}, {-3, 0}}},
		{"una fila", [][]float64{{1, 2, 3}}, [][]float64{{1}, {2}, {3}}},
		{"una columna", [][]float64{{1}, {2}, {3}}, [][]float64{{3, 2, 1}}},
		{"un elemento", [][]float64{{7}}, [][]float64{{7}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RotateClockwise(tc.input)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("resultado: %v; esperado: %v", got, tc.want)
			}
		})
	}
}

func TestRotateClockwiseRejectsEmptyInput(t *testing.T) {
	cases := []struct {
		name  string
		input [][]float64
	}{
		{"nil", nil},
		{"sin filas", [][]float64{}},
		{"sin columnas", [][]float64{{}}},
		{"fila nil", [][]float64{nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RotateClockwise(tc.input)
			if err == nil {
				t.Fatal("se esperaba un error")
			}
			if got != nil {
				t.Fatalf("se esperaba resultado nil; recibido: %v", got)
			}
		})
	}
}

func TestRotateClockwisePreservesInput(t *testing.T) {
	input := [][]float64{{1, 2, 3}, {4, 5, 6}}
	original := [][]float64{{1, 2, 3}, {4, 5, 6}}
	got, err := RotateClockwise(input)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatalf("la rotacion modifico la entrada: %v", input)
	}
	// Modificar la salida tampoco debe modificar la entrada.
	got[0][0] = 999
	if !reflect.DeepEqual(input, original) {
		t.Fatalf("la salida comparte memoria con la entrada: %v", input)
	}
}
