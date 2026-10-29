package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"mime"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/golang-jwt/jwt/v5"
)

// RegisterTokenEndpoint permite acceso de demostracion con una clave compartida.
// Sin DEMO_PASSWORD la emision queda deshabilitada; las APIs mantienen su JWT.
func RegisterTokenEndpoint(app *fiber.App, security *Config, password string) {
	expected := sha256.Sum256([]byte(password))
	attempts := limiter.New(limiter.Config{
		Max: 5, Expiration: time.Minute, SkipSuccessfulRequests: true,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "demasiados intentos; espera un minuto"})
		},
	})
	app.Post("/auth/token", func(c fiber.Ctx) error {
		c.Set("Cache-Control", "no-store")
		return c.Next()
	}, attempts, func(c fiber.Ctx) error {
		if len(password) < 12 {
			return c.Status(503).JSON(fiber.Map{"error": "generacion de tokens no configurada"})
		}
		mediaType, _, err := mime.ParseMediaType(c.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			return c.Status(415).JSON(fiber.Map{"error": "Content-Type debe ser application/json"})
		}
		if len(c.Body()) > 4096 {
			return c.Status(413).JSON(fiber.Map{"error": "solicitud demasiado grande"})
		}
		var input struct {
			Password string `json:"password"`
		}
		if err := c.Bind().JSON(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "JSON invalido"})
		}
		supplied := sha256.Sum256([]byte(input.Password))
		if subtle.ConstantTimeCompare(supplied[:], expected[:]) != 1 {
			return c.Status(401).JSON(fiber.Map{"error": "contraseña incorrecta"})
		}
		now := time.Now()
		claims := jwt.RegisteredClaims{
			Issuer: ClientIssuer, Audience: jwt.ClaimStrings{GoAudience}, Subject: "demo-reviewer",
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(security.clientSecret)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"accessToken": token, "tokenType": "Bearer", "expiresIn": 900})
	})
}
