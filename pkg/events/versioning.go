// Package events provides event bus abstractions for the CRM application.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Event Version Types
// ============================================================================

// Version represents an event version.
type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// String returns the string representation of the version.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// IsCompatibleWith checks if this version is compatible with another.
// Compatible means same major version.
func (v Version) IsCompatibleWith(other Version) bool {
	return v.Major == other.Major
}

// ParseVersion parses a version string.
func ParseVersion(s string) (Version, error) {
	var v Version
	_, err := fmt.Sscanf(s, "%d.%d.%d", &v.Major, &v.Minor, &v.Patch)
	if err != nil {
		return Version{}, fmt.Errorf("invalid version format: %s", s)
	}
	return v, nil
}

// NewVersion creates a new version.
func NewVersion(major, minor, patch int) Version {
	return Version{Major: major, Minor: minor, Patch: patch}
}

// ============================================================================
// Event Schema
// ============================================================================

// EventSchema defines the schema for an event type at a specific version.
type EventSchema struct {
	EventType   EventType       `json:"event_type"`
	Version     Version         `json:"version"`
	Fields      []SchemaField   `json:"fields"`
	Description string          `json:"description"`
	Deprecated  bool            `json:"deprecated"`
	DeprecatedMessage string    `json:"deprecated_message,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SchemaField defines a field in an event schema.
type SchemaField struct {
	Name        string      `json:"name"`
	Type        FieldType   `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
	Deprecated  bool        `json:"deprecated"`
}

// FieldType represents the type of a schema field.
type FieldType string

const (
	FieldTypeString   FieldType = "string"
	FieldTypeInt      FieldType = "int"
	FieldTypeFloat    FieldType = "float"
	FieldTypeBool     FieldType = "bool"
	FieldTypeArray    FieldType = "array"
	FieldTypeObject   FieldType = "object"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeUUID     FieldType = "uuid"
	FieldTypeJSON     FieldType = "json"
)

// Validate validates event data against the schema.
func (s *EventSchema) Validate(data map[string]interface{}) error {
	for _, field := range s.Fields {
		value, exists := data[field.Name]

		if field.Required && !exists {
			return fmt.Errorf("required field '%s' is missing", field.Name)
		}

		if exists && value != nil {
			if err := validateFieldType(field.Name, value, field.Type); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateFieldType validates a field value against its type.
func validateFieldType(name string, value interface{}, fieldType FieldType) error {
	switch fieldType {
	case FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' should be string", name)
		}
	case FieldTypeInt:
		switch value.(type) {
		case int, int32, int64, float64:
			// JSON unmarshals numbers as float64
		default:
			return fmt.Errorf("field '%s' should be int", name)
		}
	case FieldTypeFloat:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("field '%s' should be float", name)
		}
	case FieldTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' should be bool", name)
		}
	case FieldTypeArray:
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("field '%s' should be array", name)
		}
	case FieldTypeObject:
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("field '%s' should be object", name)
		}
	case FieldTypeDateTime, FieldTypeUUID, FieldTypeJSON:
		// Accept string representation
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' should be string", name)
		}
	}
	return nil
}

// ============================================================================
// Event Migration
// ============================================================================

// MigrationFunc is a function that transforms event data from one version to another.
type MigrationFunc func(data map[string]interface{}) (map[string]interface{}, error)

// Migration represents a migration from one version to another.
type Migration struct {
	EventType   EventType     `json:"event_type"`
	FromVersion Version       `json:"from_version"`
	ToVersion   Version       `json:"to_version"`
	Description string        `json:"description"`
	Migrate     MigrationFunc `json:"-"`
}

// ============================================================================
// Event Version Registry
// ============================================================================

// VersionRegistry manages event schemas and migrations.
type VersionRegistry struct {
	schemas     map[EventType]map[string]*EventSchema // eventType -> version -> schema
	migrations  map[EventType][]Migration             // eventType -> migrations
	currentVersions map[EventType]Version             // eventType -> current version
	mu          sync.RWMutex
}

