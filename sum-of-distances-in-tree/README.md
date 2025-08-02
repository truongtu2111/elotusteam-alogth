# Sum of Distances in Tree

## Problem Description

There is an undirected connected tree with `n` nodes labeled from `0` to `n - 1` and `n - 1` edges.

You are given the integer `n` and the array `edges` where `edges[i] = [ai, bi]` indicates that there is an edge between nodes `ai` and `bi` in the tree.

Return an array `answer` of length `n` where `answer[i]` is the sum of the distances between the `ith` node in the tree and all other nodes.

## Examples

### Example 1:
```
Input: n = 6, edges = [[0,1],[0,2],[2,3],[2,4],[2,5]]
Output: [8,12,6,10,10,10]
Explanation: The tree is shown above.
We can see that dist(0,1) + dist(0,2) + dist(0,3) + dist(0,4) + dist(0,5)
equals 1 + 1 + 2 + 2 + 2 = 8.
Hence, answer[0] = 8, and so on.
```

### Example 2:
```
Input: n = 1, edges = []
Output: [0]
```

### Example 3:
```
Input: n = 2, edges = [[1,0]]
Output: [1,1]
```

## Solutions Implemented

### 1. Naive Approach - O(n²) Time Complexity

**Algorithm:**
- For each node, perform BFS to calculate distances to all other nodes
- Sum up all distances for each starting node
- Time Complexity: O(n²)
- Space Complexity: O(n)

**Implementation:** `sumOfDistancesInTreeNaive()`

### 2. Optimized Approach - O(n) Time Complexity

**Algorithm:**
This solution uses a two-pass DFS with dynamic programming:

**First Pass (Post-order DFS):**
- Calculate subtree sizes for each node
- Calculate sum of distances from each node to all nodes in its subtree

**Second Pass (Pre-order DFS):**
- Re-root the tree to calculate the final answer for each node
- When moving root from parent to child:
  - Nodes in child's subtree get 1 step closer
  - Nodes outside child's subtree get 1 step farther

**Key Insight:**
When we move the root from node `u` to its child `v`:
- `answer[v] = answer[u] - subtreeSize[v] + (n - subtreeSize[v])`

**Time Complexity:** O(n)
**Space Complexity:** O(n)

**Implementation:** `sumOfDistancesInTree()`

### 3. Alternative Implementation - O(n) Time Complexity

**Algorithm:**
Similar to the optimized approach but uses:
- Map-based adjacency list for cleaner code
- Closure functions for DFS traversals
- Slightly different variable naming for clarity

**Implementation:** `sumOfDistancesInTreeAlt()`

## Performance Comparison

Benchmark results on a linear tree with 100 nodes:

```
BenchmarkSumOfDistancesInTreeNaive-8        2436            432225 ns/op
BenchmarkSumOfDistancesInTree-8            99218             11868 ns/op
BenchmarkSumOfDistancesInTreeAlt-8         37870             31583 ns/op
```

**Performance Analysis:**
- **Naive Solution:** ~432,225 ns/op
- **Optimized Solution:** ~11,868 ns/op (**36x faster**)
- **Alternative Solution:** ~31,583 ns/op (**13x faster**)

The optimized solution shows significant performance improvement, especially for larger trees.

## Key Concepts

### Tree Re-rooting Technique
The core optimization uses the "re-rooting" technique:
1. First, solve the problem for one root (typically node 0)
2. Then, use the relationship between parent and child solutions to compute answers for all other nodes

### Dynamic Programming on Trees
The solution demonstrates classic DP on trees:
- **State:** `dp[node]` = sum of distances from node to all nodes in its subtree
- **Transition:** Combine results from children subtrees
- **Re-rooting:** Transform solution when changing root

## Test Cases

The implementation includes comprehensive test cases:
- Basic examples from the problem statement
- Edge cases (single node, two nodes)
- Different tree structures (linear, star, balanced binary tree)
- Large trees for performance testing
- Consistency tests between all implementations

## Running the Code

```bash
# Run the main program with test cases
go run main.go

# Run all unit tests
go test -v

# Run benchmark tests
go test -bench=.

# Run tests with coverage
go test -cover
```

## Constraints

- `1 <= n <= 3 * 10^4`
- `edges.length == n - 1`
- `edges[i].length == 2`
- `0 <= ai, bi < n`
- `ai != bi`
- The given input represents a valid tree

## Time and Space Complexity Summary

| Solution | Time Complexity | Space Complexity | Best Use Case |
|----------|----------------|------------------|---------------|
| Naive | O(n²) | O(n) | Small trees, educational purposes |
| Optimized | O(n) | O(n) | Production code, large trees |
| Alternative | O(n) | O(n) | When code readability is prioritized |