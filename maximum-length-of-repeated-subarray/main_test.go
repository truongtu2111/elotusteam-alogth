package main

import (
	"fmt"
	"testing"
)

// Test cases for all solutions
var testCases = []struct {
	name     string
	nums1    []int
	nums2    []int
	expected int
}{
	{
		name:     "Example 1 - Basic case",
		nums1:    []int{1, 2, 3, 2, 1},
		nums2:    []int{3, 2, 1, 4, 7},
		expected: 3,
	},
	{
		name:     "Example 2 - All same elements",
		nums1:    []int{0, 0, 0, 0, 0},
		nums2:    []int{0, 0, 0, 0, 0},
		expected: 5,
	},
	{
		name:     "No common subarray",
		nums1:    []int{1, 2, 3},
		nums2:    []int{4, 5, 6},
		expected: 0,
	},
	{
		name:     "Empty first array",
		nums1:    []int{},
		nums2:    []int{1, 2, 3},
		expected: 0,
	},
	{
		name:     "Empty second array",
		nums1:    []int{1, 2, 3},
		nums2:    []int{},
		expected: 0,
	},
	{
		name:     "Single element match",
		nums1:    []int{1},
		nums2:    []int{1},
		expected: 1,
	},
	{
		name:     "Single element no match",
		nums1:    []int{1},
		nums2:    []int{2},
		expected: 0,
	},
	{
		name:     "Partial overlap",
		nums1:    []int{1, 0, 1, 0, 1},
		nums2:    []int{1, 1, 1, 1, 1},
		expected: 1,
	},
	{
		name:     "Long common subarray",
		nums1:    []int{0, 1, 1, 1, 1},
		nums2:    []int{1, 0, 1, 0, 1},
		expected: 2,
	},
	{
		name:     "Identical arrays",
		nums1:    []int{1, 2, 3, 4, 5},
		nums2:    []int{1, 2, 3, 4, 5},
		expected: 5,
	},
	{
		name:     "Common subarray at beginning",
		nums1:    []int{1, 2, 3, 4, 5},
		nums2:    []int{1, 2, 3, 6, 7},
		expected: 3,
	},
	{
		name:     "Common subarray at end",
		nums1:    []int{1, 2, 3, 4, 5},
		nums2:    []int{6, 7, 3, 4, 5},
		expected: 3,
	},
	{
		name:     "Common subarray in middle",
		nums1:    []int{1, 2, 3, 4, 5},
		nums2:    []int{6, 2, 3, 4, 7},
		expected: 3,
	},
	{
		name:     "Multiple common subarrays",
		nums1:    []int{1, 2, 1, 2, 3},
		nums2:    []int{2, 1, 2, 3, 4},
		expected: 4,
	},
}

