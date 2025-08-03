package performance

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// BenchmarkPasswordHashing benchmarks bcrypt password hashing performance
func BenchmarkPasswordHashing(b *testing.B) {
	password := "testpassword123"
	costs := []int{bcrypt.DefaultCost, bcrypt.DefaultCost + 1, bcrypt.DefaultCost + 2}

	for _, cost := range costs {
		b.Run(fmt.Sprintf("Cost%d", cost), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := bcrypt.GenerateFromPassword([]byte(password), cost)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkPasswordVerification benchmarks bcrypt password verification
func BenchmarkPasswordVerification(b *testing.B) {
	password := "testpassword123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := bcrypt.CompareHashAndPassword(hash, []byte(password))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONMarshaling benchmarks JSON marshaling performance
func BenchmarkJSONMarshaling(b *testing.B) {
	user := struct {
		ID       string    `json:"id"`
		Name     string    `json:"name"`
		Email    string    `json:"email"`
		Created  time.Time `json:"created"`
		Profile  map[string]interface{} `json:"profile"`
	}{
		ID:      "user-123",
		Name:    "John Doe",
		Email:   "john@example.com",
		Created: time.Now(),
		Profile: map[string]interface{}{
			"age":      30,
			"location": "New York",
			"interests": []string{"coding", "music", "travel"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONUnmarshaling benchmarks JSON unmarshaling performance
func BenchmarkJSONUnmarshaling(b *testing.B) {
	jsonData := `{
		"id": "user-123",
		"name": "John Doe",
		"email": "john@example.com",
		"created": "2023-01-01T00:00:00Z",
		"profile": {
			"age": 30,
			"location": "New York",
			"interests": ["coding", "music", "travel"]
		}
	}`

	var user struct {
		ID       string                 `json:"id"`
		Name     string                 `json:"name"`
		Email    string                 `json:"email"`
		Created  time.Time              `json:"created"`
		Profile  map[string]interface{} `json:"profile"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := json.Unmarshal([]byte(jsonData), &user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	sizes := []int{1024, 4096, 16384, 65536} // 1KB, 4KB, 16KB, 64KB

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%dB", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				data := make([]byte, size)
				// Prevent compiler optimization
				_ = data[0]
			}
		})
	}
}

// BenchmarkSliceOperations benchmarks slice operations
func BenchmarkSliceOperations(b *testing.B) {
	b.Run("Append", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var slice []int
			for j := 0; j < 1000; j++ {
				slice = append(slice, j)
			}
		}
	})

	b.Run("PreAllocated", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := make([]int, 0, 1000)
			for j := 0; j < 1000; j++ {
				slice = append(slice, j)
			}
		}
	})

	b.Run("DirectIndex", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := make([]int, 1000)
			for j := 0; j < 1000; j++ {
				slice[j] = j
			}
		}
	})
}

// BenchmarkMapOperations benchmarks map operations
func BenchmarkMapOperations(b *testing.B) {
	b.Run("StringKeys", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m := make(map[string]int)
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key%d", j)
				m[key] = j
			}
		}
	})

	b.Run("IntKeys", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m := make(map[int]int)
			for j := 0; j < 1000; j++ {
				m[j] = j
			}
		}
	})

	b.Run("PreAllocated", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m := make(map[string]int, 1000)
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key%d", j)
				m[key] = j
			}
		}
	})
}

// BenchmarkStringOperations benchmarks string operations
func BenchmarkStringOperations(b *testing.B) {
	b.Run("Concatenation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result string
			for j := 0; j < 100; j++ {
				result += fmt.Sprintf("item%d", j)
			}
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var builder bytes.Buffer
			for j := 0; j < 100; j++ {
				builder.WriteString(fmt.Sprintf("item%d", j))
			}
			_ = builder.String()
		}
	})
}

// BenchmarkCryptoOperations benchmarks cryptographic operations
func BenchmarkCryptoOperations(b *testing.B) {
	sizes := []int{1024, 4096, 16384} // Different data sizes

	for _, size := range sizes {
		b.Run(fmt.Sprintf("RandomGeneration%dB", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				data := make([]byte, size)
				_, err := rand.Read(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkGoroutineCreation benchmarks goroutine creation and cleanup
func BenchmarkGoroutineCreation(b *testing.B) {
	b.Run("SimpleGoroutine", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			done := make(chan bool)
			go func() {
				done <- true
			}()
			<-done
		}
	})

	b.Run("WorkerPool", func(b *testing.B) {
		workers := 10
		jobs := make(chan int, b.N)
		results := make(chan int, b.N)

		// Start workers
		for w := 0; w < workers; w++ {
			go func() {
				for job := range jobs {
					results <- job * 2
				}
			}()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jobs <- i
		}
		close(jobs)

		for i := 0; i < b.N; i++ {
			<-results
		}
	})
}

// BenchmarkMemoryUsage measures memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("MemoryIntensive", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Allocate large data structures
			data := make([][]byte, 1000)
			for j := range data {
				data[j] = make([]byte, 1024)
			}
			// Force usage to prevent optimization
			_ = data[0][0]
		}
		b.StopTimer()

		runtime.GC()
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
		b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
	})
}

// BenchmarkConcurrentAccess benchmarks concurrent access patterns
func BenchmarkConcurrentAccess(b *testing.B) {
	b.Run("ChannelCommunication", func(b *testing.B) {
		ch := make(chan int, 100)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				select {
				case ch <- 1:
				case <-ch:
				default:
				}
			}
		})
	})
}