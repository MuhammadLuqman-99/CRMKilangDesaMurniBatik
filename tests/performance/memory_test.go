// Package performance contains memory profiling tests for the CRM system.
package performance

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Memory Profiling Tests
// ============================================================================

// MemoryProfile contains memory profiling results.
type MemoryProfile struct {
	Name           string
	AllocBytes     uint64
	TotalAllocs    uint64
	HeapInUse      uint64
	HeapIdle       uint64
	HeapReleased   uint64
	HeapObjects    uint64
	GCPauseTotal   time.Duration
	GCPauseAvg     time.Duration
	GCNumCollects  uint32
	StackInUse     uint64
	MSpanInUse     uint64
	MCacheInUse    uint64
}

// CaptureMemoryProfile captures current memory statistics.
func CaptureMemoryProfile(name string) *MemoryProfile {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var gcPauseTotal time.Duration
	numPauses := m.NumGC
	if numPauses > 256 {
		numPauses = 256 // Only last 256 pauses are stored
	}
	for i := uint32(0); i < numPauses; i++ {
		gcPauseTotal += time.Duration(m.PauseNs[i])
	}

	var gcPauseAvg time.Duration
	if numPauses > 0 {
		gcPauseAvg = gcPauseTotal / time.Duration(numPauses)
	}

	return &MemoryProfile{
		Name:          name,
		AllocBytes:    m.Alloc,
		TotalAllocs:   m.TotalAlloc,
		HeapInUse:     m.HeapInuse,
		HeapIdle:      m.HeapIdle,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		GCPauseTotal:  gcPauseTotal,
		GCPauseAvg:    gcPauseAvg,
		GCNumCollects: m.NumGC,
		StackInUse:    m.StackInuse,
		MSpanInUse:    m.MSpanInuse,
		MCacheInUse:   m.MCacheInuse,
	}
}

// Print prints the memory profile.
func (p *MemoryProfile) Print() {
	fmt.Printf("\n=== Memory Profile: %s ===\n", p.Name)
	fmt.Printf("Heap Allocated:     %s\n", formatBytes(p.AllocBytes))
	fmt.Printf("Total Allocations:  %s\n", formatBytes(p.TotalAllocs))
	fmt.Printf("Heap In Use:        %s\n", formatBytes(p.HeapInUse))
	fmt.Printf("Heap Idle:          %s\n", formatBytes(p.HeapIdle))
	fmt.Printf("Heap Released:      %s\n", formatBytes(p.HeapReleased))
	fmt.Printf("Heap Objects:       %d\n", p.HeapObjects)
	fmt.Printf("GC Collections:     %d\n", p.GCNumCollects)
	fmt.Printf("GC Pause Total:     %v\n", p.GCPauseTotal)
	fmt.Printf("GC Pause Average:   %v\n", p.GCPauseAvg)
	fmt.Printf("Stack In Use:       %s\n", formatBytes(p.StackInUse))
	fmt.Printf("MSpan In Use:       %s\n", formatBytes(p.MSpanInUse))
	fmt.Printf("MCache In Use:      %s\n", formatBytes(p.MCacheInUse))
}

// Diff calculates the difference between two memory profiles.
func (p *MemoryProfile) Diff(other *MemoryProfile) *MemoryProfileDiff {
	return &MemoryProfileDiff{
		Name:             fmt.Sprintf("%s -> %s", p.Name, other.Name),
		AllocDelta:       int64(other.AllocBytes) - int64(p.AllocBytes),
		TotalAllocDelta:  int64(other.TotalAllocs) - int64(p.TotalAllocs),
		HeapInUseDelta:   int64(other.HeapInUse) - int64(p.HeapInUse),
		HeapObjectsDelta: int64(other.HeapObjects) - int64(p.HeapObjects),
		GCCollectsDelta:  int32(other.GCNumCollects) - int32(p.GCNumCollects),
	}
}