func TestFindLengthBruteForce(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := findLengthBruteForce(tc.nums1, tc.nums2)
			if result != tc.expected {
				t.Errorf("findLengthBruteForce() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestFindLength(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := findLength(tc.nums1, tc.nums2)
			if result != tc.expected {
				t.Errorf("findLength() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestFindLengthOptimized(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := findLengthOptimized(tc.nums1, tc.nums2)
			if result != tc.expected {
				t.Errorf("findLengthOptimized() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestFindLengthRollingHash(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := findLengthRollingHash(tc.nums1, tc.nums2)
			if result != tc.expected {
				t.Errorf("findLengthRollingHash() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

// Test helper function for common subarray detection
func TestHasCommonSubarray(t *testing.T) {
	tests := []struct {
		name     string
		nums1    []int
		nums2    []int
		length   int
		expected bool
	}{
		{
			name:     "Length 0 always true",
			nums1:    []int{1, 2, 3},
			nums2:    []int{4, 5, 6},
			length:   0,
			expected: true,
		},
		{
			name:     "Length 1 match",
			nums1:    []int{1, 2, 3},
			nums2:    []int{3, 4, 5},
			length:   1,
			expected: true,
		},
		{
			name:     "Length 2 match",
			nums1:    []int{1, 2, 3},
			nums2:    []int{0, 1, 2},
			length:   2,
			expected: true,
		},
		{
			name:     "Length 2 no match",
			nums1:    []int{1, 2, 3},
			nums2:    []int{4, 5, 6},
			length:   2,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasCommonSubarray(tt.nums1, tt.nums2, tt.length)
			if result != tt.expected {
				t.Errorf("hasCommonSubarray() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Test min helper function
func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}
	
	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// Benchmark tests to compare performance
func BenchmarkFindLengthBruteForce(b *testing.B) {
	nums1 := make([]int, 100)
	nums2 := make([]int, 100)
	for i := 0; i < 100; i++ {
		nums1[i] = i % 10
		nums2[i] = (i + 5) % 10
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findLengthBruteForce(nums1, nums2)
	}
}

func BenchmarkFindLength(b *testing.B) {
	nums1 := make([]int, 100)
	nums2 := make([]int, 100)
	for i := 0; i < 100; i++ {
		nums1[i] = i % 10
		nums2[i] = (i + 5) % 10
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findLength(nums1, nums2)
	}
}

func BenchmarkFindLengthOptimized(b *testing.B) {
	nums1 := make([]int, 100)
	nums2 := make([]int, 100)
	for i := 0; i < 100; i++ {
		nums1[i] = i % 10
		nums2[i] = (i + 5) % 10
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findLengthOptimized(nums1, nums2)
	}
}

func BenchmarkFindLengthRollingHash(b *testing.B) {
	nums1 := make([]int, 100)
	nums2 := make([]int, 100)
	for i := 0; i < 100; i++ {
		nums1[i] = i % 10
		nums2[i] = (i + 5) % 10
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findLengthRollingHash(nums1, nums2)
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Both arrays empty", func(t *testing.T) {
		result := findLength([]int{}, []int{})
		expected := 0
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("Large arrays with no common subarray", func(t *testing.T) {
		nums1 := make([]int, 1000)
		nums2 := make([]int, 1000)
		for i := 0; i < 1000; i++ {
			nums1[i] = 1
			nums2[i] = 2
		}
		
		result := findLength(nums1, nums2)
		expected := 0
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("Large arrays with full common subarray", func(t *testing.T) {
		n := 100
		nums1 := make([]int, n)
		nums2 := make([]int, n)
		for i := 0; i < n; i++ {
			nums1[i] = i % 10
			nums2[i] = i % 10
		}
		
		result := findLength(nums1, nums2)
		expected := n
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("Arrays with different lengths", func(t *testing.T) {
		nums1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		nums2 := []int{3, 4, 5}
		
		result := findLength(nums1, nums2)
		expected := 3
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// Test consistency between all four implementations
func TestConsistencyBetweenImplementations(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result1 := findLengthBruteForce(tc.nums1, tc.nums2)
			result2 := findLength(tc.nums1, tc.nums2)
			result3 := findLengthOptimized(tc.nums1, tc.nums2)
			result4 := findLengthRollingHash(tc.nums1, tc.nums2)
			
			if result1 != result2 {
				t.Errorf("Brute force and DP solutions differ: %v vs %v", result1, result2)
			}
			
			if result2 != result3 {
				t.Errorf("DP and optimized DP solutions differ: %v vs %v", result2, result3)
			}
			
			if result3 != result4 {
				t.Errorf("Optimized DP and rolling hash solutions differ: %v vs %v", result3, result4)
			}
			
			if result1 != tc.expected {
				t.Errorf("All solutions should match expected result: got %v, expected %v", result1, tc.expected)
			}
		})
	}
}

// Test with random data to ensure robustness
func TestRandomData(t *testing.T) {
	// Test with various sizes and patterns
	testSizes := []int{10, 50, 100}
	
	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Random data size %d", size), func(t *testing.T) {
			// Create arrays with some overlapping patterns
			nums1 := make([]int, size)
			nums2 := make([]int, size)
			
			for i := 0; i < size; i++ {
				nums1[i] = i % 5 // Pattern: 0,1,2,3,4,0,1,2,3,4...
				nums2[i] = (i + 2) % 5 // Pattern: 2,3,4,0,1,2,3,4,0,1...
			}
			
			// All implementations should give the same result
			result1 := findLengthBruteForce(nums1, nums2)
			result2 := findLength(nums1, nums2)
			result3 := findLengthOptimized(nums1, nums2)
			result4 := findLengthRollingHash(nums1, nums2)
			
			if !(result1 == result2 && result2 == result3 && result3 == result4) {
				t.Errorf("Inconsistent results: BF=%d, DP=%d, Opt=%d, Hash=%d", result1, result2, result3, result4)
			}
		})
	}
}