package main

import (
	"context"
	"errors"
	"log"
	"mime"

	"example.com/matrix-api/internal/matrix"
	"example.com/matrix-api/internal/statistics"
	"github.com/gofiber/fiber/v3"
)

// readMatrix conserva las reglas y errores del endpoint original.
func readMatrix(c fiber.Ctx) ([][]float64, error) {
	contentType, _, err := mime.ParseMediaType(c.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		return nil, fiber.NewError(415, "Content-Type debe ser application/json")
	}
	var body [][]*float64
	if err := c.Bind().JSON(&body); err != nil {
		return nil, fiber.NewError(400, "el cuerpo debe ser una matriz JSON de numeros")
	}
	input := make([][]float64, len(body))
	for i, row := range body {
		input[i] = make([]float64, len(row))
		for j, value := range row {
			if value == nil {
				return nil, fiber.NewError(400, "la matriz no admite elementos null")
			}
			input[i][j] = *value
		}
	}
	if err := matrix.Validate(input); err != nil {
		return nil, fiber.NewError(400, err.Error())
	}
	return input, nil
}

func inputError(c fiber.Ctx, err error) error {
	var httpError *fiber.Error
	if errors.As(err, &httpError) {
		return c.Status(httpError.Code).JSON(fiber.Map{"error": httpError.Message})
	}
	return c.Status(400).JSON(fiber.Map{"error": err.Error()})
}

func upstreamError(c fiber.Ctx, err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return c.Status(504).JSON(fiber.Map{"error": "API 2 excedio el tiempo de espera"})
	}
	if errors.Is(err, statistics.ErrRejected) {
		return c.Status(422).JSON(fiber.Map{"error": "la suma excede el rango numerico admitido"})
	}
	log.Printf("error de API 2: %v", err)
	return c.Status(502).JSON(fiber.Map{"error": "no se pudo obtener una respuesta valida de API 2"})
}

func rotateHandler(c fiber.Ctx, client *statistics.Client) error {
	input, err := readMatrix(c)
	if err != nil {
		return inputError(c, err)
	}
	rotated, err := matrix.RotateClockwise(input)
	if err != nil {
		return inputError(c, err)
	}
	stats, err := client.Calculate(c.Context(), rotated)
	if err != nil {
		return upstreamError(c, err)
	}
	return c.JSON(fiber.Map{"matrix": rotated, "statistics": stats})
}

// qrHandler respeta A=QR sobre la entrada y expone la rotacion por separado.
// Las estadisticas las calcula exclusivamente Node sobre Q y R.
func qrHandler(c fiber.Ctx, client *statistics.Client) error {
	input, err := readMatrix(c)
	if err != nil {
		return inputError(c, err)
	}
	qr, err := matrix.FactorQR(input)
	if err != nil {
		if errors.Is(err, matrix.ErrNumericRange) {
			return c.Status(422).JSON(fiber.Map{"error": err.Error()})
		}
		return inputError(c, err)
	}
	rotated, err := matrix.RotateClockwise(input)
	if err != nil {
		return inputError(c, err)
	}
	stats, err := client.CalculateQR(c.Context(), qr.Q, qr.R)
	if err != nil {
		return upstreamError(c, err)
	}
	return c.JSON(fiber.Map{
		"original": input, "rotated": rotated, "q": qr.Q, "r": qr.R, "statistics": stats,
		"metadata": fiber.Map{"factorization": "A = Q R", "qrMode": "full", "rotationDegrees": 90, "rotationDirection": "clockwise"},
	})
}
