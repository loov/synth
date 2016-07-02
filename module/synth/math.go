package synth

import "math"

const (
	tau = math.Pi * 2
)

func sin(x float32) float32 { return float32(math.Sin(float64(x))) }
func cos(x float32) float32 { return float32(math.Cos(float64(x))) }
