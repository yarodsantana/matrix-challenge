package main

import (
	"encoding/json"
	"example.com/matrix-api/internal/statistics"
	"github.com/gofiber/fiber/v3"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRotateHTTPValidMatrix(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Error("solicitud a Node debe ser POST JSON")
			http.Error(w, "contrato invalido", 400)
			return
		}
		var got [][]float64
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
			http.Error(w, "JSON invalido", 400)
			return
		}
		if !reflect.DeepEqual(got, [][]float64{{4, 1}, {5, 2}, {6, 3}}) {
			t.Errorf("matriz enviada sin rotacion correcta: %v", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"max":6,"min":1,"average":3.5,"sum":21,"isDiagonal":false}`))
	}))
	defer upstream.Close()
	t.Setenv("STATISTICS_URL", upstream.URL)
	app := newApp()

	req := httptest.NewRequest(
		http.MethodPost,
		"/rotate",
		strings.NewReader(`[[1,2,3],[4,5,6]]`),
	)
	req.Header.Set("Authorization", "Bearer "+testToken(t))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatalf("no se pudo ejecutar la solicitud: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("estado: %d; esperado: 200", resp.StatusCode)
	}

	var body struct {
		Matrix     [][]float64       `json:"matrix"`
		Statistics statistics.Result `json:"statistics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("respuesta JSON inválida: %v", err)
	}

	if body.Statistics != (statistics.Result{Max: 6, Min: 1, Average: 3.5, Sum: 21, IsDiagonal: false}) {
		t.Fatalf("estadisticas inesperadas: %+v", body.Statistics)
	}
	want := [][]float64{
		{4, 1},
		{5, 2},
		{6, 3},
	}

	if !reflect.DeepEqual(body.Matrix, want) {
		t.Fatalf("matriz: %v; esperada: %v", body.Matrix, want)
	}
}

func TestRotateHTTPRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		contentType string
		wantStatus  int
	}{
		{
			name:        "matriz vacia",
			body:        `[]`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "fila vacia",
			body:        `[[]]`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "filas irregulares",
			body:        `[[1,2],[3]]`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "elemento null",
			body:        `[[1,null]]`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "texto en lugar de numero",
			body:        `[[1,"dos"]]`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "JSON incompleto",
			body:        `[[1,2]`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "objeto en lugar de matriz",
			body:        `{"matrix":[[1,2]]}`,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "tipo de contenido incorrecto",
			body:        `[[1,2]]`,
			contentType: "text/plain",
			wantStatus:  http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp()

			req := httptest.NewRequest(
				http.MethodPost,
				"/rotate",
				strings.NewReader(tc.body),
			)
			req.Header.Set("Authorization", "Bearer "+testToken(t))
			req.Header.Set("Content-Type", tc.contentType)

			res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
			if err != nil {
				t.Fatalf("no se pudo procesar la solicitud: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Fatalf(
					"estado recibido: %d; esperado: %d",
					res.StatusCode,
					tc.wantStatus,
				)
			}

			var body struct {
				Error string `json:"error"`
			}

			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatalf("se esperaba un error en formato JSON: %v", err)
			}

			if body.Error == "" {
				t.Fatal("se esperaba un mensaje de error")
			}
		})
	}
}

func TestRotateHTTPUpstreamFailures(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		delay  time.Duration
		want   int
	}{
		{"error de Node", 500, `{"error":"interno"}`, 0, 502},
		{"JSON invalido", 200, `oops`, 0, 502},
		{"campos ausentes", 200, `{}`, 0, 502},
		{"respuesta excesiva", 200, strings.Repeat(" ", 64*1024+1), 0, 502},
		{"tipo incorrecto", 200, `{"max":"6","min":1,"average":3.5,"sum":21,"isDiagonal":false}`, 0, 502},
		{"JSON adicional", 200, `{"max":6,"min":1,"average":3.5,"sum":21,"isDiagonal":false}{}`, 0, 502},
		{"campo null", 200, `{"max":null,"min":0,"average":0,"sum":0,"isDiagonal":true}`, 0, 502},
		{"desbordamiento", 422, `{"error":"fuera de rango"}`, 0, 422},
		{"timeout", 200, `{}`, 100 * time.Millisecond, 504},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tc.delay)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer upstream.Close()
			timeout := 5 * time.Second
			if tc.delay > 0 {
				timeout = 20 * time.Millisecond
			}
			app := newAppWithClient(&statistics.Client{URL: upstream.URL, HTTPClient: &http.Client{Timeout: timeout}})
			req := httptest.NewRequest(http.MethodPost, "/rotate", strings.NewReader(`[[1]]`))
			req.Header.Set("Authorization", "Bearer "+testToken(t))
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.want {
				t.Fatalf("estado: %d; esperado: %d", res.StatusCode, tc.want)
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Error == "" {
				t.Fatal("falta error JSON")
			}
		})
	}
}

func TestRotateHTTPInvalidInputDoesNotCallNode(t *testing.T) {
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) }))
	defer upstream.Close()
	app := newAppWithClient(&statistics.Client{URL: upstream.URL, HTTPClient: upstream.Client()})
	req := httptest.NewRequest(http.MethodPost, "/rotate", strings.NewReader(`[[1,2],[3]]`))
	req.Header.Set("Authorization", "Bearer "+testToken(t))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 400 || calls != 0 {
		t.Fatalf("estado: %d; llamadas a Node: %d", res.StatusCode, calls)
	}
}

func TestRotateHTTPNodeUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := upstream.URL
	upstream.Close()
	app := newAppWithClient(&statistics.Client{URL: url, HTTPClient: &http.Client{Timeout: time.Second}})
	req := httptest.NewRequest(http.MethodPost, "/rotate", strings.NewReader(`[[1]]`))
	req.Header.Set("Authorization", "Bearer "+testToken(t))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req, fiber.TestConfig{Timeout: 2 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 502 {
		t.Fatalf("estado recibido: %d", res.StatusCode)
	}
}

func TestHealth(t *testing.T) {
	app := newApp()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 || body.Status != "ok" {
		t.Fatalf("salud inesperada: %d %+v", res.StatusCode, body)
	}
}

func TestRotateRequiresJWTBeforeCallingNode(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("una solicitud sin autorizacion no debe llegar a Node")
		w.WriteHeader(500)
	}))
	defer upstream.Close()
	app := newAppWithClient(&statistics.Client{URL: upstream.URL, HTTPClient: upstream.Client()})
	for _, header := range []string{"", "Bearer invalid", "Basic abc"} {
		req := httptest.NewRequest(http.MethodPost, "/rotate", strings.NewReader(`JSON invalido`))
		req.Header.Set("Authorization", header)
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != 401 {
			t.Errorf("estado: %d; esperado: 401", res.StatusCode)
		}
		var body struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		_ = res.Body.Close()
		if err != nil || body.Error == "" {
			t.Fatalf("error JSON invalido: %v", err)
		}
	}
}
