package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const ClientIssuer = "matrix-challenge"
const GoAudience = "matrix-api-go"
const ServiceIssuer = "matrix-api-go"
const NodeAudience = "matrix-api-node"

// Config mantiene claves separadas para cliente->Go y Go->Node.
type Config struct {
	clientSecret  []byte
	serviceSecret []byte
}

func New(clientSecret, serviceSecret string) (*Config, error) {
	if len([]byte(clientSecret)) < 32 || len([]byte(serviceSecret)) < 32 {
		return nil, errors.New("JWT_SECRET y SERVICE_JWT_SECRET requieren al menos 32 bytes cada uno")
	}
	if clientSecret == serviceSecret {
		return nil, errors.New("las claves JWT deben ser diferentes")
	}
	return &Config{clientSecret: []byte(clientSecret), serviceSecret: []byte(serviceSecret)}, nil
}

func (a *Config) Middleware(c fiber.Ctx) error {
	parts := strings.Fields(c.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return unauthorized(c)
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		return a.clientSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(),
		jwt.WithIssuer(ClientIssuer), jwt.WithAudience(GoAudience), jwt.WithIssuedAt(), jwt.WithLeeway(5*time.Second))
	if err != nil || !token.Valid || claims.Subject == "" {
		return unauthorized(c)
	}
	return c.Next()
}

func unauthorized(c fiber.Ctx) error {
	c.Set("WWW-Authenticate", "Bearer")
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token ausente o invalido"})
}

// Un token de servicio dura dos minutos y se emite para cada llamada a Node.
func (a *Config) ServiceToken() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer: ServiceIssuer, Subject: ServiceIssuer, Audience: jwt.ClaimStrings{NodeAudience},
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.serviceSecret)
}
