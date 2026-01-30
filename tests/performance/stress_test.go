// Package performance contains stress tests for the CRM system.
package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Stress Testing - Finding Breaking Points
// ============================================================================

// StressTestResult contains the results of a stress test.
type StressTestResult struct {
	MaxConcurrency    int
	BreakingPoint     int
	SuccessRate       float64
	AvgResponseTime   time.Duration
	P99ResponseTime   time.Duration
	ErrorsAtBreaking  int
	RecoveryTime      time.Duration
}

// TestStressFindBreakingPoint tests to find the system's breaking point.
func TestStressFindBreakingPoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)

	client := NewLoadTestClient(server.Server.URL)

	// Start with low concurrency and increase until errors occur
	concurrencyLevels := []int{10, 25, 50, 100, 200, 500}
	errorThreshold := 0.05 // 5% error rate is considered breaking point

	var breakingPoint int
	var lastSuccessfulConcurrency int

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			metrics := NewMetrics()

			ctx, cancel := context.WithCancel(context.Background())
			go monitorPeaks(ctx, metrics)

			// Run stress test
			runStressTest(concurrency, 100, metrics, func() error {
				_, err := client.DoRequest("GET", "/api/v1/customers", nil)
				return err
			})

			cancel()
			metrics.Finalize()

			// Calculate error rate
			errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)

			fmt.Printf("Concurrency %d: Total=%d, Errors=%d, ErrorRate=%.2f%%\n",
				concurrency, metrics.TotalRequests, metrics.FailedRequests, errorRate*100)

			if errorRate > errorThreshold {
				if breakingPoint == 0 {
					breakingPoint = concurrency
				}
			} else {
				lastSuccessfulConcurrency = concurrency
			}
		})
	}

	fmt.Printf("\n=== Stress Test Results ===\n")
	fmt.Printf("Last Successful Concurrency: %d\n", lastSuccessfulConcurrency)
	if breakingPoint > 0 {
		fmt.Printf("Breaking Point: %d concurrent users\n", breakingPoint)
	} else {
		fmt.Printf("Breaking Point: Not reached (system handled all load levels)\n")
	}
}

// TestStressRecovery tests system recovery after stress.
func TestStressRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress recovery test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)
	client := NewLoadTestClient(server.Server.URL)

	// Phase 1: Baseline measurement
	t.Log("Phase 1: Measuring baseline...")
	baselineMetrics := NewMetrics()
	runStressTest(10, 50, baselineMetrics, func() error {
		_, err := client.DoRequest("GET", "/api/v1/customers", nil)
		return err
	})
	baselineMetrics.Finalize()

	baselineAvg := calculateAvg(baselineMetrics.ResponseTimes)
	t.Logf("Baseline avg response time: %v", baselineAvg)

	// Phase 2: Apply stress
	t.Log("Phase 2: Applying stress...")
	stressMetrics := NewMetrics()
	runStressTest(200, 50, stressMetrics, func() error {
		_, err := client.DoRequest("GET", "/api/v1/customers", nil)
		return err
	})
	stressMetrics.Finalize()

	stressAvg := calculateAvg(stressMetrics.ResponseTimes)
	t.Logf("Stress avg response time: %v", stressAvg)

	// Phase 3: Recovery measurement
	t.Log("Phase 3: Measuring recovery...")
	time.Sleep(2 * time.Second) // Allow system to recover

	recoveryMetrics := NewMetrics()
	runStressTest(10, 50, recoveryMetrics, func() error {
		_, err := client.DoRequest("GET", "/api/v1/customers", nil)
		return err
	})
	recoveryMetrics.Finalize()

	recoveryAvg := calculateAvg(recoveryMetrics.ResponseTimes)
	t.Logf("Recovery avg response time: %v", recoveryAvg)

	// Check recovery
	// Recovery is successful if response time returns to within 2x of baseline
	recoveryThreshold := baselineAvg * 2
	if recoveryAvg > recoveryThreshold {
		t.Errorf("System did not recover properly. Expected <%v, got %v", recoveryThreshold, recoveryAvg)
	} else {
		t.Log("System recovered successfully")
	}
}

// TestStressSpikeLoad tests handling of sudden load spikes.
func TestStressSpikeLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping spike load test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)
	client := NewLoadTestClient(server.Server.URL)

	// Simulate spike load: normal -> spike -> normal
	t.Run("Load Spike", func(t *testing.T) {
		var totalErrors int64
		var totalRequests int64

		// Normal load
		t.Log("Phase 1: Normal load...")
		metrics1 := NewMetrics()
		runStressTest(20, 30, metrics1, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})
		metrics1.Finalize()
		atomic.AddInt64(&totalRequests, metrics1.TotalRequests)
		atomic.AddInt64(&totalErrors, metrics1.FailedRequests)
		t.Logf("Normal load: %d requests, %d errors", metrics1.TotalRequests, metrics1.FailedRequests)

		// Spike
		t.Log("Phase 2: Spike load...")
		metrics2 := NewMetrics()
		runStressTest(150, 30, metrics2, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})
		metrics2.Finalize()
		atomic.AddInt64(&totalRequests, metrics2.TotalRequests)
		atomic.AddInt64(&totalErrors, metrics2.FailedRequests)
		t.Logf("Spike load: %d requests, %d errors", metrics2.TotalRequests, metrics2.FailedRequests)

		// Return to normal
		t.Log("Phase 3: Return to normal...")
		metrics3 := NewMetrics()
		runStressTest(20, 30, metrics3, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})
		metrics3.Finalize()
		atomic.AddInt64(&totalRequests, metrics3.TotalRequests)
		atomic.AddInt64(&totalErrors, metrics3.FailedRequests)
		t.Logf("Normal load: %d requests, %d errors", metrics3.TotalRequests, metrics3.FailedRequests)

		// Overall error rate should be acceptable
		overallErrorRate := float64(totalErrors) / float64(totalRequests)
		t.Logf("Overall error rate: %.2f%%", overallErrorRate*100)

		if overallErrorRate > 0.1 { // 10% overall error rate during spike test
			t.Errorf("Error rate too high during spike test: %.2f%%", overallErrorRate*100)
		}
	})
}

