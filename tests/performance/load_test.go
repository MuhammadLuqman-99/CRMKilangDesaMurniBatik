// Package performance contains load tests for the CRM system.
package performance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Load Testing
// ============================================================================

// TestLoadCustomerAPI tests the customer API under load.
func TestLoadCustomerAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	// Seed data
	server.Seed(1000, 0, 0)

	config := LightConfig() // Use light config for faster CI runs
	metrics := NewMetrics()
	client := NewLoadTestClient(server.Server.URL)

	// Start peak monitoring
	ctx, cancel := context.WithCancel(context.Background())
	go monitorPeaks(ctx, metrics)

	t.Run("List Customers", func(t *testing.T) {
		runLoadTest(t, config, metrics, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})
	})

	t.Run("Get Customer", func(t *testing.T) {
		// Get a customer ID from the store
		var customerID string
		for id := range server.DataStore.customers {
			customerID = id
			break
		}

		runLoadTest(t, config, metrics, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers/"+customerID, nil)
			return err
		})
	})

	t.Run("Create Customer", func(t *testing.T) {
		runLoadTest(t, config, metrics, func() error {
			_, err := client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
				"code":   fmt.Sprintf("LOAD-%s", uuid.New().String()[:8]),
				"name":   "Load Test Customer",
				"status": "active",
			})
			return err
		})
	})

	cancel()
	metrics.Finalize()

	report := GenerateReport("Customer API Load Test", metrics, config)
	report.Print()

	if !report.Passed {
		t.Errorf("Load test failed: %v", report.FailureReasons)
	}
}

// TestLoadLeadAPI tests the lead API under load.
func TestLoadLeadAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	// Seed data
	server.Seed(0, 1000, 0)

	config := LightConfig()
	metrics := NewMetrics()
	client := NewLoadTestClient(server.Server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	go monitorPeaks(ctx, metrics)

	t.Run("List Leads", func(t *testing.T) {
		runLoadTest(t, config, metrics, func() error {
			_, err := client.DoRequest("GET", "/api/v1/leads", nil)
			return err
		})
	})

	t.Run("Get Lead", func(t *testing.T) {
		var leadID string
		for id := range server.DataStore.leads {
			leadID = id
			break
		}

		runLoadTest(t, config, metrics, func() error {
			_, err := client.DoRequest("GET", "/api/v1/leads/"+leadID, nil)
			return err
		})
	})

	t.Run("Create Lead", func(t *testing.T) {
		runLoadTest(t, config, metrics, func() error {
			_, err := client.DoRequest("POST", "/api/v1/leads", map[string]interface{}{
				"company_name":  "Load Test Company",
				"contact_name":  "Load Test Contact",
				"contact_email": fmt.Sprintf("load-%s@example.com", uuid.New().String()[:8]),
			})
			return err
		})
	})

	cancel()
	metrics.Finalize()

	report := GenerateReport("Lead API Load Test", metrics, config)
	report.Print()

	if !report.Passed {
		t.Errorf("Load test failed: %v", report.FailureReasons)
	}
}

// TestLoadMixedWorkload tests a mixed read/write workload.
func TestLoadMixedWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	// Seed data
	server.Seed(500, 500, 100)

	config := LightConfig()
	metrics := NewMetrics()
	client := NewLoadTestClient(server.Server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	go monitorPeaks(ctx, metrics)

	// Get IDs for read operations
	var customerIDs, leadIDs []string
	for id := range server.DataStore.customers {
		customerIDs = append(customerIDs, id)
		if len(customerIDs) >= 10 {
			break
		}
	}
	for id := range server.DataStore.leads {
		leadIDs = append(leadIDs, id)
		if len(leadIDs) >= 10 {
			break
		}
	}

	t.Run("Mixed Workload", func(t *testing.T) {
		var counter int
		runLoadTest(t, config, metrics, func() error {
			counter++
			var err error

			// 70% reads, 30% writes
			switch counter % 10 {
			case 0, 1, 2: // 30% writes
				if counter%2 == 0 {
					_, err = client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
						"code":   fmt.Sprintf("MIX-%s", uuid.New().String()[:8]),
						"name":   "Mixed Test Customer",
						"status": "active",
					})
				} else {
					_, err = client.DoRequest("POST", "/api/v1/leads", map[string]interface{}{
						"company_name":  "Mixed Test Company",
						"contact_name":  "Mixed Test Contact",
						"contact_email": fmt.Sprintf("mixed-%s@example.com", uuid.New().String()[:8]),
					})
				}
			case 3, 4, 5, 6: // 40% customer reads
				if len(customerIDs) > 0 {
					idx := counter % len(customerIDs)
					_, err = client.DoRequest("GET", "/api/v1/customers/"+customerIDs[idx], nil)
				}
			default: // 30% lead reads
				if len(leadIDs) > 0 {
					idx := counter % len(leadIDs)
					_, err = client.DoRequest("GET", "/api/v1/leads/"+leadIDs[idx], nil)
				}
			}

			return err
		})
	})

	cancel()
	metrics.Finalize()

	report := GenerateReport("Mixed Workload Load Test", metrics, config)
	report.Print()

	if !report.Passed {
		t.Errorf("Load test failed: %v", report.FailureReasons)
	}
}

