// Package config provides configuration management utilities for the CRM application.
// It supports loading configuration from files, environment variables, and defaults.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	MongoDB  MongoDBConfig  `mapstructure:"mongodb"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Tracer   TracerConfig   `mapstructure:"tracer"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
}

// AppConfig holds application-specific configuration.
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"` // development, staging, production
	Debug       bool   `mapstructure:"debug"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	TLSCertFile     string        `mapstructure:"tls_cert_file"`
	TLSKeyFile      string        `mapstructure:"tls_key_file"`
}

// DatabaseConfig holds PostgreSQL database configuration.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// MongoDBConfig holds MongoDB configuration.
type MongoDBConfig struct {
	URI            string        `mapstructure:"uri"`
	Database       string        `mapstructure:"database"`
	MaxPoolSize    uint64        `mapstructure:"max_pool_size"`
	MinPoolSize    uint64        `mapstructure:"min_pool_size"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	ServerTimeout  time.Duration `mapstructure:"server_timeout"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// Addr returns the Redis address.
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RabbitMQConfig holds RabbitMQ configuration.
type RabbitMQConfig struct {
	URL               string        `mapstructure:"url"`
	Exchange          string        `mapstructure:"exchange"`
	ExchangeType      string        `mapstructure:"exchange_type"`
	ReconnectDelay    time.Duration `mapstructure:"reconnect_delay"`
	MaxReconnectDelay time.Duration `mapstructure:"max_reconnect_delay"`
	PrefetchCount     int           `mapstructure:"prefetch_count"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	Issuer           string        `mapstructure:"issuer"`
	Audience         string        `mapstructure:"audience"`
	AccessExpiry     time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry    time.Duration `mapstructure:"refresh_expiry"`
	SigningAlgorithm string        `mapstructure:"signing_algorithm"`
}

// LoggerConfig holds logger configuration.
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json or console
	TimeFormat string `mapstructure:"time_format"`
	Caller     bool   `mapstructure:"caller"`
}

// TracerConfig holds distributed tracing configuration.
type TracerConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	ServiceName string  `mapstructure:"service_name"`
	Endpoint    string  `mapstructure:"endpoint"`
	SampleRate  float64 `mapstructure:"sample_rate"`
}

// SMTPConfig holds email configuration.
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
	TLS      bool   `mapstructure:"tls"`
}

// Load loads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Search for config in common locations
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/app/configs")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not an error if env vars are used
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind environment variables
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Override with environment variables
	bindEnvVars(v)

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "crm-service")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", false)

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)
	v.SetDefault("server.shutdown_timeout", 30*time.Second)
	v.SetDefault("server.tls_enabled", false)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "crm")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("database.conn_max_idle_time", 5*time.Minute)

	// MongoDB defaults
	v.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	v.SetDefault("mongodb.database", "crm")
	v.SetDefault("mongodb.max_pool_size", 100)
	v.SetDefault("mongodb.min_pool_size", 10)
	v.SetDefault("mongodb.connect_timeout", 10*time.Second)
	v.SetDefault("mongodb.server_timeout", 30*time.Second)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.dial_timeout", 5*time.Second)
	v.SetDefault("redis.read_timeout", 3*time.Second)
	v.SetDefault("redis.write_timeout", 3*time.Second)

	// RabbitMQ defaults
	v.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("rabbitmq.exchange", "crm.events")
	v.SetDefault("rabbitmq.exchange_type", "topic")
	v.SetDefault("rabbitmq.reconnect_delay", 5*time.Second)
	v.SetDefault("rabbitmq.max_reconnect_delay", 60*time.Second)
	v.SetDefault("rabbitmq.prefetch_count", 10)

	// JWT defaults
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.issuer", "crm-service")
	v.SetDefault("jwt.audience", "crm-api")
	v.SetDefault("jwt.access_expiry", 24*time.Hour)
	v.SetDefault("jwt.refresh_expiry", 7*24*time.Hour)
	v.SetDefault("jwt.signing_algorithm", "HS256")

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.time_format", time.RFC3339Nano)
	v.SetDefault("logger.caller", false)

	// Tracer defaults
	v.SetDefault("tracer.enabled", false)
	v.SetDefault("tracer.service_name", "crm-service")
	v.SetDefault("tracer.endpoint", "http://localhost:14268/api/traces")
	v.SetDefault("tracer.sample_rate", 1.0)

	// SMTP defaults
	v.SetDefault("smtp.host", "localhost")
	v.SetDefault("smtp.port", 1025)
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from", "noreply@example.com")
	v.SetDefault("smtp.from_name", "CRM System")
	v.SetDefault("smtp.tls", false)
}

// bindEnvVars binds environment variables to config keys.
func bindEnvVars(v *viper.Viper) {
	// Map environment variables to config keys
	envMappings := map[string]string{
		"APP_ENV":          "app.environment",
		"APP_DEBUG":        "app.debug",
		"APP_PORT":         "server.port",
		"DB_HOST":          "database.host",
		"DB_PORT":          "database.port",
		"DB_USER":          "database.user",
		"DB_PASSWORD":      "database.password",
		"DB_NAME":          "database.dbname",
		"MONGODB_URI":      "mongodb.uri",
		"REDIS_HOST":       "redis.host",
		"REDIS_PORT":       "redis.port",
		"REDIS_PASSWORD":   "redis.password",
		"RABBITMQ_URL":     "rabbitmq.url",
		"JWT_SECRET":       "jwt.secret",
		"JWT_EXPIRY":       "jwt.access_expiry",
		"JAEGER_ENDPOINT":  "tracer.endpoint",
		"LOG_LEVEL":        "logger.level",
		"SMTP_HOST":        "smtp.host",
		"SMTP_PORT":        "smtp.port",
		"SMTP_FROM":        "smtp.from",
	}

	for env, key := range envMappings {
		if val := os.Getenv(env); val != "" {
			v.Set(key, val)
		}
	}
}

// MustLoad loads configuration and panics on error.
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// IsDevelopment returns true if the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsStaging returns true if the environment is staging.
func (c *Config) IsStaging() bool {
	return c.App.Environment == "staging"
}
