package main

import (
	"fmt"
)

// Solution 1: Brute Force approach - O(n*m*min(n,m)) time complexity
// Check all possible subarrays and find the maximum length common subarray
func findLengthBruteForce(nums1 []int, nums2 []int) int {
	n1, n2 := len(nums1), len(nums2)
	maxLen := 0
	
	// Try all starting positions in nums1
	for i := 0; i < n1; i++ {
		// Try all starting positions in nums2
		for j := 0; j < n2; j++ {
			// Find the length of common subarray starting at i and j
			length := 0
			for k := 0; i+k < n1 && j+k < n2 && nums1[i+k] == nums2[j+k]; k++ {
				length++
			}
			if length > maxLen {
				maxLen = length
			}
		}
	}
	
	return maxLen
}

// Solution 2: Dynamic Programming approach - O(n*m) time, O(n*m) space
// Use 2D DP table where dp[i][j] represents length of common subarray ending at nums1[i-1] and nums2[j-1]
func findLength(nums1 []int, nums2 []int) int {
	n1, n2 := len(nums1), len(nums2)
	if n1 == 0 || n2 == 0 {
		return 0
	}
	
	// Create DP table
	dp := make([][]int, n1+1)
	for i := range dp {
		dp[i] = make([]int, n2+1)
	}
	
	maxLen := 0
	
	// Fill the DP table
	for i := 1; i <= n1; i++ {
		for j := 1; j <= n2; j++ {
			if nums1[i-1] == nums2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if dp[i][j] > maxLen {
					maxLen = dp[i][j]
				}
			} else {
				dp[i][j] = 0
			}
		}
	}
	
	return maxLen
}

// Solution 3: Space Optimized DP - O(n*m) time, O(min(n,m)) space
// Since we only need the previous row, we can optimize space to O(min(n,m))
func findLengthOptimized(nums1 []int, nums2 []int) int {
	n1, n2 := len(nums1), len(nums2)
	if n1 == 0 || n2 == 0 {
		return 0
	}
	
	// Ensure nums1 is the shorter array for space optimization
	if n1 > n2 {
		nums1, nums2 = nums2, nums1
		n1, n2 = n2, n1
	}
	
	// Use only two rows instead of full 2D array
	prev := make([]int, n1+1)
	curr := make([]int, n1+1)
	maxLen := 0
	
	for j := 1; j <= n2; j++ {
		for i := 1; i <= n1; i++ {
			if nums1[i-1] == nums2[j-1] {
				curr[i] = prev[i-1] + 1
				if curr[i] > maxLen {
					maxLen = curr[i]
				}
			} else {
				curr[i] = 0
			}
		}
		// Swap prev and curr for next iteration
		prev, curr = curr, prev
		// Clear curr for next use
		for i := range curr {
			curr[i] = 0
		}
	}
	
	return maxLen
}

// Solution 4: Rolling Hash approach - O(n*m*log(min(n,m))) time, O(n+m) space
// Use binary search on answer length and rolling hash for fast substring comparison
func findLengthRollingHash(nums1 []int, nums2 []int) int {
	n1, n2 := len(nums1), len(nums2)
	if n1 == 0 || n2 == 0 {
		return 0
	}
	
	// Binary search on the answer
	left, right := 0, min(n1, n2)
	result := 0
	
	for left <= right {
		mid := (left + right) / 2
		if hasCommonSubarray(nums1, nums2, mid) {
			result = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	
	return result
}

// Helper function to check if there exists a common subarray of given length
func hasCommonSubarray(nums1, nums2 []int, length int) bool {
	if length == 0 {
		return true
	}
	
	const base = 101
	const mod = 1000000007
	
	// Calculate base^(length-1) % mod
	basePow := 1
	for i := 0; i < length-1; i++ {
		basePow = (basePow * base) % mod
	}
	
	// Get all hashes of subarrays of given length in nums1
	hashes1 := make(map[int]bool)
	hash := 0
	
	// Calculate initial hash for nums1
	for i := 0; i < length; i++ {
		hash = (hash*base + nums1[i]) % mod
	}
	hashes1[hash] = true
	
	// Rolling hash for remaining subarrays in nums1
	for i := length; i < len(nums1); i++ {
		hash = (hash - (nums1[i-length]*basePow)%mod + mod) % mod
		hash = (hash*base + nums1[i]) % mod
		hashes1[hash] = true
	}
	
	// Check if any subarray hash in nums2 matches
	hash = 0
	// Calculate initial hash for nums2
	for i := 0; i < length; i++ {
		hash = (hash*base + nums2[i]) % mod
	}
	if hashes1[hash] {
		return true
	}
	
	// Rolling hash for remaining subarrays in nums2
	for i := length; i < len(nums2); i++ {
		hash = (hash - (nums2[i-length]*basePow)%mod + mod) % mod
		hash = (hash*base + nums2[i]) % mod
		if hashes1[hash] {
			return true
		}
	}
	
	return false
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Test cases
	testCases := []struct {
		nums1    []int
		nums2    []int
		expected int
	}{
		{[]int{1, 2, 3, 2, 1}, []int{3, 2, 1, 4, 7}, 3},
		{[]int{0, 0, 0, 0, 0}, []int{0, 0, 0, 0, 0}, 5},
		{[]int{1, 2, 3}, []int{4, 5, 6}, 0},
		{[]int{}, []int{1, 2, 3}, 0},
		{[]int{1, 2, 3}, []int{}, 0},
		{[]int{1}, []int{1}, 1},
		{[]int{1, 0, 1, 0, 1}, []int{1, 1, 1, 1, 1}, 1},
		{[]int{0, 1, 1, 1, 1}, []int{1, 0, 1, 0, 1}, 2},
	}
	
	fmt.Println("Testing Maximum Length of Repeated Subarray solutions:")
	fmt.Println()
	
	for i, tc := range testCases {
		fmt.Printf("Test Case %d:\n", i+1)
		fmt.Printf("Input: nums1=%v, nums2=%v\n", tc.nums1, tc.nums2)
		fmt.Printf("Expected: %d\n", tc.expected)
		
		// Test brute force solution
		result1 := findLengthBruteForce(tc.nums1, tc.nums2)
		fmt.Printf("Brute Force Solution: %d\n", result1)
		
		// Test DP solution
		result2 := findLength(tc.nums1, tc.nums2)
		fmt.Printf("DP Solution: %d\n", result2)
		
		// Test optimized DP solution
		result3 := findLengthOptimized(tc.nums1, tc.nums2)
		fmt.Printf("Optimized DP Solution: %d\n", result3)
		
		// Test rolling hash solution
		result4 := findLengthRollingHash(tc.nums1, tc.nums2)
		fmt.Printf("Rolling Hash Solution: %d\n", result4)
		
		// Verify all solutions match expected result
		if result1 == tc.expected && result2 == tc.expected && result3 == tc.expected && result4 == tc.expected {
			fmt.Printf("✓ All solutions correct!\n")
		} else {
			fmt.Printf("✗ Some solutions incorrect!\n")
		}
		
		fmt.Println()
	}
}