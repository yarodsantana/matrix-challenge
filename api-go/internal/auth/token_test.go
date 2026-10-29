package auth

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func tokenApp(password string) *fiber.App {
	security, _ := New(clientKey, serviceKey)
	app := fiber.New()
	RegisterTokenEndpoint(app, security, password)
	app.Post("/protected", security.Middleware, func(c fiber.Ctx) error { return c.SendStatus(200) })
	return app
}

func TestIssueClientToken(t *testing.T) {
	app := tokenApp("test-only-demo-password")
	req := httptest.NewRequest("POST", "/auth/token", strings.NewReader(`{"password":"test-only-demo-password"}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 || res.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("respuesta: %d", res.StatusCode)
	}
	var body struct {
		AccessToken, TokenType string
		ExpiresIn              int
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.TokenType != "Bearer" || body.ExpiresIn != 900 {
		t.Fatal("contrato de token incorrecto")
	}
	claims := &jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(body.AccessToken, claims, func(t *jwt.Token) (any, error) { return []byte(clientKey), nil }, jwt.WithAudience(GoAudience), jwt.WithIssuer(ClientIssuer), jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) != 15*time.Minute {
		t.Fatalf("JWT incorrecto: %v", err)
	}
	protected := httptest.NewRequest("POST", "/protected", nil)
	protected.Header.Set("Authorization", "Bearer "+body.AccessToken)
	result, err := app.Test(protected)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Body.Close()
	if result.StatusCode != 200 {
		t.Fatal("el token emitido no autoriza Go")
	}
}

func TestTokenErrors(t *testing.T) {
	for _, tc := range []struct {
		name, configured, body, contentType string
		want                                int
	}{
		{"incorrecta", "test-only-demo-password", `{"password":"wrong"}`, "application/json", 401},
		{"vacia", "test-only-demo-password", `{}`, "application/json", 401},
		{"JSON", "test-only-demo-password", `{`, "application/json", 400},
		{"tipo", "test-only-demo-password", `{}`, "text/plain", 415},
		{"deshabilitado", "", `{}`, "application/json", 503},
		{"exceso", "test-only-demo-password", strings.Repeat(" ", 4097), "application/json", 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/auth/token", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			res, err := tokenApp(tc.configured).Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.want {
				t.Fatalf("estado: %d, esperado: %d", res.StatusCode, tc.want)
			}
		})
	}
}

func TestTokenRateLimit(t *testing.T) {
	app := tokenApp("test-only-demo-password")
	for attempt := 0; attempt < 6; attempt++ {
		req := httptest.NewRequest("POST", "/auth/token", strings.NewReader(`{"password":"wrong"}`))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		want := 401
		if attempt == 5 {
			want = 429
		}
		if res.StatusCode != want {
			t.Fatalf("intento %d: %d, esperado %d", attempt, res.StatusCode, want)
		}
	}
}
