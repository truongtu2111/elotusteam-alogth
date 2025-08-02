# Gray Code

## Problem Description

An **n-bit gray code sequence** is a sequence of 2^n integers where:

- Every integer is in the inclusive range [0, 2^n - 1]
- The first integer is 0
- An integer appears no more than once in the sequence
- The binary representation of every pair of adjacent integers differs by exactly one bit
- The binary representation of the first and last integers differs by exactly one bit

Given an integer `n`, return any valid **n-bit gray code sequence**.

## Examples

### Example 1:
```
Input: n = 2
Output: [0,1,3,2]
Explanation: 
The binary representation of [0,1,3,2] is [00,01,11,10].
- 00 and 01 differ by one bit
- 01 and 11 differ by one bit  
- 11 and 10 differ by one bit
- 10 and 00 differ by one bit

[0,2,3,1] is also a valid gray code sequence, whose binary representation is [00,10,11,01].
- 00 and 10 differ by one bit
- 10 and 11 differ by one bit
- 11 and 01 differ by one bit
- 01 and 00 differ by one bit
```

### Example 2:
```
Input: n = 1
Output: [0,1]
```

### Example 3:
```
Input: n = 3
Output: [0,1,3,2,6,7,5,4]
Explanation:
Binary: [000,001,011,010,110,111,101,100]
Each adjacent pair differs by exactly one bit.
```

## Solutions Implemented

### 1. Recursive Approach (Original)

**Algorithm:**
- **Base Case:** For n=1, return [0,1]
- **Recursive Step:** For n>1:
  1. Get the (n-1)-bit Gray code sequence
  2. Append the original sequence
  3. Append the mirrored sequence with MSB set

**Time Complexity:** O(2^n)  
**Space Complexity:** O(n × 2^n) due to recursion stack and intermediate arrays

**Implementation:** Original recursive version (replaced by optimized versions)

### 2. Iterative Approach - Optimized

**Algorithm:**
Uses the same mirroring principle but iteratively:

1. **Initialize:** Start with [0] and build level by level
2. **For each bit level i (0 to n-1):**
   - Current sequence has length 2^i
   - Mirror the existing sequence in reverse order
   - Add 2^i to each mirrored element (set the MSB)
   - Append to result

**Key Optimizations:**
- Eliminates recursion overhead
- Pre-allocates result array with exact capacity
- In-place mirroring without extra memory
- Direct bit manipulation for MSB setting

**Time Complexity:** O(2^n)  
**Space Complexity:** O(2^n) - only the result array

**Implementation:** `grayCode()`

### 3. Mathematical Formula Approach - Most Efficient

**Algorithm:**
Uses the direct mathematical formula for Gray code:

```
G(i) = i XOR (i >> 1)
```

Where `i` is the position in the sequence (0 to 2^n - 1).

**Mathematical Insight:**
The Gray code of any number can be computed directly by XORing the number with its right-shifted version. This eliminates the need for any mirroring or building sequences.

**Time Complexity:** O(2^n)  
**Space Complexity:** O(2^n) - only the result array

**Implementation:** `grayCodeFormula()`

## Performance Comparison

Benchmark results on different values of n:

```
BenchmarkGrayCodeIterative/n=5-8         8145214               132.7 ns/op
BenchmarkGrayCodeIterative/n=10-8          348004              3587 ns/op
BenchmarkGrayCodeIterative/n=15-8           14908             75710 ns/op
BenchmarkGrayCodeFormula/n=5-8           10343102               110.5 ns/op
BenchmarkGrayCodeFormula/n=10-8            411723              2716 ns/op
BenchmarkGrayCodeFormula/n=15-8             21092             69455 ns/op
```

**Performance Analysis:**
- **Formula Approach:** ~15-20% faster across all test sizes
- **Iterative Approach:** Good balance of readability and performance
- **Both approaches:** Scale linearly with O(2^n) complexity
- **Memory allocation:** Identical patterns for both optimized methods

## Key Concepts

