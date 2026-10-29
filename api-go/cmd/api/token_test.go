package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/matrix-api/internal/statistics"
	"github.com/gofiber/fiber/v3"
)

func TestDemoTokenAuthorizesRotate(t *testing.T) {
	t.Setenv("DEMO_PASSWORD", "test-only-demo-password")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"max":1,"min":1,"average":1,"sum":1,"isDiagonal":true}`))
	}))
	defer upstream.Close()
	app := newAppWithClient(&statistics.Client{URL: upstream.URL, HTTPClient: upstream.Client()})
	req := httptest.NewRequest("POST", "/auth/token", strings.NewReader(`{"password":"test-only-demo-password"}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	var body struct{ AccessToken string }
	err = json.NewDecoder(res.Body).Decode(&body)
	_ = res.Body.Close()
	if err != nil || res.StatusCode != 200 || body.AccessToken == "" {
		t.Fatalf("token no emitido: %d %v", res.StatusCode, err)
	}
	rotate := httptest.NewRequest("POST", "/rotate", strings.NewReader(`[[1]]`))
	rotate.Header.Set("Content-Type", "application/json")
	rotate.Header.Set("Authorization", "Bearer "+body.AccessToken)
	result, err := app.Test(rotate, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Body.Close()
	if result.StatusCode != 200 {
		t.Fatalf("rotacion no autorizada: %d", result.StatusCode)
	}
}
