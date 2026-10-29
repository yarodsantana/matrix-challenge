package main

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"path/filepath"
	"testing"
	"time"
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
