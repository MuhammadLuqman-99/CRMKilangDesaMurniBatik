// Package discovery provides service discovery abstractions for the CRM application.
package discovery

import (
	"context"
	"hash/fnv"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Load Balancer Strategies
// ============================================================================

// RoundRobinBalancer implements round-robin load balancing.
type RoundRobinBalancer struct {
	counter uint64
}

// NewRoundRobinBalancer creates a new round-robin balancer.
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// Select selects the next instance using round-robin.
func (b *RoundRobinBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	count := atomic.AddUint64(&b.counter, 1)
	index := int(count-1) % len(instances)
	return instances[index], nil
}

// WeightedRoundRobinBalancer implements weighted round-robin load balancing.
type WeightedRoundRobinBalancer struct {
	currentWeight int
	maxWeight     int
	gcd           int
	lastIndex     int
	mu            sync.Mutex
}

// NewWeightedRoundRobinBalancer creates a new weighted round-robin balancer.
func NewWeightedRoundRobinBalancer() *WeightedRoundRobinBalancer {
	return &WeightedRoundRobinBalancer{
		lastIndex: -1,
	}
}

// Select selects an instance using weighted round-robin.
func (b *WeightedRoundRobinBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Calculate GCD and max weight if needed
	weights := make([]int, len(instances))
	for i, inst := range instances {
		weights[i] = inst.Weight
		if weights[i] == 0 {
			weights[i] = 1
		}
	}

	b.gcd = gcdSlice(weights)
	b.maxWeight = maxSlice(weights)

	for {
		b.lastIndex = (b.lastIndex + 1) % len(instances)

		if b.lastIndex == 0 {
			b.currentWeight = b.currentWeight - b.gcd
			if b.currentWeight <= 0 {
				b.currentWeight = b.maxWeight
				if b.currentWeight == 0 {
					return nil, ErrNoHealthyInstances
				}
			}
		}

		weight := instances[b.lastIndex].Weight
		if weight == 0 {
			weight = 1
		}

		if weight >= b.currentWeight {
			return instances[b.lastIndex], nil
		}
	}
}

// RandomBalancer implements random load balancing.
type RandomBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewRandomBalancer creates a new random balancer.
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select selects a random instance.
func (b *RandomBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	b.mu.Lock()
	index := b.rand.Intn(len(instances))
	b.mu.Unlock()

	return instances[index], nil
}

// WeightedRandomBalancer implements weighted random load balancing.
type WeightedRandomBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewWeightedRandomBalancer creates a new weighted random balancer.
func NewWeightedRandomBalancer() *WeightedRandomBalancer {
	return &WeightedRandomBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select selects an instance using weighted random selection.
func (b *WeightedRandomBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Calculate total weight
	totalWeight := 0
	for _, inst := range instances {
		weight := inst.Weight
		if weight == 0 {
			weight = 1
		}
		totalWeight += weight
	}

	b.mu.Lock()
	randomWeight := b.rand.Intn(totalWeight)
	b.mu.Unlock()

	// Select instance based on weight
	current := 0
	for _, inst := range instances {
		weight := inst.Weight
		if weight == 0 {
			weight = 1
		}
		current += weight
		if randomWeight < current {
			return inst, nil
		}
	}

	return instances[len(instances)-1], nil
}

// LeastConnectionsBalancer implements least connections load balancing.
type LeastConnectionsBalancer struct {
	connections map[string]*int64
	mu          sync.RWMutex
}

// NewLeastConnectionsBalancer creates a new least connections balancer.
func NewLeastConnectionsBalancer() *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{
		connections: make(map[string]*int64),
	}
}

// Select selects the instance with the least active connections.
func (b *LeastConnectionsBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	var selected *ServiceInstance
	minConns := int64(-1)

	for _, inst := range instances {
		conns, ok := b.connections[inst.ID]
		if !ok {
			var zero int64
			b.connections[inst.ID] = &zero
			conns = &zero
		}

		if minConns < 0 || *conns < minConns {
			minConns = *conns
			selected = inst
		}
	}

	return selected, nil
}

// IncrementConnections increments the connection count for an instance.
func (b *LeastConnectionsBalancer) IncrementConnections(instanceID string) {
	b.mu.RLock()
	conns, ok := b.connections[instanceID]
	b.mu.RUnlock()

	if ok {
		atomic.AddInt64(conns, 1)
	}
}

// DecrementConnections decrements the connection count for an instance.
func (b *LeastConnectionsBalancer) DecrementConnections(instanceID string) {
	b.mu.RLock()
	conns, ok := b.connections[instanceID]
	b.mu.RUnlock()

	if ok {
		atomic.AddInt64(conns, -1)
	}
}

// IPHashBalancer implements IP hash load balancing for session affinity.
type IPHashBalancer struct{}

// NewIPHashBalancer creates a new IP hash balancer.
func NewIPHashBalancer() *IPHashBalancer {
	return &IPHashBalancer{}
}

// Select selects an instance based on client IP hash.
func (b *IPHashBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	return nil, ErrNoHealthyInstances // Need client IP
}

// SelectWithIP selects an instance based on client IP.
func (b *IPHashBalancer) SelectWithIP(instances []*ServiceInstance, clientIP string) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	hash := fnv.New32a()
	hash.Write([]byte(clientIP))
	index := int(hash.Sum32()) % len(instances)

	return instances[index], nil
}

