package main

import (
	"log"
	"net/http"
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
	log.Fatal(newApp().Listen(":" + port))
}

func newApp() *fiber.App {
	url := os.Getenv("STATISTICS_URL")
	if url == "" {
		url = "http://localhost:3000/statistics"
	}
	timeout := 3 * time.Second
	if value := os.Getenv("STATISTICS_TIMEOUT_SECONDS"); value != "" {
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds < 1 || seconds > 120 {
			panic("STATISTICS_TIMEOUT_SECONDS debe estar entre 1 y 120")
		}
		timeout = time.Duration(seconds) * time.Second
	}
	return newAppWithClient(&statistics.Client{URL: url, HTTPClient: &http.Client{
		Timeout:       timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}})
}

func newAppWithClient(client *statistics.Client) *fiber.App {
	security, err := auth.New(os.Getenv("JWT_SECRET"), os.Getenv("SERVICE_JWT_SECRET"))
	if err != nil {
		panic(err)
	}
	client.Token = security.ServiceToken
	app := fiber.New(fiber.Config{BodyLimit: 1024 * 1024})
	app.Get("/health", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Post("/rotate", security.Middleware, func(c fiber.Ctx) error { return rotateHandler(c, client) })
	app.Post("/qr", security.Middleware, func(c fiber.Ctx) error { return qrHandler(c, client) })
	web.Register(app, os.Getenv("FRONTEND_DIR"))
	return app
}