// MemoryProfileDiff represents the difference between two profiles.
type MemoryProfileDiff struct {
	Name             string
	AllocDelta       int64
	TotalAllocDelta  int64
	HeapInUseDelta   int64
	HeapObjectsDelta int64
	GCCollectsDelta  int32
}

// Print prints the memory profile difference.
func (d *MemoryProfileDiff) Print() {
	fmt.Printf("\n=== Memory Diff: %s ===\n", d.Name)
	fmt.Printf("Alloc Delta:        %s\n", formatBytesDelta(d.AllocDelta))
	fmt.Printf("Total Alloc Delta:  %s\n", formatBytesDelta(d.TotalAllocDelta))
	fmt.Printf("Heap In Use Delta:  %s\n", formatBytesDelta(d.HeapInUseDelta))
	fmt.Printf("Heap Objects Delta: %+d\n", d.HeapObjectsDelta)
	fmt.Printf("GC Collections:     %+d\n", d.GCCollectsDelta)
}

// TestMemoryBaseline establishes memory baseline.
func TestMemoryBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory tests in short mode")
	}

	// Force GC to get clean baseline
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	baseline := CaptureMemoryProfile("Baseline")
	baseline.Print()

	// Create test server
	server := NewTestServer()

	afterServer := CaptureMemoryProfile("After Server Creation")
	afterServer.Print()

	serverDiff := baseline.Diff(afterServer)
	serverDiff.Print()

	server.Close()

	// Force GC after cleanup
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	afterClose := CaptureMemoryProfile("After Server Close")
	afterClose.Print()

	closeDiff := afterServer.Diff(afterClose)
	closeDiff.Print()

	// Check for memory leaks
	leakThreshold := int64(10 * 1024 * 1024) // 10MB threshold
	if afterClose.AllocBytes-baseline.AllocBytes > uint64(leakThreshold) {
		t.Errorf("Potential memory leak: %s retained after cleanup",
			formatBytes(afterClose.AllocBytes-baseline.AllocBytes))
	}
}

// TestMemoryUnderLoad tests memory behavior under load.
func TestMemoryUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory load tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	client := NewLoadTestClient(server.Server.URL)

	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	before := CaptureMemoryProfile("Before Load")
	before.Print()

	// Run load test
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
					"code":   fmt.Sprintf("MEM-%d-%d", id, j),
					"name":   "Memory Test Customer",
					"status": "active",
				})
			}
		}(i)
	}
	wg.Wait()

	duringLoad := CaptureMemoryProfile("During Load")
	duringLoad.Print()

	loadDiff := before.Diff(duringLoad)
	loadDiff.Print()

	// Force GC and measure recovery
	runtime.GC()
	time.Sleep(time.Second)

	afterGC := CaptureMemoryProfile("After GC")
	afterGC.Print()

	gcRecoveryDiff := duringLoad.Diff(afterGC)
	gcRecoveryDiff.Print()

	// Memory should be released after GC
	if afterGC.HeapInUse > duringLoad.HeapInUse {
		t.Log("Warning: Heap usage increased after GC")
	}
}

// TestMemoryAllocationPatterns tests memory allocation patterns.
func TestMemoryAllocationPatterns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory allocation tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	client := NewLoadTestClient(server.Server.URL)

	patterns := []struct {
		name      string
		operation func()
		count     int
	}{
		{
			name: "Small Reads",
			operation: func() {
				client.DoRequest("GET", "/api/v1/customers", nil)
			},
			count: 1000,
		},
		{
			name: "Small Writes",
			operation: func() {
				client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
					"code":   fmt.Sprintf("ALLOC-%s", uuid.New().String()[:8]),
					"name":   "Allocation Test",
					"status": "active",
				})
			},
			count: 1000,
		},
		{
			name: "Mixed Operations",
			operation: func() {
				client.DoRequest("GET", "/api/v1/customers", nil)
				client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
					"code":   fmt.Sprintf("MIX-%s", uuid.New().String()[:8]),
					"name":   "Mixed Test",
					"status": "active",
				})
			},
			count: 500,
		},
	}

	fmt.Println("\n=== Memory Allocation Patterns ===")
	fmt.Printf("%-20s %15s %15s %15s %15s\n",
		"Pattern", "Allocs/Op", "Bytes/Op", "Total Alloc", "Objects")
	fmt.Println(strings.Repeat("-", 85))

	for _, pattern := range patterns {
		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		var before runtime.MemStats
		runtime.ReadMemStats(&before)

		for i := 0; i < pattern.count; i++ {
			pattern.operation()
		}

		var after runtime.MemStats
		runtime.ReadMemStats(&after)

		allocsPerOp := (after.Mallocs - before.Mallocs) / uint64(pattern.count)
		bytesPerOp := (after.TotalAlloc - before.TotalAlloc) / uint64(pattern.count)
		totalAlloc := after.TotalAlloc - before.TotalAlloc
		objectsCreated := after.HeapObjects - before.HeapObjects

		fmt.Printf("%-20s %15d %15s %15s %15d\n",
			pattern.name, allocsPerOp, formatBytes(bytesPerOp),
			formatBytes(totalAlloc), objectsCreated)
	}
}