// TestStressResourceExhaustion tests behavior under resource exhaustion.
func TestStressResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource exhaustion test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)
	client := NewLoadTestClient(server.Server.URL)

	t.Run("Connection Exhaustion", func(t *testing.T) {
		// Try to exhaust connections with many concurrent requests
		var wg sync.WaitGroup
		var successCount, errorCount int64
		concurrency := 500
		requestsPerGoroutine := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					_, err := client.DoRequest("GET", "/api/v1/customers", nil)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}
			}()
		}

		wg.Wait()

		total := successCount + errorCount
		successRate := float64(successCount) / float64(total) * 100

		t.Logf("Connection exhaustion test: %d/%d succeeded (%.2f%%)",
			successCount, total, successRate)

		// We expect some failures under extreme load, but majority should succeed
		if successRate < 50 {
			t.Errorf("Too many failures under connection pressure: %.2f%% success rate", successRate)
		}
	})

	t.Run("Memory Pressure", func(t *testing.T) {
		// Create many objects to increase memory pressure
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		// Run requests while creating memory pressure
		var wg sync.WaitGroup
		var successCount int64

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					_, err := client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
						"code":   fmt.Sprintf("MEM-%s", uuid.New().String()[:8]),
						"name":   "Memory Pressure Test",
						"status": "active",
					})
					if err == nil {
						atomic.AddInt64(&successCount, 1)
					}
				}
			}()
		}

		wg.Wait()

		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		memGrowthMB := float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024

		t.Logf("Memory pressure test: %d succeeded, memory growth: %.2f MB",
			successCount, memGrowthMB)

		// Memory growth should be bounded
		if memGrowthMB > 100 { // 100 MB threshold
			t.Errorf("Excessive memory growth under pressure: %.2f MB", memGrowthMB)
		}
	})
}

// TestStressGoroutineLeaks tests for goroutine leaks under stress.
func TestStressGoroutineLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine leak test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 0, 0)
	client := NewLoadTestClient(server.Server.URL)

	initialGoroutines := runtime.NumGoroutine()
	t.Logf("Initial goroutines: %d", initialGoroutines)

	// Run multiple stress cycles
	for cycle := 1; cycle <= 3; cycle++ {
		t.Logf("Stress cycle %d...", cycle)

		metrics := NewMetrics()
		runStressTest(100, 50, metrics, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})
		metrics.Finalize()

		// Allow goroutines to clean up
		time.Sleep(time.Second)
		runtime.GC()
		time.Sleep(time.Second)
	}

	finalGoroutines := runtime.NumGoroutine()
	goroutineGrowth := finalGoroutines - initialGoroutines

	t.Logf("Final goroutines: %d (growth: %d)", finalGoroutines, goroutineGrowth)

	// Allow for some variance, but flag significant leaks
	if goroutineGrowth > 50 {
		t.Errorf("Potential goroutine leak detected: %d goroutines added", goroutineGrowth)
	}
}

// TestStressEndurance tests system behavior under extended load.
func TestStressEndurance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping endurance test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)
	client := NewLoadTestClient(server.Server.URL)

	duration := 30 * time.Second
	concurrency := 50

	t.Logf("Running endurance test for %v with %d concurrent users...", duration, concurrency)

	metrics := NewMetrics()
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	go monitorPeaks(ctx, metrics)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	requestTypes := []func() error{
		func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		},
		func() error {
			_, err := client.DoRequest("GET", "/api/v1/leads", nil)
			return err
		},
		func() error {
			_, err := client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
				"code":   fmt.Sprintf("END-%s", uuid.New().String()[:8]),
				"name":   "Endurance Customer",
				"status": "active",
			})
			return err
		},
	}

	counter := 0
	for {
		select {
		case <-ctx.Done():
			goto done
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func(reqNum int) {
				defer wg.Done()
				defer func() { <-semaphore }()

				request := requestTypes[reqNum%len(requestTypes)]
				start := time.Now()
				err := request()
				duration := time.Since(start)

				metrics.RecordRequest(duration, err)
			}(counter)
			counter++
		}
	}

done:
	wg.Wait()
	metrics.Finalize()

	config := LightConfig()
	config.ConcurrentUsers = concurrency
	config.TestDuration = duration

	report := GenerateReport("Endurance Test", metrics, config)
	report.Print()

	// Check for degradation over time
	if report.P99ResponseTime > 5*time.Second {
		t.Errorf("Response time degraded during endurance test: P99=%v", report.P99ResponseTime)
	}

	if report.ErrorRate > 0.05 {
		t.Errorf("Error rate too high during endurance test: %.2f%%", report.ErrorRate*100)
	}
}

// ============================================================================
// Stress Test Helpers
// ============================================================================

// runStressTest runs a stress test with given concurrency.
func runStressTest(concurrency, requestsPerUser int, metrics *Metrics, request func() error) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < concurrency*requestsPerUser; i++ {
		semaphore <- struct{}{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			start := time.Now()
			err := request()
			duration := time.Since(start)

			metrics.RecordRequest(duration, err)
		}()
	}

	wg.Wait()
}
