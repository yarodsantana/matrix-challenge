package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const clientKey = "test-only-client-secret-32-bytes-minimum"
const serviceKey = "test-only-service-secret-32-bytes-minimum"

func TestMiddleware(t *testing.T) {
	security, err := New(clientKey, serviceKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	base := func() jwt.MapClaims {
		return jwt.MapClaims{
			"iss": ClientIssuer, "aud": GoAudience, "sub": "test-user",
			"iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
		}
	}
	cases := []struct {
		name   string
		change func(jwt.MapClaims)
		method jwt.SigningMethod
		key    string
		want   int
	}{
		{"valido", nil, jwt.SigningMethodHS256, clientKey, 200},
		{"firma incorrecta", nil, jwt.SigningMethodHS256, serviceKey, 401},
		{"vencido", func(c jwt.MapClaims) { c["exp"] = now.Add(-time.Minute).Unix() }, jwt.SigningMethodHS256, clientKey, 401},
		{"sin expiracion", func(c jwt.MapClaims) { delete(c, "exp") }, jwt.SigningMethodHS256, clientKey, 401},
		{"audiencia incorrecta", func(c jwt.MapClaims) { c["aud"] = NodeAudience }, jwt.SigningMethodHS256, clientKey, 401},
		{"emisor incorrecto", func(c jwt.MapClaims) { c["iss"] = "otro" }, jwt.SigningMethodHS256, clientKey, 401},
		{"sin sujeto", func(c jwt.MapClaims) { delete(c, "sub") }, jwt.SigningMethodHS256, clientKey, 401},
		{"emision futura", func(c jwt.MapClaims) { c["iat"] = now.Add(time.Minute).Unix() }, jwt.SigningMethodHS256, clientKey, 401},
		{"aun no valido", func(c jwt.MapClaims) { c["nbf"] = now.Add(time.Minute).Unix() }, jwt.SigningMethodHS256, clientKey, 401},
		{"algoritmo no permitido", nil, jwt.SigningMethodHS384, clientKey, 401},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := base()
			if tc.change != nil {
				tc.change(claims)
			}
			token, err := jwt.NewWithClaims(tc.method, claims).SignedString([]byte(tc.key))
			if err != nil {
				t.Fatal(err)
			}
			app := fiber.New()
			app.Post("/rotate", security.Middleware, func(c fiber.Ctx) error { return c.SendStatus(200) })
			req := httptest.NewRequest("POST", "/rotate", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.want {
				t.Fatalf("estado: %d; esperado: %d", res.StatusCode, tc.want)
			}
			if tc.want == 401 && res.Header.Get("WWW-Authenticate") != "Bearer" {
				t.Fatal("falta WWW-Authenticate")
			}
		})
	}
	for _, header := range []string{"", "Bearer", "Basic abc", "Bearer no-es-jwt", "Bearer a b"} {
		t.Run("header "+header, func(t *testing.T) {
			app := fiber.New()
			app.Post("/rotate", security.Middleware, func(c fiber.Ctx) error { t.Error("no debe llegar al handler"); return nil })
			req := httptest.NewRequest("POST", "/rotate", nil)
			req.Header.Set("Authorization", header)
			res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 401 {
				t.Fatalf("estado: %d", res.StatusCode)
			}
		})
	}
}

func TestServiceToken(t *testing.T) {
	security, _ := New(clientKey, serviceKey)
	signed, err := security.ServiceToken()
	if err != nil {
		t.Fatal(err)
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(signed, claims, func(t *jwt.Token) (any, error) { return []byte(serviceKey), nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithAudience(NodeAudience), jwt.WithIssuer(ServiceIssuer), jwt.WithSubject(ServiceIssuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		t.Fatalf("token de servicio invalido: %v", err)
	}
	if claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) != 2*time.Minute {
		t.Fatal("duracion incorrecta")
	}
}

func TestRejectsInvalidConfiguration(t *testing.T) {
	for _, pair := range [][2]string{{"", serviceKey}, {clientKey, ""}, {"short", serviceKey}, {clientKey, clientKey}} {
		if _, err := New(pair[0], pair[1]); err == nil {
			t.Fatal("se acepto configuracion insegura")
		}
	}
}
