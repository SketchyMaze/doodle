package physics

// Lerp performs linear interpolation between two numbers.
//
// a and b are the two bounds of the number, and t is a fraction between 0 and
// 1 that will return a number between a and b. If t=0, returns a; if t=1,
// returns b.
func Lerp(a, b, t float64) float64 {
	return (1.0-t)*a + t*b
}

// LerpInt runs lerp using integers.
func LerpInt(a, b int, t float64) float64 {
	return Lerp(float64(a), float64(b), t)
}
