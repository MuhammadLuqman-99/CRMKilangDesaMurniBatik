// Package performance contains query optimization benchmarks.
package performance

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Query Optimization Benchmarks
// ============================================================================

// QueryBenchmarkResult contains benchmark results for a query.
type QueryBenchmarkResult struct {
	QueryName     string
	Iterations    int
	TotalTime     time.Duration
	MinTime       time.Duration
	MaxTime       time.Duration
	AvgTime       time.Duration
	P50Time       time.Duration
	P95Time       time.Duration
	P99Time       time.Duration
	OpsPerSecond  float64
}

// TestQueryOptimization tests various query patterns for optimization opportunities.
func TestQueryOptimization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping query optimization tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	// Seed with various data sizes
	dataSizes := []struct {
		name      string
		customers int
		leads     int
		users     int
	}{
		{"Small", 100, 100, 10},
		{"Medium", 1000, 1000, 100},
		{"Large", 10000, 10000, 1000},
	}

	for _, size := range dataSizes {
		t.Run(size.name+"Dataset", func(t *testing.T) {
			// Clear and reseed
			server.DataStore.customers = make(map[string]*CustomerData)
			server.DataStore.leads = make(map[string]*LeadData)
			server.DataStore.users = make(map[string]*UserData)
			server.Seed(size.customers, size.leads, size.users)

			client := NewLoadTestClient(server.Server.URL)

			// Test list operations
			results := []QueryBenchmarkResult{}

			// List all customers
			result := benchmarkQuery("List Customers", 100, func() error {
				_, err := client.DoRequest("GET", "/api/v1/customers", nil)
				return err
			})
			results = append(results, result)

			// List all leads
			result = benchmarkQuery("List Leads", 100, func() error {
				_, err := client.DoRequest("GET", "/api/v1/leads", nil)
				return err
			})
			results = append(results, result)

			// Get single customer
			var customerID string
			for id := range server.DataStore.customers {
				customerID = id
				break
			}
			result = benchmarkQuery("Get Customer by ID", 100, func() error {
				_, err := client.DoRequest("GET", "/api/v1/customers/"+customerID, nil)
				return err
			})
			results = append(results, result)

			// Get single lead
			var leadID string
			for id := range server.DataStore.leads {
				leadID = id
				break
			}
			result = benchmarkQuery("Get Lead by ID", 100, func() error {
				_, err := client.DoRequest("GET", "/api/v1/leads/"+leadID, nil)
				return err
			})
			results = append(results, result)

			// Print results
			fmt.Printf("\n=== Query Benchmark Results (%s Dataset: %d customers, %d leads) ===\n",
				size.name, size.customers, size.leads)
			printQueryBenchmarkResults(results)
		})
	}
}

// TestQueryScalability tests how query performance scales with data size.
func TestQueryScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping query scalability tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	client := NewLoadTestClient(server.Server.URL)

	dataSizes := []int{100, 500, 1000, 5000, 10000}
	results := make(map[string][]QueryBenchmarkResult)

	for _, size := range dataSizes {
		// Clear and reseed
		server.DataStore.customers = make(map[string]*CustomerData)
		server.Seed(size, 0, 0)

		// Benchmark list operation
		result := benchmarkQuery(fmt.Sprintf("List %d Customers", size), 50, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})
		results["List"] = append(results["List"], result)

		// Benchmark single get
		var customerID string
		for id := range server.DataStore.customers {
			customerID = id
			break
		}
		result = benchmarkQuery(fmt.Sprintf("Get from %d Customers", size), 50, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers/"+customerID, nil)
			return err
		})
		results["Get"] = append(results["Get"], result)
	}

	// Print scalability analysis
	fmt.Println("\n=== Query Scalability Analysis ===")

	for queryType, queryResults := range results {
		fmt.Printf("\n%s Operation:\n", queryType)
		fmt.Printf("%-15s %15s %15s %15s\n", "Data Size", "Avg Time", "P95 Time", "Ops/Sec")
		fmt.Println(strings.Repeat("-", 65))

		for i, result := range queryResults {
			fmt.Printf("%-15d %15v %15v %15.2f\n",
				dataSizes[i], result.AvgTime, result.P95Time, result.OpsPerSecond)
		}

		// Calculate scalability factor
		if len(queryResults) >= 2 {
			firstAvg := queryResults[0].AvgTime.Seconds()
			lastAvg := queryResults[len(queryResults)-1].AvgTime.Seconds()
			dataFactor := float64(dataSizes[len(dataSizes)-1]) / float64(dataSizes[0])
			timeFactor := lastAvg / firstAvg

			fmt.Printf("\nScalability: Data increased %.1fx, time increased %.1fx\n",
				dataFactor, timeFactor)

			if timeFactor > dataFactor {
				fmt.Println("WARNING: Query time growing faster than data size (consider optimization)")
			} else {
				fmt.Println("OK: Query scales well with data size")
			}
		}
	}
}

