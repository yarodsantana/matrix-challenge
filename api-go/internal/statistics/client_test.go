package statistics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const validQR = `{"max":-1,"min":-2,"average":-1.5,"sum":-3,"isDiagonal":true,"matrixCount":2,"valueCount":2,"diagonalMatrices":["q","r"],"diagonalTolerance":1e-10,"perMatrix":{"q":{"max":-1,"min":-1,"average":-1,"sum":-1,"isDiagonal":true},"r":{"max":-2,"min":-2,"average":-2,"sum":-2,"isDiagonal":true}}}`

func TestQRResponseContract(t *testing.T) {
	cases := []struct {
		name, body string
		valid      bool
	}{
		{"valida", validQR, true},
		{"falta metadata", `{"max":1,"min":1,"average":1,"sum":2,"isDiagonal":true}`, false},
		{"conteo incorrecto", strings.Replace(validQR, `"valueCount":2`, `"valueCount":99`, 1), false},
		{"diagonal inconsistente", strings.Replace(validQR, `"diagonalMatrices":["q","r"]`, `"diagonalMatrices":[]`, 1), false},
		{"diagonal duplicada", strings.Replace(validQR, `"diagonalMatrices":["q","r"]`, `"diagonalMatrices":["q","q"]`, 1), false},
		{"matriz ausente", strings.Replace(validQR, `"perMatrix":{"q":`, `"perMatrix":{"other":`, 1), false},
		{"campo null", strings.Replace(validQR, `"max":-1`, `"max":null`, 1), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/statistics/qr" || r.Header.Get("Authorization") != "Bearer test-service-token" {
					t.Error("contrato de solicitud incorrecto")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer upstream.Close()
			client := Client{URL: upstream.URL + "/statistics", HTTPClient: upstream.Client(), Token: func() (string, error) { return "test-service-token", nil }}
			result, err := client.CalculateQR(context.Background(), [][]float64{{-1}}, [][]float64{{-2}})
			if (err == nil) != tc.valid {
				t.Fatalf("validacion inesperada: %v", err)
			}
			if tc.valid && result.Sum != -3 {
				t.Fatal("resultado incorrecto")
			}
		})
	}
}
