package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"example.com/matrix-api/internal/statistics"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const testClientSecret = "test-only-client-secret-32-bytes-minimum"
const testServiceSecret = "test-only-service-secret-32-bytes-minimum"

func TestMain(m *testing.M) {
	uiDir, _ := filepath.Abs(filepath.Join("..", "..", "..", "frontend"))
	_ = os.Setenv("FRONTEND_DIR", uiDir)
	// Claves publicas de pruebas, nunca utilizadas como valores productivos.
	_ = os.Setenv("JWT_SECRET", testClientSecret)
	_ = os.Setenv("SERVICE_JWT_SECRET", testServiceSecret)
	os.Exit(m.Run())
}

func testToken(t *testing.T) string {
	t.Helper()
	now := time.Now()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "matrix-challenge", "aud": "matrix-api-go", "sub": "test-user",
		"iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
	}).SignedString([]byte(testClientSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// Adaptadores de pruebas: el arranque productivo propaga errores de configuracion.
func newApp() *fiber.App {
	app, err := configuredApp()
	if err != nil {
		panic(err)
	}
	return app
}

func newAppWithClient(client *statistics.Client) *fiber.App {
	app, err := buildApp(client)
	if err != nil {
		panic(err)
	}
	return app
}
