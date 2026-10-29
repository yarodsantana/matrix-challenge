package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/matrix-api/internal/statistics"
	"github.com/gofiber/fiber/v3"
)

const scalarQRStats = `{"max":-1,"min":-2,"average":-1.5,"sum":-3,"isDiagonal":true,"matrixCount":2,"valueCount":2,"diagonalMatrices":["q","r"],"diagonalTolerance":1e-10,"perMatrix":{"q":{"max":-1,"min":-1,"average":-1,"sum":-1,"isDiagonal":true},"r":{"max":-2,"min":-2,"average":-2,"sum":-2,"isDiagonal":true}}}`

func TestQRHTTP(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/statistics/qr" {
			t.Errorf("ruta incorrecta: %s", r.URL.Path)
		}
		var input struct{ Q, R [][]float64 }
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		if len(input.Q) != 1 || len(input.R) != 1 {
			t.Error("Node debe recibir ambas matrices")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(scalarQRStats))
	}))
	defer upstream.Close()
	app := newAppWithClient(&statistics.Client{URL: upstream.URL + "/statistics", HTTPClient: upstream.Client()})
	req := httptest.NewRequest("POST", "/qr", strings.NewReader(`[[2]]`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken(t))
	res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("estado: %d", res.StatusCode)
	}
	var body struct {
		Q, R, Original, Rotated [][]float64
		Statistics              statistics.QRResult
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Q[0][0] != -1 || body.R[0][0] != -2 || body.Original[0][0] != 2 || body.Rotated[0][0] != 2 || body.Statistics.Sum != -3 {
		t.Fatalf("resultado inesperado: %+v", body)
	}
}

func TestQRRejectsInputBeforeCallingNode(t *testing.T) {
	app := newApp()
	cases := []struct {
		body string
		want int
	}{
		{`[]`, 400}, {`[[null]]`, 400}, {`[[1,2],[3]]`, 400},
		{`{"matrix":[[1]]}`, 400}, {`[[1.7976931348623157e308],[1.7976931348623157e308]]`, 422},
		{"[" + strings.Repeat("[1],", 100) + "[1]]", 400},
	}
	for _, tc := range cases {
		req := httptest.NewRequest("POST", "/qr", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testToken(t))
		res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != tc.want {
			t.Fatalf("estado: %d; esperado: %d", res.StatusCode, tc.want)
		}
	}
	req := httptest.NewRequest("POST", "/qr", strings.NewReader(`[[1]]`))
	res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 401 {
		t.Fatal("QR debe exigir JWT")
	}
}

func TestFrontendAssets(t *testing.T) {
	app := newApp()
	for _, path := range []string{"/", "/app.js", "/styles.css"} {
		req := httptest.NewRequest("GET", path, nil)
		res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if err != nil || res.StatusCode != 200 || len(body) == 0 {
			t.Fatalf("asset %s no disponible: %d %v", path, res.StatusCode, err)
		}
		if res.Header.Get("Content-Security-Policy") == "" {
			t.Fatal("falta CSP")
		}
	}
}
