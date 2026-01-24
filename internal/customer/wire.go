//go:build wireinject
// +build wireinject

// Package customer provides the Customer service with wire dependency injection.
package customer

import (
	"context"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/ports"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/usecase"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/infrastructure/cache"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/infrastructure/messaging"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/infrastructure/persistence/mongodb"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/interfaces/http"
)

// Config contains configuration for the Customer service.
type Config struct {
	MongoDB    MongoDBConfig
	RabbitMQ   RabbitMQConfig
	Redis      RedisConfig
	HTTP       HTTPConfig
}

// MongoDBConfig contains MongoDB configuration.
type MongoDBConfig struct {
	URI      string
	Database string
}

// RabbitMQConfig contains RabbitMQ configuration.
type RabbitMQConfig struct {
	URI      string
	Exchange string
}

// RedisConfig contains Redis configuration.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// HTTPConfig contains HTTP server configuration.
type HTTPConfig struct {
	Port            int
	ReadTimeout     int
	WriteTimeout    int
	ShutdownTimeout int
}

// Service represents the Customer service with all dependencies.
type Service struct {
	Router         *http.Router
	UnitOfWork     *mongodb.UnitOfWork
	EventPublisher *messaging.RabbitMQPublisher
	Logger         *zap.Logger
}

// ProviderSet is the wire provider set for the Customer service.
var ProviderSet = wire.NewSet(
	// Repositories
	ProvideCustomerRepository,
	ProvideContactRepository,
	ProvideNoteRepository,
	ProvideActivityRepository,
	ProvideSegmentRepository,
	ProvideImportRepository,
	ProvideOutboxRepository,
	ProvideUnitOfWork,

	// Infrastructure
	ProvideEventPublisher,
	ProvideCacheService,

	// Use cases - Customer
	ProvideCreateCustomerUseCase,
	ProvideGetCustomerUseCase,
	ProvideUpdateCustomerUseCase,
	ProvideDeleteCustomerUseCase,
	ProvideSearchCustomersUseCase,

	// Use cases - Contact
	ProvideAddContactUseCase,
	ProvideGetContactUseCase,
	ProvideUpdateContactUseCase,
	ProvideDeleteContactUseCase,
	ProvideListContactsUseCase,
	ProvideSetPrimaryContactUseCase,

	// HTTP
	ProvideHandler,
	ProvideRouter,
)

// ============================================================================
// Repository Providers
// ============================================================================

// ProvideCustomerRepository provides a CustomerRepository.
func ProvideCustomerRepository(db *mongo.Database) domain.CustomerRepository {
	return mongodb.NewCustomerRepository(db)
}

// ProvideContactRepository provides a ContactRepository.
func ProvideContactRepository(db *mongo.Database) domain.ContactRepository {
	return mongodb.NewContactRepository(db)
}

// ProvideNoteRepository provides a NoteRepository.
func ProvideNoteRepository(db *mongo.Database) domain.NoteRepository {
	return mongodb.NewNoteRepository(db)
}

// ProvideActivityRepository provides an ActivityRepository.
func ProvideActivityRepository(db *mongo.Database) domain.ActivityRepository {
	return mongodb.NewActivityRepository(db)
}

// ProvideSegmentRepository provides a SegmentRepository.
func ProvideSegmentRepository(db *mongo.Database) domain.SegmentRepository {
	return mongodb.NewSegmentRepository(db)
}

// ProvideImportRepository provides an ImportRepository.
func ProvideImportRepository(db *mongo.Database) domain.ImportRepository {
	return mongodb.NewImportRepository(db)
}

// ProvideOutboxRepository provides an OutboxRepository.
func ProvideOutboxRepository(db *mongo.Database) *mongodb.OutboxRepository {
	return mongodb.NewOutboxRepository(db)
}

// ProvideUnitOfWork provides a UnitOfWork.
func ProvideUnitOfWork(client *mongo.Client, db *mongo.Database) *mongodb.UnitOfWork {
	return mongodb.NewUnitOfWork(client, db)
}

// ============================================================================
// Infrastructure Providers
// ============================================================================

