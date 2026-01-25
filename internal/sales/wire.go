//go:build wireinject
// +build wireinject

package sales

import (
	"database/sql"

	"github.com/google/wire"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/kilang-desa-murni/crm/internal/sales/application/usecase"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
	"github.com/kilang-desa-murni/crm/internal/sales/infrastructure/messaging"
	"github.com/kilang-desa-murni/crm/internal/sales/infrastructure/persistence/postgres"
	saleshttp "github.com/kilang-desa-murni/crm/internal/sales/interfaces/http"
)

// ============================================================================
// Provider Sets
// ============================================================================

// RepositorySet provides all repository implementations
var RepositorySet = wire.NewSet(
	postgres.NewLeadRepository,
	wire.Bind(new(domain.LeadRepository), new(*postgres.LeadRepository)),

	postgres.NewOpportunityRepository,
	wire.Bind(new(domain.OpportunityRepository), new(*postgres.OpportunityRepository)),

	postgres.NewDealRepository,
	wire.Bind(new(domain.DealRepository), new(*postgres.DealRepository)),

	postgres.NewPipelineRepository,
	wire.Bind(new(domain.PipelineRepository), new(*postgres.PipelineRepository)),

	postgres.NewOutboxRepository,
)

// UseCaseSet provides all use case implementations
var UseCaseSet = wire.NewSet(
	usecase.NewLeadUseCase,
	wire.Bind(new(usecase.LeadUseCase), new(*usecase.LeadUseCaseImpl)),

	usecase.NewOpportunityUseCase,
	wire.Bind(new(usecase.OpportunityUseCase), new(*usecase.OpportunityUseCaseImpl)),

	usecase.NewDealUseCase,
	wire.Bind(new(usecase.DealUseCase), new(*usecase.DealUseCaseImpl)),

	usecase.NewPipelineUseCase,
	wire.Bind(new(usecase.PipelineUseCase), new(*usecase.PipelineUseCaseImpl)),
)

// MessagingSet provides messaging implementations
var MessagingSet = wire.NewSet(
	messaging.NewRabbitMQPublisher,
)

// HTTPSet provides HTTP handler implementations
var HTTPSet = wire.NewSet(
	saleshttp.NewHandler,
	saleshttp.NewRouter,
)

// ============================================================================
// Infrastructure Dependencies
// ============================================================================

// InfrastructureDependencies holds external infrastructure dependencies
type InfrastructureDependencies struct {
	DB          *sql.DB
	RedisClient *redis.Client
	AMQPConn    *amqp.Connection
}

// ============================================================================
// Service Configuration
// ============================================================================

// ServiceConfig holds service configuration
type ServiceConfig struct {
	// HTTP Configuration
	HTTPConfig saleshttp.MiddlewareConfig

	// RabbitMQ Configuration
	RabbitMQExchange string
	RabbitMQRoutingKey string
}

// ============================================================================
// Wire Injectors
// ============================================================================

// InitializeSalesService creates the fully wired sales service
func InitializeSalesService(
	deps InfrastructureDependencies,
	config ServiceConfig,
) (*SalesService, error) {
	wire.Build(
		// Infrastructure
		RepositorySet,
		MessagingSet,

		// Application
		UseCaseSet,

		// Interfaces
		HTTPSet,

		// Service assembly
		wire.Struct(new(saleshttp.HandlerDependencies), "*"),
		NewSalesService,
	)
	return nil, nil
}

// ============================================================================
// Sales Service
// ============================================================================

// SalesService is the main service that holds all components
type SalesService struct {
	// HTTP Handler
	Handler *saleshttp.Handler

	// Event Publisher
	EventPublisher *messaging.RabbitMQPublisher

	// Repositories (for direct access if needed)
	LeadRepo        domain.LeadRepository
	OpportunityRepo domain.OpportunityRepository
	DealRepo        domain.DealRepository
	PipelineRepo    domain.PipelineRepository
}

// NewSalesService creates a new sales service instance
func NewSalesService(
	handler *saleshttp.Handler,
	publisher *messaging.RabbitMQPublisher,
	leadRepo domain.LeadRepository,
	opportunityRepo domain.OpportunityRepository,
	dealRepo domain.DealRepository,
	pipelineRepo domain.PipelineRepository,
) *SalesService {
	return &SalesService{
		Handler:         handler,
		EventPublisher:  publisher,
		LeadRepo:        leadRepo,
		OpportunityRepo: opportunityRepo,
		DealRepo:        dealRepo,
		PipelineRepo:    pipelineRepo,
	}
}

// Start starts the sales service background processes
func (s *SalesService) Start() error {
	// Start outbox processor
	if s.EventPublisher != nil {
		go s.EventPublisher.StartOutboxProcessor()
	}
	return nil
}

// Stop stops the sales service gracefully
func (s *SalesService) Stop() error {
	// Stop event publisher
	if s.EventPublisher != nil {
		return s.EventPublisher.Close()
	}
	return nil
}
