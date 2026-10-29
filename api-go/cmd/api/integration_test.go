//go:build integration

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGoToRealExpress(t *testing.T) {
	binary := os.Getenv("NODE_BINARY")
	if binary == "" {
		binary = "node"
	}
	cmd := exec.Command(binary, "--input-type=module", "-e", `
 import { createApp } from './src/app.js';
 const server = createApp().listen(0, '127.0.0.1', () => console.log(server.address().port));
 `)
	cmd.Dir = filepath.Join("..", "..", "..", "api-node")
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			ready <- scanner.Text()
		} else {
			ready <- ""
		}
	}()
	var port string
	select {
	case port = <-ready:
		if port == "" {
			t.Fatal("Express no anuncio un puerto")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Express no inicio a tiempo")
	}
	t.Setenv("STATISTICS_URL", fmt.Sprintf("http://127.0.0.1:%s/statistics", port))
	// Verifica tambien que el token emitido por la herramienta Node sea aceptado por Go.
	issuer := exec.Command(binary, filepath.Join("scripts", "create-token.mjs"))
	issuer.Dir = filepath.Join("..", "..", "..")
	output, err := issuer.Output()
	if err != nil {
		t.Fatalf("herramienta local de tokens: %v", err)
	}
	clientToken := strings.TrimSpace(string(output))
	app := newApp()
	cases := []struct {
		name, input string
		diagonal    bool
	}{
		{"rectangular", `[[1,2,3],[4,5,6]]`, false},
		// Una diagonal no nula rota a una antidiagonal: Node debe evaluar la salida.
		{"diagonal de entrada deja de ser diagonal", `[[1,0],[0,2]]`, false},
		{"antidiagonal de entrada se vuelve diagonal", `[[0,1],[2,0]]`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/rotate", strings.NewReader(tc.input))
			req.Header.Set("Authorization", "Bearer "+clientToken)
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				t.Fatalf("estado: %d", res.StatusCode)
			}
			var body struct {
				Matrix     [][]float64 `json:"matrix"`
				Statistics struct {
					Max        float64 `json:"max"`
					Min        float64 `json:"min"`
					Average    float64 `json:"average"`
					Sum        float64 `json:"sum"`
					IsDiagonal bool    `json:"isDiagonal"`
				} `json:"statistics"`
			}
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Statistics.IsDiagonal != tc.diagonal {
				t.Fatalf("diagonal de la salida incorrecta: %+v", body)
			}
			if tc.name == "rectangular" {
				if body.Statistics.Max != 6 || body.Statistics.Min != 1 || body.Statistics.Sum != 21 || body.Statistics.Average != 3.5 {
					t.Fatalf("estadisticas incorrectas: %+v", body)
				}
			}
		})
	}

	if runtime.GOOS == "windows" {
		t.Run("smoke PowerShell por HTTP real con JWT", func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			liveApp := newApp()
			stopped := make(chan error, 1)
			go func() { stopped <- liveApp.Listener(listener) }()
			t.Cleanup(func() {
				_ = liveApp.ShutdownWithTimeout(3 * time.Second)
				_ = listener.Close()
				select {
				case <-stopped:
				case <-time.After(5 * time.Second):
					t.Error("Fiber no se detuvo")
				}
			})
			script, err := filepath.Abs(filepath.Join("..", "..", "..", "scripts", "smoke-test.ps1"))
			if err != nil {
				t.Fatal(err)
			}
			smoke := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script,
				"-BaseUrl", "http://"+listener.Addr().String(), "-Token", clientToken)
			output, err := smoke.CombinedOutput()
			if err != nil {
				t.Fatalf("smoke fallo: %v\n%s", err, output)
			}
			t.Log(string(output))
		})
	}
}