// TestMemoryGCBehavior tests garbage collection behavior.
func TestMemoryGCBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GC behavior tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	client := NewLoadTestClient(server.Server.URL)

	fmt.Println("\n=== GC Behavior Analysis ===")

	// Measure GC pause times under different loads
	loads := []struct {
		name        string
		concurrency int
		operations  int
	}{
		{"Light Load", 10, 100},
		{"Medium Load", 50, 100},
		{"Heavy Load", 100, 100},
	}

	for _, load := range loads {
		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		var before runtime.MemStats
		runtime.ReadMemStats(&before)

		startTime := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < load.concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < load.operations; j++ {
					client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
						"code":   fmt.Sprintf("GC-%s", uuid.New().String()[:8]),
						"name":   "GC Test",
						"status": "active",
					})
				}
			}()
		}
		wg.Wait()

		elapsed := time.Since(startTime)

		var after runtime.MemStats
		runtime.ReadMemStats(&after)

		gcCount := after.NumGC - before.NumGC
		var gcPauseTotal time.Duration
		for i := before.NumGC; i < after.NumGC && i < 256; i++ {
			gcPauseTotal += time.Duration(after.PauseNs[i%256])
		}

		var gcPauseAvg time.Duration
		if gcCount > 0 {
			gcPauseAvg = gcPauseTotal / time.Duration(gcCount)
		}

		gcOverhead := float64(gcPauseTotal) / float64(elapsed) * 100

		fmt.Printf("\n%s:\n", load.name)
		fmt.Printf("  Duration:        %v\n", elapsed)
		fmt.Printf("  GC Collections:  %d\n", gcCount)
		fmt.Printf("  GC Pause Total:  %v\n", gcPauseTotal)
		fmt.Printf("  GC Pause Avg:    %v\n", gcPauseAvg)
		fmt.Printf("  GC Overhead:     %.2f%%\n", gcOverhead)
		fmt.Printf("  Memory Alloc:    %s\n", formatBytes(after.Alloc-before.Alloc))

		// GC overhead should be reasonable
		if gcOverhead > 5 { // 5% threshold
			t.Logf("Warning: High GC overhead (%.2f%%) under %s", gcOverhead, load.name)
		}
	}
}