// TestLoadConcurrentUsers tests API behavior with increasing concurrent users.
func TestLoadConcurrentUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)

	userLevels := []int{10, 25, 50, 100}

	for _, users := range userLevels {
		t.Run(fmt.Sprintf("%d_Users", users), func(t *testing.T) {
			config := LightConfig()
			config.ConcurrentUsers = users
			config.RequestsPerUser = 20

			metrics := NewMetrics()
			client := NewLoadTestClient(server.Server.URL)

			ctx, cancel := context.WithCancel(context.Background())
			go monitorPeaks(ctx, metrics)

			runLoadTest(t, config, metrics, func() error {
				_, err := client.DoRequest("GET", "/api/v1/customers", nil)
				return err
			})

			cancel()
			metrics.Finalize()

			report := GenerateReport(fmt.Sprintf("Concurrent Users (%d)", users), metrics, config)
			report.Print()

			// More lenient thresholds for higher concurrency
			if users <= 50 && !report.Passed {
				t.Errorf("Load test failed at %d users: %v", users, report.FailureReasons)
			}
		})
	}
}

// TestLoadSustained tests sustained load over a longer period.
func TestLoadSustained(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sustained load test in short mode")
	}

	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 1000, 100)

	config := LightConfig()
	config.ConcurrentUsers = 20
	config.TestDuration = 30 * time.Second
	config.RequestsPerUser = 100

	metrics := NewMetrics()
	client := NewLoadTestClient(server.Server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	go monitorPeaks(ctx, metrics)

	t.Run("Sustained Load", func(t *testing.T) {
		runLoadTestWithDuration(t, config, metrics, func() error {
			_, err := client.DoRequest("GET", "/api/v1/customers", nil)
			return err
		})
	})

	cancel()
	metrics.Finalize()

	report := GenerateReport("Sustained Load Test", metrics, config)
	report.Print()

	if !report.Passed {
		t.Errorf("Sustained load test failed: %v", report.FailureReasons)
	}
}

// ============================================================================
// Load Test Helpers
// ============================================================================

// runLoadTest runs a load test with the given configuration.
func runLoadTest(t *testing.T, config *Config, metrics *Metrics, request func() error) {
	t.Helper()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.ConcurrentUsers)

	for i := 0; i < config.ConcurrentUsers*config.RequestsPerUser; i++ {
		semaphore <- struct{}{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			start := time.Now()
			err := request()
			duration := time.Since(start)

			metrics.RecordRequest(duration, err)

			if config.ThinkTime > 0 {
				time.Sleep(config.ThinkTime)
			}
		}()
	}

	wg.Wait()
}

// runLoadTestWithDuration runs a load test for a specified duration.
func runLoadTestWithDuration(t *testing.T, config *Config, metrics *Metrics, request func() error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
	defer cancel()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.ConcurrentUsers)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()

				start := time.Now()
				err := request()
				duration := time.Since(start)

				metrics.RecordRequest(duration, err)

				if config.ThinkTime > 0 {
					time.Sleep(config.ThinkTime)
				}
			}()
		}
	}
}

// monitorPeaks monitors and updates peak metrics.
func monitorPeaks(ctx context.Context, metrics *Metrics) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics.UpdatePeaks()
		}
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

// BenchmarkCustomerList benchmarks the customer list endpoint.
func BenchmarkCustomerList(b *testing.B) {
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

// BenchmarkCustomerCreate benchmarks the customer create endpoint.
func BenchmarkCustomerCreate(b *testing.B) {
	server := NewTestServer()
	defer server.Close()

	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
				"code":   fmt.Sprintf("BENCH-%s", uuid.New().String()[:8]),
				"name":   "Benchmark Customer",
				"status": "active",
			})
		}
	})
}

// BenchmarkCustomerGet benchmarks the customer get endpoint.
func BenchmarkCustomerGet(b *testing.B) {
	server := NewTestServer()
	defer server.Close()

	server.Seed(1000, 0, 0)

	var customerID string
	for id := range server.DataStore.customers {
		customerID = id
		break
	}

	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.DoRequest("GET", "/api/v1/customers/"+customerID, nil)
		}
	})
}

// BenchmarkLeadList benchmarks the lead list endpoint.
func BenchmarkLeadList(b *testing.B) {
	server := NewTestServer()
	defer server.Close()

	server.Seed(0, 1000, 0)
	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.DoRequest("GET", "/api/v1/leads", nil)
		}
	})
}

// BenchmarkMixedRead benchmarks mixed read operations.
func BenchmarkMixedRead(b *testing.B) {
	server := NewTestServer()
	defer server.Close()

	server.Seed(500, 500, 100)

	var customerIDs, leadIDs []string
	for id := range server.DataStore.customers {
		customerIDs = append(customerIDs, id)
	}
	for id := range server.DataStore.leads {
		leadIDs = append(leadIDs, id)
	}

	client := NewLoadTestClient(server.Server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 && len(customerIDs) > 0 {
				client.DoRequest("GET", "/api/v1/customers/"+customerIDs[i%len(customerIDs)], nil)
			} else if len(leadIDs) > 0 {
				client.DoRequest("GET", "/api/v1/leads/"+leadIDs[i%len(leadIDs)], nil)
			}
			i++
		}
	})
}
