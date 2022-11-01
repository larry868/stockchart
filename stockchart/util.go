package stockchart

func imin(a int, b int) int {
	if b < a {
		return b
	}
	return a
}

func imax(a int, b int) int {
	if b > a {
		return b
	}
	return a
}

func fmin(a float64, b float64) float64 {
	if b < a {
		return b
	}
	return a
}

func fmax(a float64, b float64) float64 {
	if b > a {
		return b
	}
	return a
}
