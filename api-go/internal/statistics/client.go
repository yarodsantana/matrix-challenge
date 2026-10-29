package statistics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Result mantiene el contrato original de estadisticas para una matriz.
type Result struct {
	Max        float64 `json:"max"`
	Min        float64 `json:"min"`
	Average    float64 `json:"average"`
	Sum        float64 `json:"sum"`
	IsDiagonal bool    `json:"isDiagonal"`
}

type QRResult struct {
	Result
	MatrixCount       int               `json:"matrixCount"`
	ValueCount        int               `json:"valueCount"`
	DiagonalMatrices  []string          `json:"diagonalMatrices"`
	DiagonalTolerance float64           `json:"diagonalTolerance"`
	PerMatrix         map[string]Result `json:"perMatrix"`
}

var ErrRejected = errors.New("datos fuera del rango admitido por API 2")

type Client struct {
	Token      func() (string, error)
	URL        string
	HTTPClient *http.Client
}

func (c *Client) Calculate(ctx context.Context, matrix [][]float64) (Result, error) {
	data, err := c.post(ctx, c.URL, matrix)
	if err != nil {
		return Result{}, err
	}
	return decodeResult(data)
}

func (c *Client) CalculateQR(ctx context.Context, q, r [][]float64) (QRResult, error) {
	data, err := c.post(ctx, strings.TrimRight(c.URL, "/")+"/qr", map[string]any{"q": q, "r": r})
	if err != nil {
		return QRResult{}, err
	}
	aggregate, err := decodeResult(data)
	if err != nil {
		return QRResult{}, err
	}
	var result QRResult
	if err := json.Unmarshal(data, &result); err != nil {
		return QRResult{}, err
	}
	result.Result = aggregate
	expectedCount := len(q)*len(q[0]) + len(r)*len(r[0])
	if result.MatrixCount != 2 || result.ValueCount != expectedCount || result.DiagonalMatrices == nil || result.DiagonalTolerance != 1e-10 {
		return QRResult{}, errors.New("metadatos QR de API 2 invalidos")
	}
	var raw struct {
		PerMatrix map[string]json.RawMessage `json:"perMatrix"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return QRResult{}, err
	}
	result.PerMatrix = make(map[string]Result, 2)
	diagonalNames := make(map[string]bool)
	for _, name := range result.DiagonalMatrices {
		if (name != "q" && name != "r") || diagonalNames[name] {
			return QRResult{}, errors.New("lista diagonal QR invalida")
		}
		diagonalNames[name] = true
	}
	for _, name := range []string{"q", "r"} {
		stats, err := decodeResult(raw.PerMatrix[name])
		if err != nil {
			return QRResult{}, fmt.Errorf("estadisticas de %s: %w", name, err)
		}
		if stats.IsDiagonal != diagonalNames[name] {
			return QRResult{}, errors.New("diagonales QR inconsistentes")
		}
		result.PerMatrix[name] = stats
	}
	if result.IsDiagonal != (len(result.DiagonalMatrices) > 0) {
		return QRResult{}, errors.New("indicador diagonal QR inconsistente")
	}
	return result, nil
}

// post comparte autenticacion, timeout del cliente y limites entre los dos contratos.
func (c *Client) post(ctx context.Context, url string, payload any) ([]byte, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("codificar matriz: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("crear solicitud: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token == nil {
		return nil, errors.New("falta proveedor de token de servicio")
	}
	token, err := c.Token()
	if err != nil {
		return nil, fmt.Errorf("firmar token de servicio: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llamar API 2: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, ErrRejected
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 2 devolvio estado %d", resp.StatusCode)
	}
	const maxBytes = 64 * 1024
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("leer API 2: %w", err)
	}
	if len(data) > maxBytes {
		return nil, errors.New("respuesta de API 2 demasiado grande")
	}
	return data, nil
}

func decodeResult(data []byte) (Result, error) {
	var body struct {
		Max        *float64 `json:"max"`
		Min        *float64 `json:"min"`
		Average    *float64 `json:"average"`
		Sum        *float64 `json:"sum"`
		IsDiagonal *bool    `json:"isDiagonal"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return Result{}, fmt.Errorf("JSON de API 2 invalido: %w", err)
	}
	if body.Max == nil || body.Min == nil || body.Average == nil || body.Sum == nil || body.IsDiagonal == nil {
		return Result{}, errors.New("respuesta de API 2 incompleta")
	}
	return Result{Max: *body.Max, Min: *body.Min, Average: *body.Average, Sum: *body.Sum, IsDiagonal: *body.IsDiagonal}, nil
}