### Gray Code Properties
1. **Hamming Distance:** Adjacent elements differ by exactly 1 bit
2. **Cyclic Property:** Last and first elements also differ by 1 bit
3. **Reflection Property:** Can be built by mirroring smaller Gray codes
4. **Mathematical Formula:** Direct computation using XOR operations

### Mirroring Technique
The core construction method:
1. Take existing n-bit Gray code: [G₀, G₁, ..., G₂ⁿ⁻¹]
2. Create (n+1)-bit Gray code: [G₀, G₁, ..., G₂ⁿ⁻¹, G₂ⁿ⁻¹+2ⁿ, ..., G₁+2ⁿ, G₀+2ⁿ]
3. The mirrored part ensures the transition property is maintained

### Binary Reflected Gray Code (BRGC)
The implemented solution generates the standard Binary Reflected Gray Code:
- Most commonly used Gray code variant
- Optimal for minimizing switching in digital circuits
- Used in rotary encoders and error correction

## Applications

### Real-world Uses:
- **Digital Circuits:** Minimize switching noise in counters
- **Rotary Encoders:** Reduce errors during position sensing
- **Genetic Algorithms:** Efficient mutation operations
- **Image Processing:** Dithering and error diffusion
- **Telecommunications:** Error correction codes

## Test Cases

The implementation includes comprehensive test coverage:

### Functional Tests:
- Basic cases (n=0,1,2,3) with expected outputs
- Both implementation approaches validation
- Equivalence testing between iterative and formula methods
- Edge cases and larger values (n=10)

### Property Validation:
- Correct sequence length (2^n elements)
- Range validation [0, 2^n-1]
- Starts with zero
- No duplicate values
- Adjacent elements differ by exactly 1 bit
- Cyclic property (last→first differs by 1 bit)

### Performance Tests:
- Benchmark comparisons between approaches
- Scalability testing for different n values
- Memory allocation pattern analysis

## Running the Code

```bash
# Run the main program with examples and verification
go run main.go

# Run all unit tests with verbose output
go test -v

# Run benchmark tests to compare performance
go test -bench=.

# Run tests with coverage report
go test -cover

# Test specific functions
go test -run TestGrayCodeProperties
go test -run TestBothApproachesEquivalent
```

## Constraints

- `1 <= n <= 16`
- The solution must return a valid Gray code sequence
- All Gray code properties must be satisfied
- Efficient memory usage for large n values

## Time and Space Complexity Summary

| Solution | Time Complexity | Space Complexity | Recursion | Memory Allocations | Best Use Case |
|----------|----------------|------------------|-----------|-------------------|---------------|
| Original Recursive | O(2^n) | O(n × 2^n) | Yes | Multiple arrays | Educational purposes |
| Iterative Optimized | O(2^n) | O(2^n) | No | Single array | Production code, balanced approach |
| Mathematical Formula | O(2^n) | O(2^n) | No | Single array | Maximum performance, direct computation |

## Algorithm Visualization

### Building Gray Code for n=3:

```
n=1: [0, 1]
     Binary: [0, 1]

n=2: [0, 1] + [3, 2]  (mirror [1,0] and add 2¹)
     Binary: [00, 01, 11, 10]

n=3: [0, 1, 3, 2] + [6, 7, 5, 4]  (mirror [2,3,1,0] and add 2²)
     Binary: [000, 001, 011, 010, 110, 111, 101, 100]
```

### Formula Approach for n=3:

```
i=0: 0 XOR (0>>1) = 0 XOR 0 = 0 → 000
i=1: 1 XOR (1>>1) = 1 XOR 0 = 1 → 001  
i=2: 2 XOR (2>>1) = 2 XOR 1 = 3 → 011
i=3: 3 XOR (3>>1) = 3 XOR 1 = 2 → 010
i=4: 4 XOR (4>>1) = 4 XOR 2 = 6 → 110
i=5: 5 XOR (5>>1) = 5 XOR 2 = 7 → 111
i=6: 6 XOR (6>>1) = 6 XOR 3 = 5 → 101
i=7: 7 XOR (7>>1) = 7 XOR 3 = 4 → 100
```

Result: [0, 1, 3, 2, 6, 7, 5, 4] ✓