// NewVersionRegistry creates a new version registry.
func NewVersionRegistry() *VersionRegistry {
	return &VersionRegistry{
		schemas:         make(map[EventType]map[string]*EventSchema),
		migrations:      make(map[EventType][]Migration),
		currentVersions: make(map[EventType]Version),
	}
}

// RegisterSchema registers an event schema.
func (r *VersionRegistry) RegisterSchema(schema *EventSchema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.schemas[schema.EventType]; !ok {
		r.schemas[schema.EventType] = make(map[string]*EventSchema)
	}

	versionKey := schema.Version.String()
	if _, exists := r.schemas[schema.EventType][versionKey]; exists {
		return fmt.Errorf("schema already registered: %s v%s", schema.EventType, versionKey)
	}

	r.schemas[schema.EventType][versionKey] = schema

	// Update current version if this is newer
	if current, ok := r.currentVersions[schema.EventType]; !ok || schema.Version.Compare(current) > 0 {
		r.currentVersions[schema.EventType] = schema.Version
	}

	return nil
}

// GetSchema retrieves a schema for an event type and version.
func (r *VersionRegistry) GetSchema(eventType EventType, version Version) (*EventSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.schemas[eventType]
	if !ok {
		return nil, fmt.Errorf("no schemas registered for event type: %s", eventType)
	}

	schema, ok := versions[version.String()]
	if !ok {
		return nil, fmt.Errorf("schema not found: %s v%s", eventType, version.String())
	}

	return schema, nil
}

// GetCurrentSchema retrieves the current (latest) schema for an event type.
func (r *VersionRegistry) GetCurrentSchema(eventType EventType) (*EventSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	version, ok := r.currentVersions[eventType]
	if !ok {
		return nil, fmt.Errorf("no schemas registered for event type: %s", eventType)
	}

	return r.schemas[eventType][version.String()], nil
}

// GetCurrentVersion returns the current version for an event type.
func (r *VersionRegistry) GetCurrentVersion(eventType EventType) (Version, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	version, ok := r.currentVersions[eventType]
	if !ok {
		return Version{}, fmt.Errorf("no schemas registered for event type: %s", eventType)
	}

	return version, nil
}

// ListSchemas lists all schemas for an event type.
func (r *VersionRegistry) ListSchemas(eventType EventType) []*EventSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.schemas[eventType]
	if !ok {
		return nil
	}

	schemas := make([]*EventSchema, 0, len(versions))
	for _, schema := range versions {
		schemas = append(schemas, schema)
	}

	// Sort by version
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Version.Compare(schemas[j].Version) < 0
	})

	return schemas
}

// RegisterMigration registers a migration between versions.
func (r *VersionRegistry) RegisterMigration(migration Migration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate versions exist
	versions, ok := r.schemas[migration.EventType]
	if !ok {
		return fmt.Errorf("no schemas registered for event type: %s", migration.EventType)
	}

	if _, ok := versions[migration.FromVersion.String()]; !ok {
		return fmt.Errorf("from version not found: %s", migration.FromVersion.String())
	}

	if _, ok := versions[migration.ToVersion.String()]; !ok {
		return fmt.Errorf("to version not found: %s", migration.ToVersion.String())
	}

	r.migrations[migration.EventType] = append(r.migrations[migration.EventType], migration)

	// Sort migrations by from version
	sort.Slice(r.migrations[migration.EventType], func(i, j int) bool {
		return r.migrations[migration.EventType][i].FromVersion.Compare(
			r.migrations[migration.EventType][j].FromVersion) < 0
	})

	return nil
}

