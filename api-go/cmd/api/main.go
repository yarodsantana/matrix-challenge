package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"example.com/matrix-api/internal/auth"
	"example.com/matrix-api/internal/statistics"
	"example.com/matrix-api/internal/web"
	"github.com/gofiber/fiber/v3"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		log.Fatal("PORT debe ser un entero entre 1 y 65535")
	}
	app, err := configuredApp()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(app.Listen(":" + port))
}

// configuredApp valida configuracion antes de abrir el puerto, sin panics productivos.
func configuredApp() (*fiber.App, error) {
	endpoint := os.Getenv("STATISTICS_URL")
	if endpoint == "" {
		endpoint = "http://localhost:3000/statistics"
	}
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return nil, errors.New("STATISTICS_URL debe ser una URL HTTP(S) sin credenciales, query ni fragmento")
	}
	timeout := 3 * time.Second
	if value := os.Getenv("STATISTICS_TIMEOUT_SECONDS"); value != "" {
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds < 1 || seconds > 120 {
			return nil, errors.New("STATISTICS_TIMEOUT_SECONDS debe estar entre 1 y 120")
		}
		timeout = time.Duration(seconds) * time.Second
	}
	return buildApp(&statistics.Client{URL: endpoint, HTTPClient: &http.Client{
		Timeout:       timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}})
}

func buildApp(client *statistics.Client) (*fiber.App, error) {
	security, err := auth.New(os.Getenv("JWT_SECRET"), os.Getenv("SERVICE_JWT_SECRET"))
	if err != nil {
		return nil, err
	}
	client.Token = security.ServiceToken
	app := fiber.New(fiber.Config{BodyLimit: 1024 * 1024, ErrorHandler: jsonError})
	app.Get("/health", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Post("/rotate", security.Middleware, func(c fiber.Ctx) error { return rotateHandler(c, client) })
	app.Post("/qr", security.Middleware, func(c fiber.Ctx) error { return qrHandler(c, client) })
	if err := web.Register(app, os.Getenv("FRONTEND_DIR")); err != nil {
		return nil, err
	}
	return app, nil
}