// TestConcurrentQueryPerformance tests query performance under concurrent access.
func TestConcurrentQueryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent query tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(5000, 5000, 500)

	client := NewLoadTestClient(server.Server.URL)

	concurrencyLevels := []int{1, 5, 10, 25, 50}

	fmt.Println("\n=== Concurrent Query Performance ===")
	fmt.Printf("%-15s %15s %15s %15s %15s\n",
		"Concurrency", "Avg Time", "P95 Time", "P99 Time", "Throughput")
	fmt.Println(strings.Repeat("-", 80))

	for _, concurrency := range concurrencyLevels {
		result := benchmarkConcurrentQuery("List Customers", concurrency, 100, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})

		fmt.Printf("%-15d %15v %15v %15v %15.2f req/s\n",
			concurrency, result.AvgTime, result.P95Time, result.P99Time, result.OpsPerSecond)
	}
}

// TestReadWriteQueryMix tests performance of mixed read/write queries.
func TestReadWriteQueryMix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping read/write mix tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)

	client := NewLoadTestClient(server.Server.URL)

	// Test different read/write ratios
	ratios := []struct {
		name       string
		readRatio  int
		writeRatio int
	}{
		{"90% Read / 10% Write", 90, 10},
		{"70% Read / 30% Write", 70, 30},
		{"50% Read / 50% Write", 50, 50},
	}

	// Get some IDs for read operations
	var customerIDs []string
	for id := range server.DataStore.customers {
		customerIDs = append(customerIDs, id)
		if len(customerIDs) >= 100 {
			break
		}
	}

	fmt.Println("\n=== Read/Write Query Mix Performance ===")

	for _, ratio := range ratios {
		t.Run(ratio.name, func(t *testing.T) {
			var times []time.Duration
			var errors int
			totalOps := 500
			counter := 0

			for i := 0; i < totalOps; i++ {
				counter++
				var start time.Time
				var err error

				if counter%100 < ratio.readRatio {
					// Read operation
					idx := i % len(customerIDs)
					start = time.Now()
					_, err = client.DoRequest("GET", "/api/v1/customers/"+customerIDs[idx], nil)
				} else {
					// Write operation
					start = time.Now()
					_, err = client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
						"code":   fmt.Sprintf("MIX-%s", uuid.New().String()[:8]),
						"name":   "Mix Test Customer",
						"status": "active",
					})
				}

				duration := time.Since(start)
				times = append(times, duration)

				if err != nil {
					errors++
				}
			}

			// Calculate statistics
			sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

			avgTime := calculateAvg(times)
			p95Time := percentile(times, 95)
			p99Time := percentile(times, 99)
			errorRate := float64(errors) / float64(totalOps) * 100

			fmt.Printf("\n%s:\n", ratio.name)
			fmt.Printf("  Average Time: %v\n", avgTime)
			fmt.Printf("  P95 Time:     %v\n", p95Time)
			fmt.Printf("  P99 Time:     %v\n", p99Time)
			fmt.Printf("  Error Rate:   %.2f%%\n", errorRate)
		})
	}
}

// TestBatchQueryPerformance tests performance of batch operations.
func TestBatchQueryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping batch query tests in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	client := NewLoadTestClient(server.Server.URL)

	batchSizes := []int{10, 50, 100, 500}

	fmt.Println("\n=== Batch Query Performance ===")
	fmt.Printf("%-15s %15s %15s %15s\n",
		"Batch Size", "Total Time", "Per Item", "Items/Sec")
	fmt.Println(strings.Repeat("-", 65))

	for _, batchSize := range batchSizes {
		start := time.Now()

		// Create batch of customers
		for i := 0; i < batchSize; i++ {
			client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
				"code":   fmt.Sprintf("BATCH-%d-%s", batchSize, uuid.New().String()[:8]),
				"name":   fmt.Sprintf("Batch Customer %d", i),
				"status": "active",
			})
		}

		totalTime := time.Since(start)
		perItem := totalTime / time.Duration(batchSize)
		itemsPerSec := float64(batchSize) / totalTime.Seconds()

		fmt.Printf("%-15d %15v %15v %15.2f\n",
			batchSize, totalTime, perItem, itemsPerSec)
	}
}

