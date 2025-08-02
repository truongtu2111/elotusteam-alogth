package main

import (
	"reflect"
	"testing"
)

// Test cases for all solutions
var testCases = []struct {
	name     string
	n        int
	edges    [][]int
	expected []int
}{
	{
		name:     "Example 1 - Complex tree",
		n:        6,
		edges:    [][]int{{0, 1}, {0, 2}, {2, 3}, {2, 4}, {2, 5}},
		expected: []int{8, 12, 6, 10, 10, 10},
	},
	{
		name:     "Example 2 - Single node",
		n:        1,
		edges:    [][]int{},
		expected: []int{0},
	},
	{
		name:     "Example 3 - Two nodes",
		n:        2,
		edges:    [][]int{{1, 0}},
		expected: []int{1, 1},
	},
	{
		name:     "Linear tree",
		n:        4,
		edges:    [][]int{{0, 1}, {1, 2}, {2, 3}},
		expected: []int{6, 4, 4, 6},
	},
	{
		name:     "Star tree",
		n:        5,
		edges:    [][]int{{0, 1}, {0, 2}, {0, 3}, {0, 4}},
		expected: []int{4, 7, 7, 7, 7},
	},
	{
		name:     "Three nodes linear",
		n:        3,
		edges:    [][]int{{0, 1}, {1, 2}},
		expected: []int{3, 2, 3},
	},
	{
		name:     "Balanced binary tree",
		n:        7,
		edges:    [][]int{{0, 1}, {0, 2}, {1, 3}, {1, 4}, {2, 5}, {2, 6}},
		expected: []int{10, 11, 11, 16, 16, 16, 16},
	},
}

func TestSumOfDistancesInTreeNaive(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sumOfDistancesInTreeNaive(tc.n, tc.edges)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("sumOfDistancesInTreeNaive() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestSumOfDistancesInTree(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sumOfDistancesInTree(tc.n, tc.edges)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("sumOfDistancesInTree() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestSumOfDistancesInTreeAlt(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sumOfDistancesInTreeAlt(tc.n, tc.edges)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("sumOfDistancesInTreeAlt() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

// Test helper function for BFS distance calculation
func TestCalculateDistanceSum(t *testing.T) {
	tests := []struct {
		name     string
		source   int
		graph    [][]int
		n        int
		expected int
	}{
		{
			name:     "Simple linear graph",
			source:   0,
			graph:    [][]int{{1}, {0, 2}, {1}},
			n:        3,
			expected: 3, // distance 1 to node 1, distance 2 to node 2
		},
		{
			name:     "Star graph from center",
			source:   0,
			graph:    [][]int{{1, 2, 3}, {0}, {0}, {0}},
			n:        4,
			expected: 3, // distance 1 to each of nodes 1, 2, 3
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDistanceSum(tt.source, tt.graph, tt.n)
			if result != tt.expected {
				t.Errorf("calculateDistanceSum() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests to compare performance
func BenchmarkSumOfDistancesInTreeNaive(b *testing.B) {
	n := 100
	edges := make([][]int, n-1)
	// Create a linear tree for benchmarking
	for i := 0; i < n-1; i++ {
		edges[i] = []int{i, i + 1}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumOfDistancesInTreeNaive(n, edges)
	}
}

func BenchmarkSumOfDistancesInTree(b *testing.B) {
	n := 100
	edges := make([][]int, n-1)
	// Create a linear tree for benchmarking
	for i := 0; i < n-1; i++ {
		edges[i] = []int{i, i + 1}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumOfDistancesInTree(n, edges)
	}
}

func BenchmarkSumOfDistancesInTreeAlt(b *testing.B) {
	n := 100
	edges := make([][]int, n-1)
	// Create a linear tree for benchmarking
	for i := 0; i < n-1; i++ {
		edges[i] = []int{i, i + 1}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumOfDistancesInTreeAlt(n, edges)
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Empty edges with single node", func(t *testing.T) {
		result := sumOfDistancesInTree(1, [][]int{})
		expected := []int{0}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	t.Run("Large linear tree", func(t *testing.T) {
		n := 10
		edges := make([][]int, n-1)
		for i := 0; i < n-1; i++ {
			edges[i] = []int{i, i + 1}
		}
		
		// For a linear tree of length n, the sum of distances from node i is:
		// sum of (|i-j|) for all j != i
		result := sumOfDistancesInTree(n, edges)
		
		// Verify the result makes sense (should be symmetric around the middle)
		if len(result) != n {
			t.Errorf("Expected result length %d, got %d", n, len(result))
		}
		
		// For linear tree, nodes at the ends should have larger sums
		if result[0] <= result[n/2] || result[n-1] <= result[n/2] {
			t.Errorf("End nodes should have larger distance sums than middle nodes")
		}
	})
}

// Test consistency between all three implementations
func TestConsistencyBetweenImplementations(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result1 := sumOfDistancesInTreeNaive(tc.n, tc.edges)
			result2 := sumOfDistancesInTree(tc.n, tc.edges)
			result3 := sumOfDistancesInTreeAlt(tc.n, tc.edges)
			
			if !reflect.DeepEqual(result1, result2) {
				t.Errorf("Naive and optimized solutions differ: %v vs %v", result1, result2)
			}
			
			if !reflect.DeepEqual(result2, result3) {
				t.Errorf("Optimized and alternative solutions differ: %v vs %v", result2, result3)
			}
			
			if !reflect.DeepEqual(result1, tc.expected) {
				t.Errorf("All solutions should match expected result: got %v, expected %v", result1, tc.expected)
			}
		})
	}
}