package web

import (
	"github.com/gofiber/fiber/v3"
	"path/filepath"
)

// Register publica solo tres archivos conocidos; no acepta rutas del usuario.
// Compartir origen con la API evita CORS y un proxy adicional en el despliegue.
func Register(app *fiber.App, directory string) {
	if directory == "" {
		directory = filepath.Join("..", "frontend")
	}
	directory, _ = filepath.Abs(directory)
	files := map[string]string{"/": "index.html", "/app.js": "app.js", "/styles.css": "styles.css"}
	for route, file := range files {
		path := filepath.Join(directory, file)
		app.Get(route, func(c fiber.Ctx) error {
			c.Set("X-Content-Type-Options", "nosniff")
			c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'self'")
			return c.SendFile(path)
		})
	}
}