// GetMigrationPath finds the migration path from one version to another.
func (r *VersionRegistry) GetMigrationPath(eventType EventType, from, to Version) ([]Migration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if from.Compare(to) == 0 {
		return nil, nil // No migration needed
	}

	migrations, ok := r.migrations[eventType]
	if !ok {
		return nil, fmt.Errorf("no migrations registered for event type: %s", eventType)
	}

	// Find path using BFS
	path := r.findMigrationPath(migrations, from, to)
	if path == nil {
		return nil, fmt.Errorf("no migration path from %s to %s", from.String(), to.String())
	}

	return path, nil
}

// findMigrationPath finds a migration path using BFS.
func (r *VersionRegistry) findMigrationPath(migrations []Migration, from, to Version) []Migration {
	type node struct {
		version Version
		path    []Migration
	}

	visited := make(map[string]bool)
	queue := []node{{version: from, path: nil}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.version.Compare(to) == 0 {
			return current.path
		}

		versionKey := current.version.String()
		if visited[versionKey] {
			continue
		}
		visited[versionKey] = true

		for _, m := range migrations {
			if m.FromVersion.Compare(current.version) == 0 {
				newPath := make([]Migration, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = m
				queue = append(queue, node{version: m.ToVersion, path: newPath})
			}
		}
	}

	return nil
}

// ============================================================================
// Event Versioner
// ============================================================================

// EventVersioner handles event versioning and migration.
type EventVersioner struct {
	registry *VersionRegistry
}

// NewEventVersioner creates a new event versioner.
func NewEventVersioner(registry *VersionRegistry) *EventVersioner {
	return &EventVersioner{
		registry: registry,
	}
}

// VersionedEvent represents an event with version information.
type VersionedEvent struct {
	*Event
	SchemaVersion Version `json:"schema_version"`
}

// CreateVersionedEvent creates a versioned event.
func (v *EventVersioner) CreateVersionedEvent(event *Event) (*VersionedEvent, error) {
	version, err := v.registry.GetCurrentVersion(event.Type)
	if err != nil {
		// Default to 1.0.0 if no schema registered
		version = NewVersion(1, 0, 0)
	}

	return &VersionedEvent{
		Event:         event,
		SchemaVersion: version,
	}, nil
}

// Upcast converts an event from an older version to the current version.
func (v *EventVersioner) Upcast(ctx context.Context, event *VersionedEvent) (*VersionedEvent, error) {
	currentVersion, err := v.registry.GetCurrentVersion(event.Type)
	if err != nil {
		return event, nil // No schema, return as-is
	}

	if event.SchemaVersion.Compare(currentVersion) >= 0 {
		return event, nil // Already at current or newer version
	}

	// Get migration path
	path, err := v.registry.GetMigrationPath(event.Type, event.SchemaVersion, currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration path: %w", err)
	}

	// Parse payload
	var data map[string]interface{}
	if err := json.Unmarshal(event.Payload, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply migrations
	for _, migration := range path {
		data, err = migration.Migrate(data)
		if err != nil {
			return nil, fmt.Errorf("migration from %s to %s failed: %w",
				migration.FromVersion.String(), migration.ToVersion.String(), err)
		}
	}

	// Re-encode payload
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal migrated payload: %w", err)
	}

	// Create new versioned event
	newEvent := *event.Event
	newEvent.Payload = payload

	return &VersionedEvent{
		Event:         &newEvent,
		SchemaVersion: currentVersion,
	}, nil
}

// Downcast converts an event to an older version for compatibility.
func (v *EventVersioner) Downcast(ctx context.Context, event *VersionedEvent, targetVersion Version) (*VersionedEvent, error) {
	if event.SchemaVersion.Compare(targetVersion) <= 0 {
		return event, nil // Already at target or older version
	}

	// Get migration path (reverse)
	path, err := v.registry.GetMigrationPath(event.Type, event.SchemaVersion, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration path: %w", err)
	}

	// Parse payload
	var data map[string]interface{}
	if err := json.Unmarshal(event.Payload, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Apply migrations in reverse
	for i := len(path) - 1; i >= 0; i-- {
		migration := path[i]
		data, err = migration.Migrate(data)
		if err != nil {
			return nil, fmt.Errorf("downcast migration failed: %w", err)
		}
	}

	// Re-encode payload
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal migrated payload: %w", err)
	}

	// Create new versioned event
	newEvent := *event.Event
	newEvent.Payload = payload

	return &VersionedEvent{
		Event:         &newEvent,
		SchemaVersion: targetVersion,
	}, nil
}

// ValidateEvent validates an event against its schema.
func (v *EventVersioner) ValidateEvent(event *VersionedEvent) error {
	schema, err := v.registry.GetSchema(event.Type, event.SchemaVersion)
	if err != nil {
		return nil // No schema, skip validation
	}

	if schema.Deprecated {
		// Log warning but don't fail
		fmt.Printf("Warning: Event type %s v%s is deprecated: %s\n",
			event.Type, event.SchemaVersion.String(), schema.DeprecatedMessage)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(event.Payload, &data); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return schema.Validate(data)
}

// ============================================================================
// Schema Builder
// ============================================================================

// SchemaBuilder provides a fluent interface for building event schemas.
type SchemaBuilder struct {
	schema *EventSchema
}

// NewSchemaBuilder creates a new schema builder.
func NewSchemaBuilder(eventType EventType, version Version) *SchemaBuilder {
	return &SchemaBuilder{
		schema: &EventSchema{
			EventType: eventType,
			Version:   version,
			Fields:    make([]SchemaField, 0),
			CreatedAt: time.Now(),
		},
	}
}

// Description sets the schema description.
func (b *SchemaBuilder) Description(desc string) *SchemaBuilder {
	b.schema.Description = desc
	return b
}

// Deprecated marks the schema as deprecated.
func (b *SchemaBuilder) Deprecated(message string) *SchemaBuilder {
	b.schema.Deprecated = true
	b.schema.DeprecatedMessage = message
	return b
}

// Field adds a field to the schema.
func (b *SchemaBuilder) Field(name string, fieldType FieldType, required bool) *SchemaBuilder {
	b.schema.Fields = append(b.schema.Fields, SchemaField{
		Name:     name,
		Type:     fieldType,
		Required: required,
	})
	return b
}

// FieldWithDefault adds a field with a default value.
func (b *SchemaBuilder) FieldWithDefault(name string, fieldType FieldType, defaultValue interface{}) *SchemaBuilder {
	b.schema.Fields = append(b.schema.Fields, SchemaField{
		Name:     name,
		Type:     fieldType,
		Required: false,
		Default:  defaultValue,
	})
	return b
}

// FieldWithDescription adds a field with a description.
func (b *SchemaBuilder) FieldWithDescription(name string, fieldType FieldType, required bool, desc string) *SchemaBuilder {
	b.schema.Fields = append(b.schema.Fields, SchemaField{
		Name:        name,
		Type:        fieldType,
		Required:    required,
		Description: desc,
	})
	return b
}

// DeprecatedField adds a deprecated field.
func (b *SchemaBuilder) DeprecatedField(name string, fieldType FieldType, required bool) *SchemaBuilder {
	b.schema.Fields = append(b.schema.Fields, SchemaField{
		Name:       name,
		Type:       fieldType,
		Required:   required,
		Deprecated: true,
	})
	return b
}

// Build builds and returns the schema.
func (b *SchemaBuilder) Build() *EventSchema {
	return b.schema
}

// ============================================================================
// Migration Builder
// ============================================================================

// MigrationBuilder provides a fluent interface for building migrations.
type MigrationBuilder struct {
	migration Migration
	transforms []func(map[string]interface{}) error
}

// NewMigrationBuilder creates a new migration builder.
func NewMigrationBuilder(eventType EventType, from, to Version) *MigrationBuilder {
	return &MigrationBuilder{
		migration: Migration{
			EventType:   eventType,
			FromVersion: from,
			ToVersion:   to,
		},
		transforms: make([]func(map[string]interface{}) error, 0),
	}
}

// Description sets the migration description.
func (b *MigrationBuilder) Description(desc string) *MigrationBuilder {
	b.migration.Description = desc
	return b
}

// RenameField adds a field rename transformation.
func (b *MigrationBuilder) RenameField(from, to string) *MigrationBuilder {
	b.transforms = append(b.transforms, func(data map[string]interface{}) error {
		if value, ok := data[from]; ok {
			data[to] = value
			delete(data, from)
		}
		return nil
	})
	return b
}

// AddField adds a new field with a default value.
func (b *MigrationBuilder) AddField(name string, defaultValue interface{}) *MigrationBuilder {
	b.transforms = append(b.transforms, func(data map[string]interface{}) error {
		if _, ok := data[name]; !ok {
			data[name] = defaultValue
		}
		return nil
	})
	return b
}

// RemoveField removes a field.
func (b *MigrationBuilder) RemoveField(name string) *MigrationBuilder {
	b.transforms = append(b.transforms, func(data map[string]interface{}) error {
		delete(data, name)
		return nil
	})
	return b
}

// TransformField transforms a field value.
func (b *MigrationBuilder) TransformField(name string, transform func(interface{}) (interface{}, error)) *MigrationBuilder {
	b.transforms = append(b.transforms, func(data map[string]interface{}) error {
		if value, ok := data[name]; ok {
			newValue, err := transform(value)
			if err != nil {
				return fmt.Errorf("failed to transform field %s: %w", name, err)
			}
			data[name] = newValue
		}
		return nil
	})
	return b
}

// MergeFields merges multiple fields into one.
func (b *MigrationBuilder) MergeFields(fields []string, target string, merger func(map[string]interface{}) interface{}) *MigrationBuilder {
	b.transforms = append(b.transforms, func(data map[string]interface{}) error {
		fieldValues := make(map[string]interface{})
		for _, field := range fields {
			if value, ok := data[field]; ok {
				fieldValues[field] = value
			}
		}
		data[target] = merger(fieldValues)
		for _, field := range fields {
			delete(data, field)
		}
		return nil
	})
	return b
}

// SplitField splits one field into multiple.
func (b *MigrationBuilder) SplitField(source string, splitter func(interface{}) map[string]interface{}) *MigrationBuilder {
	b.transforms = append(b.transforms, func(data map[string]interface{}) error {
		if value, ok := data[source]; ok {
			for k, v := range splitter(value) {
				data[k] = v
			}
			delete(data, source)
		}
		return nil
	})
	return b
}

// Custom adds a custom transformation.
func (b *MigrationBuilder) Custom(transform func(map[string]interface{}) error) *MigrationBuilder {
	b.transforms = append(b.transforms, transform)
	return b
}

// Build builds and returns the migration.
func (b *MigrationBuilder) Build() Migration {
	b.migration.Migrate = func(data map[string]interface{}) (map[string]interface{}, error) {
		result := make(map[string]interface{})
		for k, v := range data {
			result[k] = v
		}

		for _, transform := range b.transforms {
			if err := transform(result); err != nil {
				return nil, err
			}
		}

		return result, nil
	}
	return b.migration
}

// ============================================================================
// Versioned Event Bus
// ============================================================================

// VersionedEventBus wraps an event bus with versioning support.
type VersionedEventBus struct {
	bus       EventBus
	versioner *EventVersioner
}

// NewVersionedEventBus creates a new versioned event bus.
func NewVersionedEventBus(bus EventBus, versioner *EventVersioner) *VersionedEventBus {
	return &VersionedEventBus{
		bus:       bus,
		versioner: versioner,
	}
}

// Publish publishes a versioned event.
func (b *VersionedEventBus) Publish(ctx context.Context, event *Event) error {
	versionedEvent, err := b.versioner.CreateVersionedEvent(event)
	if err != nil {
		return fmt.Errorf("failed to create versioned event: %w", err)
	}

	if err := b.versioner.ValidateEvent(versionedEvent); err != nil {
		return fmt.Errorf("event validation failed: %w", err)
	}

	// Add version to metadata
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Metadata["schema_version"] = versionedEvent.SchemaVersion.String()

	return b.bus.Publish(ctx, event)
}

// Subscribe subscribes to events with automatic version migration.
func (b *VersionedEventBus) Subscribe(ctx context.Context, eventType EventType, handler Handler) error {
	wrappedHandler := func(ctx context.Context, event *Event) error {
		// Extract version from metadata
		versionStr, ok := event.Metadata["schema_version"]
		if !ok {
			// No version, assume current
			return handler(ctx, event)
		}

		version, err := ParseVersion(versionStr)
		if err != nil {
			return fmt.Errorf("invalid schema version: %w", err)
		}

		versionedEvent := &VersionedEvent{
			Event:         event,
			SchemaVersion: version,
		}

		// Upcast to current version
		upcastedEvent, err := b.versioner.Upcast(ctx, versionedEvent)
		if err != nil {
			return fmt.Errorf("failed to upcast event: %w", err)
		}

		return handler(ctx, upcastedEvent.Event)
	}

	return b.bus.Subscribe(ctx, eventType, wrappedHandler)
}

// Close closes the underlying event bus.
func (b *VersionedEventBus) Close() error {
	return b.bus.Close()
}

// ============================================================================
// Common Event Schemas
// ============================================================================

// RegisterCommonSchemas registers common event schemas.
func RegisterCommonSchemas(registry *VersionRegistry) error {
	// Customer Created v1.0.0
	customerCreatedV1 := NewSchemaBuilder(EventCustomerCreated, NewVersion(1, 0, 0)).
		Description("Customer created event").
		Field("customer_id", FieldTypeUUID, true).
		Field("name", FieldTypeString, true).
		Field("email", FieldTypeString, true).
		Field("phone", FieldTypeString, false).
		Build()

	if err := registry.RegisterSchema(customerCreatedV1); err != nil {
		return err
	}

	// Customer Created v1.1.0 (added address)
	customerCreatedV11 := NewSchemaBuilder(EventCustomerCreated, NewVersion(1, 1, 0)).
		Description("Customer created event with address support").
		Field("customer_id", FieldTypeUUID, true).
		Field("name", FieldTypeString, true).
		Field("email", FieldTypeString, true).
		Field("phone", FieldTypeString, false).
		Field("address", FieldTypeObject, false).
		Build()

	if err := registry.RegisterSchema(customerCreatedV11); err != nil {
		return err
	}

	// Migration from v1.0.0 to v1.1.0
	customerMigration := NewMigrationBuilder(EventCustomerCreated, NewVersion(1, 0, 0), NewVersion(1, 1, 0)).
		Description("Add address field").
		AddField("address", nil).
		Build()

	if err := registry.RegisterMigration(customerMigration); err != nil {
		return err
	}

	// Lead Created v1.0.0
	leadCreatedV1 := NewSchemaBuilder(EventLeadCreated, NewVersion(1, 0, 0)).
		Description("Lead created event").
		Field("lead_id", FieldTypeUUID, true).
		Field("customer_id", FieldTypeUUID, false).
		Field("source", FieldTypeString, true).
		Field("status", FieldTypeString, true).
		Field("value", FieldTypeFloat, false).
		Build()

	if err := registry.RegisterSchema(leadCreatedV1); err != nil {
		return err
	}

	// Order Created v1.0.0
	orderCreatedV1 := NewSchemaBuilder(EventOrderCreated, NewVersion(1, 0, 0)).
		Description("Order created event").
		Field("order_id", FieldTypeUUID, true).
		Field("customer_id", FieldTypeUUID, true).
		Field("items", FieldTypeArray, true).
		Field("total", FieldTypeFloat, true).
		Field("status", FieldTypeString, true).
		Build()

	if err := registry.RegisterSchema(orderCreatedV1); err != nil {
		return err
	}

	return nil
}