// TestMemoryLeakDetection tests for memory leaks.
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	client := NewLoadTestClient(server.Server.URL)

	fmt.Println("\n=== Memory Leak Detection ===")

	// Run multiple cycles and check for consistent memory growth
	numCycles := 5
	operationsPerCycle := 500

	var memoryReadings []uint64

	for cycle := 1; cycle <= numCycles; cycle++ {
		// Run operations
		for i := 0; i < operationsPerCycle; i++ {
			client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
				"code":   fmt.Sprintf("LEAK-%d-%d", cycle, i),
				"name":   "Leak Detection Customer",
				"status": "active",
			})
			client.DoRequest("GET", "/api/v1/customers", nil)
		}

		// Force GC
		runtime.GC()
		time.Sleep(500 * time.Millisecond)

		// Record memory
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memoryReadings = append(memoryReadings, m.Alloc)

		fmt.Printf("Cycle %d: Heap Alloc = %s\n", cycle, formatBytes(m.Alloc))
	}

	// Analyze for leaks
	// Check if memory is growing consistently across cycles
	isLeaking := true
	for i := 1; i < len(memoryReadings); i++ {
		// Allow for some variance (10% growth per cycle is suspicious)
		threshold := float64(memoryReadings[i-1]) * 1.10
		if float64(memoryReadings[i]) < threshold {
			isLeaking = false
			break
		}
	}

	if isLeaking {
		growth := memoryReadings[len(memoryReadings)-1] - memoryReadings[0]
		t.Errorf("Potential memory leak detected: %s growth over %d cycles",
			formatBytes(growth), numCycles)
	} else {
		t.Log("No significant memory leak detected")
	}
}

// TestMemoryObjectPooling tests object pooling effectiveness.
func TestMemoryObjectPooling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping object pooling tests in short mode")
	}

	// Test sync.Pool effectiveness
	type TestObject struct {
		Data [1024]byte
	}

	pool := sync.Pool{
		New: func() interface{} {
			return &TestObject{}
		},
	}

	fmt.Println("\n=== Object Pooling Test ===")

	// Without pool
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for i := 0; i < 10000; i++ {
		obj := &TestObject{}
		_ = obj
	}

	var afterNoPool runtime.MemStats
	runtime.ReadMemStats(&afterNoPool)

	allocsWithoutPool := afterNoPool.Mallocs - before.Mallocs
	bytesWithoutPool := afterNoPool.TotalAlloc - before.TotalAlloc

	// With pool
	runtime.GC()
	runtime.ReadMemStats(&before)

	for i := 0; i < 10000; i++ {
		obj := pool.Get().(*TestObject)
		pool.Put(obj)
	}

	var afterWithPool runtime.MemStats
	runtime.ReadMemStats(&afterWithPool)

	allocsWithPool := afterWithPool.Mallocs - before.Mallocs
	bytesWithPool := afterWithPool.TotalAlloc - before.TotalAlloc

	fmt.Printf("\nWithout Pool:\n")
	fmt.Printf("  Allocations: %d\n", allocsWithoutPool)
	fmt.Printf("  Bytes:       %s\n", formatBytes(bytesWithoutPool))

	fmt.Printf("\nWith Pool:\n")
	fmt.Printf("  Allocations: %d\n", allocsWithPool)
	fmt.Printf("  Bytes:       %s\n", formatBytes(bytesWithPool))

	if allocsWithPool > 0 {
		reduction := float64(allocsWithoutPool-allocsWithPool) / float64(allocsWithoutPool) * 100
		fmt.Printf("\nAllocation Reduction: %.2f%%\n", reduction)
	}
}

// ============================================================================
// Memory Benchmarks
// ============================================================================

// BenchmarkMemoryAllocSmall benchmarks small memory allocations.
func BenchmarkMemoryAllocSmall(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	client := NewLoadTestClient(server.Server.URL)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.DoRequest("GET", "/api/v1/customers", nil)
	}
}

// BenchmarkMemoryAllocLarge benchmarks larger memory allocations.
func BenchmarkMemoryAllocLarge(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	server.Seed(10000, 0, 0)
	client := NewLoadTestClient(server.Server.URL)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.DoRequest("GET", "/api/v1/customers", nil)
	}
}

// BenchmarkMemoryCreateCustomer benchmarks memory allocation for customer creation.
func BenchmarkMemoryCreateCustomer(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	client := NewLoadTestClient(server.Server.URL)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
			"code":   fmt.Sprintf("ALLOC-%d", i),
			"name":   "Allocation Benchmark",
			"status": "active",
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatBytesDelta(delta int64) string {
	if delta >= 0 {
		return "+" + formatBytes(uint64(delta))
	}
	return "-" + formatBytes(uint64(-delta))
}
