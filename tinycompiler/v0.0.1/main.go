package main

import "fmt"

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

// square root of a fixed-point number
// stored in a 32 bit integer variable, shift is the precision
func sqrt(n, shift int) int {
	var x, xold, none int

	if n > 65535 { // pay attention to potential overflows
		return 2.0 * sqrt(n/4, shift)
	}
	x = shift        // initial guess 1.0, can do better, but oh well
	none = n * shift // need to compensate for fixp division
	for {
		xold = x
		x = (x + none/x) / 2
		if abs(x-xold) <= 1 {
			return x
		}
	}
}

func main() {
	// 25735 is approximately equal to pi * 8192;
	// expected value of the output is sqrt(pi) * 8192 approx 14519
	fmt.Println(sqrt(25735, 8192))
}