// ProvideEventPublisher provides an EventPublisher.
func ProvideEventPublisher(
	ctx context.Context,
	config RabbitMQConfig,
	outboxRepo *mongodb.OutboxRepository,
	logger *zap.Logger,
) (ports.EventPublisher, error) {
	publisher, err := messaging.NewRabbitMQPublisher(
		ctx,
		config.URI,
		config.Exchange,
		outboxRepo,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return publisher, nil
}

// ProvideCacheService provides a CacheService.
func ProvideCacheService(config RedisConfig, logger *zap.Logger) (ports.CacheService, error) {
	if config.Addr == "" {
		// Return in-memory cache for development
		return cache.NewInMemoryCache(), nil
	}
	return cache.NewRedisCache(config.Addr, config.Password, config.DB, logger)
}

// ============================================================================
// Use Case Providers - Customer
// ============================================================================

// ProvideCreateCustomerUseCase provides a CreateCustomerUseCase.
func ProvideCreateCustomerUseCase(
	customerRepo domain.CustomerRepository,
	eventPublisher ports.EventPublisher,
	cacheService ports.CacheService,
	logger *zap.Logger,
) *usecase.CreateCustomerUseCase {
	return usecase.NewCreateCustomerUseCase(customerRepo, eventPublisher, cacheService, logger)
}

// ProvideGetCustomerUseCase provides a GetCustomerUseCase.
func ProvideGetCustomerUseCase(
	customerRepo domain.CustomerRepository,
	cacheService ports.CacheService,
	logger *zap.Logger,
) *usecase.GetCustomerUseCase {
	return usecase.NewGetCustomerUseCase(customerRepo, cacheService, logger)
}

// ProvideUpdateCustomerUseCase provides an UpdateCustomerUseCase.
func ProvideUpdateCustomerUseCase(
	customerRepo domain.CustomerRepository,
	eventPublisher ports.EventPublisher,
	cacheService ports.CacheService,
	logger *zap.Logger,
) *usecase.UpdateCustomerUseCase {
	return usecase.NewUpdateCustomerUseCase(customerRepo, eventPublisher, cacheService, logger)
}

// ProvideDeleteCustomerUseCase provides a DeleteCustomerUseCase.
func ProvideDeleteCustomerUseCase(
	customerRepo domain.CustomerRepository,
	eventPublisher ports.EventPublisher,
	cacheService ports.CacheService,
	logger *zap.Logger,
) *usecase.DeleteCustomerUseCase {
	return usecase.NewDeleteCustomerUseCase(customerRepo, eventPublisher, cacheService, logger)
}

// ProvideSearchCustomersUseCase provides a SearchCustomersUseCase.
func ProvideSearchCustomersUseCase(
	customerRepo domain.CustomerRepository,
	logger *zap.Logger,
) *usecase.SearchCustomersUseCase {
	return usecase.NewSearchCustomersUseCase(customerRepo, logger)
}

// ============================================================================
// Use Case Providers - Contact
// ============================================================================

// ProvideAddContactUseCase provides an AddContactUseCase.
func ProvideAddContactUseCase(
	contactRepo domain.ContactRepository,
	customerRepo domain.CustomerRepository,
	eventPublisher ports.EventPublisher,
	logger *zap.Logger,
) *usecase.AddContactUseCase {
	return usecase.NewAddContactUseCase(contactRepo, customerRepo, eventPublisher, logger)
}

// ProvideGetContactUseCase provides a GetContactUseCase.
func ProvideGetContactUseCase(
	contactRepo domain.ContactRepository,
	logger *zap.Logger,
) *usecase.GetContactUseCase {
	return usecase.NewGetContactUseCase(contactRepo, logger)
}

// ProvideUpdateContactUseCase provides an UpdateContactUseCase.
func ProvideUpdateContactUseCase(
	contactRepo domain.ContactRepository,
	eventPublisher ports.EventPublisher,
	logger *zap.Logger,
) *usecase.UpdateContactUseCase {
	return usecase.NewUpdateContactUseCase(contactRepo, eventPublisher, logger)
}

// ProvideDeleteContactUseCase provides a DeleteContactUseCase.
func ProvideDeleteContactUseCase(
	contactRepo domain.ContactRepository,
	eventPublisher ports.EventPublisher,
	logger *zap.Logger,
) *usecase.DeleteContactUseCase {
	return usecase.NewDeleteContactUseCase(contactRepo, eventPublisher, logger)
}

// ProvideListContactsUseCase provides a ListContactsUseCase.
func ProvideListContactsUseCase(
	contactRepo domain.ContactRepository,
	logger *zap.Logger,
) *usecase.ListContactsUseCase {
	return usecase.NewListContactsUseCase(contactRepo, logger)
}

// ProvideSetPrimaryContactUseCase provides a SetPrimaryContactUseCase.
func ProvideSetPrimaryContactUseCase(
	contactRepo domain.ContactRepository,
	eventPublisher ports.EventPublisher,
	logger *zap.Logger,
) *usecase.SetPrimaryContactUseCase {
	return usecase.NewSetPrimaryContactUseCase(contactRepo, eventPublisher, logger)
}

// ============================================================================
// HTTP Providers
// ============================================================================

// ProvideHandler provides HTTP handlers.
func ProvideHandler(deps http.HandlerDependencies) *http.Handler {
	return http.NewHandler(deps)
}

// ProvideRouter provides the HTTP router.
func ProvideRouter(handler *http.Handler, config *http.RouterConfig) *http.Router {
	return http.NewRouter(handler, config)
}

// ============================================================================
// Wire Injector Function
// ============================================================================

// InitializeService initializes the Customer service with all dependencies.
func InitializeService(
	ctx context.Context,
	client *mongo.Client,
	db *mongo.Database,
	config Config,
	logger *zap.Logger,
) (*Service, func(), error) {
	wire.Build(ProviderSet, wire.Struct(new(Service), "*"))
	return nil, nil, nil
}