// ============================================================================
// Query Benchmark Helpers
// ============================================================================

// benchmarkQuery runs a benchmark for a single query type.
func benchmarkQuery(name string, iterations int, query func() error) QueryBenchmarkResult {
	var times []time.Duration
	var totalTime time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		query()
		duration := time.Since(start)

		times = append(times, duration)
		totalTime += duration
	}

	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	return QueryBenchmarkResult{
		QueryName:    name,
		Iterations:   iterations,
		TotalTime:    totalTime,
		MinTime:      times[0],
		MaxTime:      times[len(times)-1],
		AvgTime:      totalTime / time.Duration(iterations),
		P50Time:      percentile(times, 50),
		P95Time:      percentile(times, 95),
		P99Time:      percentile(times, 99),
		OpsPerSecond: float64(iterations) / totalTime.Seconds(),
	}
}

// benchmarkConcurrentQuery runs a benchmark with concurrent queries.
func benchmarkConcurrentQuery(name string, concurrency, totalOps int, query func() error) QueryBenchmarkResult {
	var times []time.Duration
	var timesMu sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, concurrency)
	start := time.Now()

	for i := 0; i < totalOps; i++ {
		semaphore <- struct{}{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			opStart := time.Now()
			query()
			duration := time.Since(opStart)

			timesMu.Lock()
			times = append(times, duration)
			timesMu.Unlock()
		}()
	}

	wg.Wait()
	totalTime := time.Since(start)

	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

	return QueryBenchmarkResult{
		QueryName:    name,
		Iterations:   totalOps,
		TotalTime:    totalTime,
		MinTime:      times[0],
		MaxTime:      times[len(times)-1],
		AvgTime:      calculateAvg(times),
		P50Time:      percentile(times, 50),
		P95Time:      percentile(times, 95),
		P99Time:      percentile(times, 99),
		OpsPerSecond: float64(totalOps) / totalTime.Seconds(),
	}
}

// printQueryBenchmarkResults prints benchmark results in a formatted table.
func printQueryBenchmarkResults(results []QueryBenchmarkResult) {
	fmt.Printf("%-25s %12s %12s %12s %12s %12s\n",
		"Query", "Avg", "P50", "P95", "P99", "Ops/Sec")
	fmt.Println(strings.Repeat("-", 90))

	for _, r := range results {
		fmt.Printf("%-25s %12v %12v %12v %12v %12.2f\n",
			r.QueryName, r.AvgTime, r.P50Time, r.P95Time, r.P99Time, r.OpsPerSecond)
	}
}

// ============================================================================
// Standard Benchmarks (go test -bench)
// ============================================================================

// BenchmarkQueryListSmall benchmarks list query with small dataset.
func BenchmarkQueryListSmall(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	server.Seed(100, 0, 0)
	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.DoRequest("GET", "/api/v1/customers", nil)
	}
}

// BenchmarkQueryListMedium benchmarks list query with medium dataset.
func BenchmarkQueryListMedium(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	server.Seed(1000, 0, 0)
	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.DoRequest("GET", "/api/v1/customers", nil)
	}
}

// BenchmarkQueryListLarge benchmarks list query with large dataset.
func BenchmarkQueryListLarge(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	server.Seed(10000, 0, 0)
	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.DoRequest("GET", "/api/v1/customers", nil)
	}
}

// BenchmarkQueryGetByID benchmarks single item retrieval.
func BenchmarkQueryGetByID(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	server.Seed(10000, 0, 0)

	var customerID string
	for id := range server.DataStore.customers {
		customerID = id
		break
	}

	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.DoRequest("GET", "/api/v1/customers/"+customerID, nil)
	}
}

// BenchmarkQueryCreateCustomer benchmarks customer creation.
func BenchmarkQueryCreateCustomer(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
			"code":   fmt.Sprintf("BENCH-%d", i),
			"name":   "Benchmark Customer",
			"status": "active",
		})
	}
}

// BenchmarkQueryParallelList benchmarks parallel list queries.
func BenchmarkQueryParallelList(b *testing.B) {
	server := NewTestServer()
	defer server.Close()
	server.Seed(1000, 0, 0)
	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.DoRequest("GET", "/api/v1/customers", nil)
		}
	})
}
