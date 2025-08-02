package main

import (
	"fmt"
)

// Solution 1: Naive approach - O(n²) time complexity
// For each node, calculate distances to all other nodes using BFS
func sumOfDistancesInTreeNaive(n int, edges [][]int) []int {
	if n == 1 {
		return []int{0}
	}
	
	// Build adjacency list
	graph := make([][]int, n)
	for _, edge := range edges {
		a, b := edge[0], edge[1]
		graph[a] = append(graph[a], b)
		graph[b] = append(graph[b], a)
	}
	
	result := make([]int, n)
	
	// For each node, calculate sum of distances to all other nodes
	for i := 0; i < n; i++ {
		result[i] = calculateDistanceSum(i, graph, n)
	}
	
	return result
}

// BFS to calculate sum of distances from a source node to all other nodes
func calculateDistanceSum(source int, graph [][]int, n int) int {
	visited := make([]bool, n)
	queue := []int{source}
	visited[source] = true
	distance := 0
	totalSum := 0
	
	for len(queue) > 0 {
		size := len(queue)
		for i := 0; i < size; i++ {
			node := queue[0]
			queue = queue[1:]
			
			if node != source {
				totalSum += distance
			}
			
			for _, neighbor := range graph[node] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
		distance++
	}
	
	return totalSum
}

// Solution 2: Optimized approach - O(n) time complexity
// Uses two-pass DFS with dynamic programming
func sumOfDistancesInTree(n int, edges [][]int) []int {
	if n == 1 {
		return []int{0}
	}
	
	// Build adjacency list
	graph := make([][]int, n)
	for _, edge := range edges {
		a, b := edge[0], edge[1]
		graph[a] = append(graph[a], b)
		graph[b] = append(graph[b], a)
	}
	
	// Arrays to store subtree information
	subtreeSize := make([]int, n)  // Number of nodes in subtree rooted at i
	subtreeSum := make([]int, n)   // Sum of distances from i to all nodes in its subtree
	result := make([]int, n)       // Final answer
	
	// First DFS: Calculate subtree sizes and sums from root (node 0)
	dfs1(0, -1, graph, subtreeSize, subtreeSum)
	
	// Second DFS: Re-root the tree to calculate answer for each node
	result[0] = subtreeSum[0]
	dfs2(0, -1, graph, subtreeSize, result, n)
	
	return result
}

// First DFS: Calculate subtree sizes and distance sums
func dfs1(node, parent int, graph [][]int, subtreeSize, subtreeSum []int) {
	subtreeSize[node] = 1
	subtreeSum[node] = 0
	
	for _, child := range graph[node] {
		if child != parent {
			dfs1(child, node, graph, subtreeSize, subtreeSum)
			subtreeSize[node] += subtreeSize[child]
			subtreeSum[node] += subtreeSum[child] + subtreeSize[child]
		}
	}
}

// Second DFS: Re-root the tree to calculate final answers
func dfs2(node, parent int, graph [][]int, subtreeSize, result []int, n int) {
	for _, child := range graph[node] {
		if child != parent {
			// When we move root from node to child:
			// - Nodes in child's subtree get 1 step closer
			// - Nodes outside child's subtree get 1 step farther
			result[child] = result[node] - subtreeSize[child] + (n - subtreeSize[child])
			dfs2(child, node, graph, subtreeSize, result, n)
		}
	}
}

// Solution 3: Alternative implementation using map for cleaner code
func sumOfDistancesInTreeAlt(n int, edges [][]int) []int {
	if n == 1 {
		return []int{0}
	}
	
	// Build adjacency list using map
	graph := make(map[int][]int)
	for i := 0; i < n; i++ {
		graph[i] = []int{}
	}
	
	for _, edge := range edges {
		a, b := edge[0], edge[1]
		graph[a] = append(graph[a], b)
		graph[b] = append(graph[b], a)
	}
	
	count := make([]int, n)  // count[i] = number of nodes in subtree i
	dist := make([]int, n)   // dist[i] = sum of distances from node i to all nodes in its subtree
	ans := make([]int, n)    // final answer
	
	// Post-order DFS
	var postOrder func(int, int)
	postOrder = func(node, parent int) {
		count[node] = 1
		for _, child := range graph[node] {
			if child != parent {
				postOrder(child, node)
				count[node] += count[child]
				dist[node] += dist[child] + count[child]
			}
		}
	}
	
	// Pre-order DFS
	var preOrder func(int, int)
	preOrder = func(node, parent int) {
		for _, child := range graph[node] {
			if child != parent {
				ans[child] = ans[node] - count[child] + (n - count[child])
				preOrder(child, node)
			}
		}
	}
	
	postOrder(0, -1)
	ans[0] = dist[0]
	preOrder(0, -1)
	
	return ans
}

func main() {
	// Test cases
	testCases := []struct {
		n     int
		edges [][]int
		expected []int
	}{
		{6, [][]int{{0,1},{0,2},{2,3},{2,4},{2,5}}, []int{8,12,6,10,10,10}},
		{1, [][]int{}, []int{0}},
		{2, [][]int{{1,0}}, []int{1,1}},
		{4, [][]int{{0,1},{1,2},{2,3}}, []int{6,4,4,6}},
		{5, [][]int{{0,1},{0,2},{0,3},{0,4}}, []int{4,7,7,7,7}},
		{7, [][]int{{0,1},{0,2},{1,3},{1,4},{2,5},{2,6}}, []int{10,11,11,16,16,16,16}},
	}
	
	fmt.Println("Testing Sum of Distances in Tree solutions:")
	fmt.Println()
	
	for i, tc := range testCases {
		fmt.Printf("Test Case %d:\n", i+1)
		fmt.Printf("Input: n=%d, edges=%v\n", tc.n, tc.edges)
		fmt.Printf("Expected: %v\n", tc.expected)
		
		// Test naive solution
		result1 := sumOfDistancesInTreeNaive(tc.n, tc.edges)
		fmt.Printf("Naive Solution: %v\n", result1)
		
		// Test optimized solution
		result2 := sumOfDistancesInTree(tc.n, tc.edges)
		fmt.Printf("Optimized Solution: %v\n", result2)
		
		// Test alternative solution
		result3 := sumOfDistancesInTreeAlt(tc.n, tc.edges)
		fmt.Printf("Alternative Solution: %v\n", result3)
		
		fmt.Println()
	}
}