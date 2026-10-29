package matrix

// RotateClockwise rota 90 grados en sentido horario y conserva la entrada.
func RotateClockwise(input [][]float64) ([][]float64, error) {
	if err := Validate(input); err != nil {
		return nil, err
	}
	rows, cols := len(input), len(input[0])
	output := makeMatrix(cols, rows)
	for r, row := range input {
		for c, value := range row {
			output[c][rows-1-r] = value
		}
	}
	return output, nil
}
