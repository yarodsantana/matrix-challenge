package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestConfiguredAppRejectsInvalidEnvironment(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"STATISTICS_URL", "ftp://example.com"},
		{"STATISTICS_URL", "https://user:password@example.com/statistics"},
		{"STATISTICS_URL", "http://example.com/statistics?x=1"},
		{"STATISTICS_TIMEOUT_SECONDS", "0"},
		{"STATISTICS_TIMEOUT_SECONDS", "invalid"},
		{"JWT_SECRET", ""},
		{"SERVICE_JWT_SECRET", testClientSecret},
	} {
		t.Run(tc.key+" invalid", func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := configuredApp(); err == nil {
				t.Fatal("se acepto configuracion invalida")
			}
		})
	}
}

func TestHTTPFrameworkErrorsUseJSON(t *testing.T) {
	app := newApp()
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/missing", "", 404},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testToken(t))
		res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
		if err != nil {
			t.Fatal(err)
		}
		var body struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		_ = res.Body.Close()
		if res.StatusCode != tc.status || err != nil || body.Error == "" {
			t.Fatalf("error HTTP invalido: %d %v", res.StatusCode, err)
		}
	}
}

// app.Test devuelve un error de transporte para cuerpos excesivos; este caso
// se comprueba sobre un puerto real para observar la respuesta HTTP del servidor.
func TestOversizedRequestUsesJSON(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	app := newApp()
	stopped := make(chan error, 1)
	go func() { stopped <- app.Listener(listener) }()
	t.Cleanup(func() {
		_ = app.ShutdownWithTimeout(3 * time.Second)
		_ = listener.Close()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			t.Error("servidor no detenido")
		}
	})
	// Se envia Content-Length excesivo sin subir el cuerpo. Asi se observa el
	// rechazo antes del upload y se evita la carrera de reset TCP en Windows.
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprintf(conn, "POST /qr HTTP/1.1\r\nHost: localhost\r\nContent-Type: application/json\r\nAuthorization: Bearer %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n", testToken(t), 1024*1024+1)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Error string `json:"error"`
	}
	err = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != 413 || err != nil || body.Error == "" {
		t.Fatalf("error HTTP invalido: %d %v", res.StatusCode, err)
	}
}
