package main

import "fmt"

// grayCode generates an n-bit Gray code sequence using iterative approach
// Optimizations:
// 1. Iterative instead of recursive (eliminates call stack overhead)
// 2. Pre-allocate result slice with exact capacity
// 3. In-place mirroring without extra memory for previous sequences
// 4. Direct bit manipulation for MSB setting
func grayCode(n int) []int {
	if n == 0 {
		return []int{0}
	}

	// Pre-allocate result with exact capacity 2^n
	result := make([]int, 1<<n)
	result[0] = 0 // Start with 0

	// Build Gray code iteratively for each bit level
	for i := 0; i < n; i++ {
		// Current sequence length is 2^i
		currentLen := 1 << i
		// MSB value for this level is 2^i
		msb := currentLen

		// Mirror the existing sequence and add MSB
		// Copy in reverse order to create the mirror effect
		for j := 0; j < currentLen; j++ {
			result[currentLen+j] = result[currentLen-1-j] + msb
		}
	}

	return result
}

// Alternative optimized solution using mathematical formula
// This is the most efficient approach with O(2^n) time and O(1) extra space
func grayCodeFormula(n int) []int {
	result := make([]int, 1<<n)
	for i := 0; i < 1<<n; i++ {
		// Gray code formula: G(i) = i XOR (i >> 1)
		result[i] = i ^ (i >> 1)
	}
	return result
}

func main() {
	// Test both optimized approaches
	fmt.Println("=== Iterative Approach ===")
	fmt.Println("n=1:", grayCode(1)) // [0, 1]
	fmt.Println("n=2:", grayCode(2)) // [0, 1, 3, 2]
	fmt.Println("n=3:", grayCode(3)) // [0, 1, 3, 2, 6, 7, 5, 4]

	fmt.Println("\n=== Formula Approach ===")
	fmt.Println("n=1:", grayCodeFormula(1)) // [0, 1]
	fmt.Println("n=2:", grayCodeFormula(2)) // [0, 1, 3, 2]
	fmt.Println("n=3:", grayCodeFormula(3)) // [0, 1, 3, 2, 6, 7, 5, 4]

	// Verify both approaches produce identical results
	n := 4
	iterative := grayCode(n)
	formula := grayCodeFormula(n)

	fmt.Printf("\n=== Verification (n=%d) ===\n", n)
	fmt.Printf("Sequences identical: %v\n", equalSlices(iterative, formula))
	fmt.Printf("Length: %d (expected: %d)\n", len(iterative), 1<<n)

	// Verify Gray code properties for n=3
	n = 3
	sequence := grayCode(n)
	fmt.Printf("\n=== Gray Code Properties (n=%d) ===\n", n)
	fmt.Printf("Sequence: %v\n", sequence)

	allValid := true
	for i := 0; i < len(sequence); i++ {
		curr := sequence[i]
		next := sequence[(i+1)%len(sequence)] // wrap around for last element

		// Check if adjacent numbers differ by exactly one bit
		xor := curr ^ next
		bitCount := popCount(xor)

		isValid := bitCount == 1
		allValid = allValid && isValid

		fmt.Printf("%d (%0*b) -> %d (%0*b): %d bit diff %s\n",
			curr, n, curr, next, n, next, bitCount,
			map[bool]string{true: "✓", false: "✗"}[isValid])
	}

	fmt.Printf("\nAll transitions valid: %s\n",
		map[bool]string{true: "✓ PASS", false: "✗ FAIL"}[allValid])
}

// Helper function to compare two slices
func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Optimized bit counting using built-in function concept
func popCount(x int) int {
	count := 0
	for x > 0 {
		count += x & 1
		x >>= 1
	}
	return count
}