// ConsistentHashBalancer implements consistent hashing for load balancing.
type ConsistentHashBalancer struct {
	replicas int
	ring     []hashRingEntry
	mu       sync.RWMutex
}

type hashRingEntry struct {
	hash     uint32
	instance *ServiceInstance
}

// NewConsistentHashBalancer creates a new consistent hash balancer.
func NewConsistentHashBalancer(replicas int) *ConsistentHashBalancer {
	if replicas <= 0 {
		replicas = 100
	}
	return &ConsistentHashBalancer{
		replicas: replicas,
		ring:     make([]hashRingEntry, 0),
	}
}

// Select is not supported for consistent hash - use SelectWithKey instead.
func (b *ConsistentHashBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	return nil, ErrNoHealthyInstances // Need a key
}

// SelectWithKey selects an instance based on a key.
func (b *ConsistentHashBalancer) SelectWithKey(instances []*ServiceInstance, key string) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	b.mu.Lock()
	b.buildRing(instances)
	b.mu.Unlock()

	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.ring) == 0 {
		return nil, ErrNoHealthyInstances
	}

	hash := hashKey(key)

	// Binary search for the first entry with hash >= our hash
	idx := sort.Search(len(b.ring), func(i int) bool {
		return b.ring[i].hash >= hash
	})

	if idx >= len(b.ring) {
		idx = 0
	}

	return b.ring[idx].instance, nil
}

// buildRing builds the consistent hash ring.
func (b *ConsistentHashBalancer) buildRing(instances []*ServiceInstance) {
	b.ring = make([]hashRingEntry, 0, len(instances)*b.replicas)

	for _, inst := range instances {
		for i := 0; i < b.replicas; i++ {
			key := inst.ID + "-" + string(rune(i))
			hash := hashKey(key)
			b.ring = append(b.ring, hashRingEntry{
				hash:     hash,
				instance: inst,
			})
		}
	}

	// Sort ring by hash
	sort.Slice(b.ring, func(i, j int) bool {
		return b.ring[i].hash < b.ring[j].hash
	})
}

// hashKey computes a hash for a key.
func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// ResponseTimeBalancer selects instances based on response time.
type ResponseTimeBalancer struct {
	responseTimes map[string]*responseTimeStats
	mu            sync.RWMutex
	rand          *rand.Rand
}

type responseTimeStats struct {
	avgResponseTime time.Duration
	count           int64
	lastUpdate      time.Time
}

