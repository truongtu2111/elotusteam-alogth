package main

import (
	"fmt"
	"reflect"
	"testing"
)

// TestGrayCodeBasicCases tests basic functionality for small values of n
func TestGrayCodeBasicCases(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected []int
	}{
		{"n=0", 0, []int{0}},
		{"n=1", 1, []int{0, 1}},
		{"n=2", 2, []int{0, 1, 3, 2}},
		{"n=3", 3, []int{0, 1, 3, 2, 6, 7, 5, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grayCode(tt.n)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("grayCode(%d) = %v, want %v", tt.n, result, tt.expected)
			}
		})
	}
}

// TestGrayCodeFormulaBasicCases tests the formula approach for basic cases
func TestGrayCodeFormulaBasicCases(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected []int
	}{
		{"n=1", 1, []int{0, 1}},
		{"n=2", 2, []int{0, 1, 3, 2}},
		{"n=3", 3, []int{0, 1, 3, 2, 6, 7, 5, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grayCodeFormula(tt.n)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("grayCodeFormula(%d) = %v, want %v", tt.n, result, tt.expected)
			}
		})
	}
}

// TestGrayCodeProperties validates Gray code properties
func TestGrayCodeProperties(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5}

	for _, n := range testCases {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			sequence := grayCode(n)

			// Test 1: Correct length (2^n)
			expectedLen := 1 << n
			if len(sequence) != expectedLen {
				t.Errorf("Length = %d, want %d", len(sequence), expectedLen)
			}

			// Test 2: Starts with 0
			if sequence[0] != 0 {
				t.Errorf("First element = %d, want 0", sequence[0])
			}

			// Test 3: All elements in range [0, 2^n - 1]
			maxVal := (1 << n) - 1
			for i, val := range sequence {
				if val < 0 || val > maxVal {
					t.Errorf("Element at index %d = %d, out of range [0, %d]", i, val, maxVal)
				}
			}

			// Test 4: No duplicates
			seen := make(map[int]bool)
			for i, val := range sequence {
				if seen[val] {
					t.Errorf("Duplicate value %d found at index %d", val, i)
				}
				seen[val] = true
			}

			// Test 5: Adjacent elements differ by exactly one bit
			for i := 0; i < len(sequence); i++ {
				curr := sequence[i]
				next := sequence[(i+1)%len(sequence)] // wrap around

				xor := curr ^ next
				bitCount := popCount(xor)

				if bitCount != 1 {
					t.Errorf("Adjacent elements %d and %d differ by %d bits, want 1", curr, next, bitCount)
				}
			}
		})
	}
}

// TestBothApproachesEquivalent verifies both implementations produce identical results
func TestBothApproachesEquivalent(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 6}

	for _, n := range testCases {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			iterativeResult := grayCode(n)
			formulaResult := grayCodeFormula(n)

			if !reflect.DeepEqual(iterativeResult, formulaResult) {
				t.Errorf("Iterative and formula approaches differ for n=%d\nIterative: %v\nFormula: %v",
					n, iterativeResult, formulaResult)
			}
		})
	}
}

// TestGrayCodeEdgeCases tests edge cases and constraints
func TestGrayCodeEdgeCases(t *testing.T) {
	t.Run("n=0", func(t *testing.T) {
		result := grayCode(0)
		expected := []int{0}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("grayCode(0) = %v, want %v", result, expected)
		}
	})

	t.Run("n=1", func(t *testing.T) {
		result := grayCode(1)
		expected := []int{0, 1}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("grayCode(1) = %v, want %v", result, expected)
		}
	})

	// Test larger values within constraint (n <= 16)
	t.Run("n=10", func(t *testing.T) {
		result := grayCode(10)
		expectedLen := 1 << 10 // 1024
		if len(result) != expectedLen {
			t.Errorf("Length for n=10: got %d, want %d", len(result), expectedLen)
		}

		// Verify first few and last few elements follow Gray code properties
		for i := 0; i < 5; i++ {
			curr := result[i]
			next := result[i+1]
			if popCount(curr^next) != 1 {
				t.Errorf("Elements at %d and %d don't differ by 1 bit", i, i+1)
			}
		}
	})
}

// BenchmarkGrayCodeIterative benchmarks the iterative approach
func BenchmarkGrayCodeIterative(b *testing.B) {
	benchmarks := []struct {
		name string
		n    int
	}{
		{"n=5", 5},
		{"n=10", 10},
		{"n=15", 15},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = grayCode(bm.n)
			}
		})
	}
}

// BenchmarkGrayCodeFormula benchmarks the formula approach
func BenchmarkGrayCodeFormula(b *testing.B) {
	benchmarks := []struct {
		name string
		n    int
	}{
		{"n=5", 5},
		{"n=10", 10},
		{"n=15", 15},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = grayCodeFormula(bm.n)
			}
		})
	}
}

// TestHelperFunctions tests utility functions
func TestPopCount(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"three", 3, 2},    // 11 in binary
		{"seven", 7, 3},    // 111 in binary
		{"fifteen", 15, 4}, // 1111 in binary
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := popCount(tt.input)
			if result != tt.expected {
				t.Errorf("popCount(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEqualSlices(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []int
		expected bool
	}{
		{"equal_slices", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"different_values", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"different_lengths", []int{1, 2}, []int{1, 2, 3}, false},
		{"empty_slices", []int{}, []int{}, true},
		{"one_empty", []int{1}, []int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equalSlices(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("equalSlices(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