// NewResponseTimeBalancer creates a new response time balancer.
func NewResponseTimeBalancer() *ResponseTimeBalancer {
	return &ResponseTimeBalancer{
		responseTimes: make(map[string]*responseTimeStats),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select selects the instance with the lowest average response time.
func (b *ResponseTimeBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	var selected *ServiceInstance
	minTime := time.Duration(-1)

	for _, inst := range instances {
		stats, ok := b.responseTimes[inst.ID]
		if !ok {
			// No stats yet, give it a chance
			return inst, nil
		}

		if minTime < 0 || stats.avgResponseTime < minTime {
			minTime = stats.avgResponseTime
			selected = inst
		}
	}

	if selected == nil {
		// Fallback to random
		return instances[b.rand.Intn(len(instances))], nil
	}

	return selected, nil
}

// RecordResponseTime records the response time for an instance.
func (b *ResponseTimeBalancer) RecordResponseTime(instanceID string, duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	stats, ok := b.responseTimes[instanceID]
	if !ok {
		stats = &responseTimeStats{}
		b.responseTimes[instanceID] = stats
	}

	// Calculate exponential moving average
	alpha := 0.3
	if stats.count == 0 {
		stats.avgResponseTime = duration
	} else {
		stats.avgResponseTime = time.Duration(
			float64(stats.avgResponseTime)*(1-alpha) + float64(duration)*alpha,
		)
	}

	stats.count++
	stats.lastUpdate = time.Now()
}

// ZoneAwareBalancer selects instances preferring the same zone.
type ZoneAwareBalancer struct {
	localZone string
	fallback  LoadBalancer
}

// NewZoneAwareBalancer creates a new zone-aware balancer.
func NewZoneAwareBalancer(localZone string, fallback LoadBalancer) *ZoneAwareBalancer {
	if fallback == nil {
		fallback = NewRoundRobinBalancer()
	}
	return &ZoneAwareBalancer{
		localZone: localZone,
		fallback:  fallback,
	}
}

// Select selects an instance, preferring instances in the local zone.
func (b *ZoneAwareBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Filter instances in local zone
	localInstances := make([]*ServiceInstance, 0)
	for _, inst := range instances {
		if inst.Zone == b.localZone {
			localInstances = append(localInstances, inst)
		}
	}

	// Use local instances if available
	if len(localInstances) > 0 {
		return b.fallback.Select(localInstances)
	}

	// Fallback to all instances
	return b.fallback.Select(instances)
}

// ============================================================================
// Composite Load Balancer
// ============================================================================

// CompositeBalancer combines multiple load balancing strategies.
type CompositeBalancer struct {
	primary   LoadBalancer
	secondary LoadBalancer
	healthCheck func(instance *ServiceInstance) bool
}

// NewCompositeBalancer creates a new composite balancer.
func NewCompositeBalancer(primary, secondary LoadBalancer, healthCheck func(*ServiceInstance) bool) *CompositeBalancer {
	return &CompositeBalancer{
		primary:     primary,
		secondary:   secondary,
		healthCheck: healthCheck,
	}
}

// Select selects an instance using the primary strategy, falling back to secondary.
func (b *CompositeBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Try primary
	inst, err := b.primary.Select(instances)
	if err == nil && (b.healthCheck == nil || b.healthCheck(inst)) {
		return inst, nil
	}

	// Fallback to secondary
	if b.secondary != nil {
		return b.secondary.Select(instances)
	}

	return nil, ErrNoHealthyInstances
}

// ============================================================================
// Load Balancer Factory
// ============================================================================

// BalancerType represents the type of load balancer.
type BalancerType string

const (
	BalancerTypeRoundRobin         BalancerType = "round_robin"
	BalancerTypeWeightedRoundRobin BalancerType = "weighted_round_robin"
	BalancerTypeRandom             BalancerType = "random"
	BalancerTypeWeightedRandom     BalancerType = "weighted_random"
	BalancerTypeLeastConnections   BalancerType = "least_connections"
	BalancerTypeIPHash             BalancerType = "ip_hash"
	BalancerTypeConsistentHash     BalancerType = "consistent_hash"
	BalancerTypeResponseTime       BalancerType = "response_time"
)

// NewLoadBalancer creates a load balancer of the specified type.
func NewLoadBalancer(balancerType BalancerType) LoadBalancer {
	switch balancerType {
	case BalancerTypeRoundRobin:
		return NewRoundRobinBalancer()
	case BalancerTypeWeightedRoundRobin:
		return NewWeightedRoundRobinBalancer()
	case BalancerTypeRandom:
		return NewRandomBalancer()
	case BalancerTypeWeightedRandom:
		return NewWeightedRandomBalancer()
	case BalancerTypeLeastConnections:
		return NewLeastConnectionsBalancer()
	case BalancerTypeResponseTime:
		return NewResponseTimeBalancer()
	default:
		return NewRoundRobinBalancer()
	}
}

// ============================================================================
// Service Client with Load Balancing
// ============================================================================

// ServiceClient provides service discovery with load balancing.
type ServiceClient struct {
	discovery ServiceDiscovery
	balancer  LoadBalancer
	cache     *ServiceCache
}

// ServiceClientConfig configures the service client.
type ServiceClientConfig struct {
	BalancerType BalancerType
	CacheTTL     time.Duration
}

// NewServiceClient creates a new service client.
func NewServiceClient(discovery ServiceDiscovery, config ServiceClientConfig) *ServiceClient {
	balancer := NewLoadBalancer(config.BalancerType)

	var cache *ServiceCache
	if config.CacheTTL > 0 {
		cache = NewServiceCache(config.CacheTTL)
	}

	return &ServiceClient{
		discovery: discovery,
		balancer:  balancer,
		cache:     cache,
	}
}

// GetInstance retrieves a single service instance using load balancing.
func (c *ServiceClient) GetInstance(ctx context.Context, serviceName string) (*ServiceInstance, error) {
	var instances []*ServiceInstance
	var err error

	// Check cache first
	if c.cache != nil {
		if cached, ok := c.cache.Get(serviceName); ok {
			instances = cached
		}
	}

	// Fetch from discovery if not cached
	if instances == nil {
		instances, err = c.discovery.GetHealthyService(ctx, serviceName)
		if err != nil {
			return nil, err
		}

		// Update cache
		if c.cache != nil {
			c.cache.Set(serviceName, instances)
		}
	}

	return c.balancer.Select(instances)
}

// GetAllInstances retrieves all healthy instances of a service.
func (c *ServiceClient) GetAllInstances(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	if c.cache != nil {
		if cached, ok := c.cache.Get(serviceName); ok {
			return cached, nil
		}
	}

	instances, err := c.discovery.GetHealthyService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.Set(serviceName, instances)
	}

	return instances, nil
}

// Watch watches for service changes and updates the cache.
func (c *ServiceClient) Watch(ctx context.Context, serviceName string) error {
	return c.discovery.Watch(ctx, serviceName, func(instances []*ServiceInstance) {
		if c.cache != nil {
			healthy := make([]*ServiceInstance, 0)
			for _, inst := range instances {
				if inst.IsHealthy() {
					healthy = append(healthy, inst)
				}
			}
			c.cache.Set(serviceName, healthy)
		}
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

// gcd calculates the greatest common divisor of two numbers.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// gcdSlice calculates the GCD of a slice of numbers.
func gcdSlice(nums []int) int {
	if len(nums) == 0 {
		return 1
	}
	result := nums[0]
	for i := 1; i < len(nums); i++ {
		result = gcd(result, nums[i])
	}
	return result
}

// maxSlice returns the maximum value in a slice.
func maxSlice(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for i := 1; i < len(nums); i++ {
		if nums[i] > max {
			max = nums[i]
		}
	}
	return max
}
