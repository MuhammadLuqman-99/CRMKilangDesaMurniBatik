// Package domain contains the domain layer for the IAM service.
package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Permission represents a permission value object.
// Permissions follow the pattern: resource:action (e.g., "users:read", "customers:*")
type Permission struct {
	resource string
	action   string
}

// Common permission actions
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"
	ActionAll    = "*"
)

// Common permission resources
const (
	ResourceUsers         = "users"
	ResourceRoles         = "roles"
	ResourceTenants       = "tenants"
	ResourceCustomers     = "customers"
	ResourceContacts      = "contacts"
	ResourceCompanies     = "companies"
	ResourceLeads         = "leads"
	ResourceOpportunities = "opportunities"
	ResourceDeals         = "deals"
	ResourcePipelines     = "pipelines"
	ResourceStages        = "stages"
	ResourceActivities    = "activities"
	ResourceNotes         = "notes"
	ResourceTasks         = "tasks"
	ResourceSettings      = "settings"
	ResourceReports       = "reports"
	ResourceAll           = "*"
)

// Permission validation regex
var permissionRegex = regexp.MustCompile(`^[a-z_*]+:[a-z_*]+$`)

// NewPermission creates a new Permission value object.
func NewPermission(resource, action string) (Permission, error) {
	resource = strings.ToLower(strings.TrimSpace(resource))
	action = strings.ToLower(strings.TrimSpace(action))

	if resource == "" {
		return Permission{}, ErrInvalidPermissionResource
	}

	if action == "" {
		return Permission{}, ErrInvalidPermissionAction
	}

	p := Permission{
		resource: resource,
		action:   action,
	}

	if !p.IsValid() {
		return Permission{}, ErrInvalidPermissionFormat
	}

	return p, nil
}

// MustNewPermission creates a new Permission or panics if invalid.
func MustNewPermission(resource, action string) Permission {
	p, err := NewPermission(resource, action)
	if err != nil {
		panic(err)
	}
	return p
}

// ParsePermission parses a permission string (e.g., "users:read").
func ParsePermission(s string) (Permission, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Handle wildcard "*" permission
	if s == "*" {
		return Permission{resource: "*", action: "*"}, nil
	}

	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Permission{}, ErrInvalidPermissionFormat
	}

	return NewPermission(parts[0], parts[1])
}

// MustParsePermission parses a permission string or panics if invalid.
func MustParsePermission(s string) Permission {
	p, err := ParsePermission(s)
	if err != nil {
		panic(err)
	}
	return p
}

// Resource returns the permission resource.
func (p Permission) Resource() string {
	return p.resource
}

// Action returns the permission action.
func (p Permission) Action() string {
	return p.action
}

// String returns the string representation of the permission.
func (p Permission) String() string {
	if p.resource == "*" && p.action == "*" {
		return "*"
	}
	return fmt.Sprintf("%s:%s", p.resource, p.action)
}

// IsValid validates the permission format.
func (p Permission) IsValid() bool {
	if p.resource == "*" && p.action == "*" {
		return true
	}
	return permissionRegex.MatchString(p.String())
}

// IsWildcard returns true if this is a full wildcard permission.
func (p Permission) IsWildcard() bool {
	return p.resource == "*" && p.action == "*"
}

// HasWildcardResource returns true if the resource is a wildcard.
func (p Permission) HasWildcardResource() bool {
	return p.resource == "*"
}

// HasWildcardAction returns true if the action is a wildcard.
func (p Permission) HasWildcardAction() bool {
	return p.action == "*"
}

// Implies checks if this permission implies (grants) another permission.
// For example, "users:*" implies "users:read", and "*" implies everything.
func (p Permission) Implies(other Permission) bool {
	// Full wildcard implies everything
	if p.IsWildcard() {
		return true
	}

	// Check resource match
	resourceMatch := p.resource == other.resource || p.resource == "*"

	// Check action match
	actionMatch := p.action == other.action || p.action == "*"

	return resourceMatch && actionMatch
}

// Equals checks if two permissions are equal.
func (p Permission) Equals(other Permission) bool {
	return p.resource == other.resource && p.action == other.action
}

// PermissionSet represents a set of permissions.
type PermissionSet struct {
	permissions map[string]Permission
}

// NewPermissionSet creates a new empty permission set.
func NewPermissionSet() *PermissionSet {
	return &PermissionSet{
		permissions: make(map[string]Permission),
	}
}

// NewPermissionSetFromStrings creates a permission set from string slice.
func NewPermissionSetFromStrings(perms []string) (*PermissionSet, error) {
	ps := NewPermissionSet()
	for _, s := range perms {
		p, err := ParsePermission(s)
		if err != nil {
			return nil, fmt.Errorf("invalid permission '%s': %w", s, err)
		}
		ps.Add(p)
	}
	return ps, nil
}

// Add adds a permission to the set.
func (ps *PermissionSet) Add(p Permission) {
	ps.permissions[p.String()] = p
}

// Remove removes a permission from the set.
func (ps *PermissionSet) Remove(p Permission) {
	delete(ps.permissions, p.String())
}

// Contains checks if the set contains a specific permission.
func (ps *PermissionSet) Contains(p Permission) bool {
	_, exists := ps.permissions[p.String()]
	return exists
}

// HasPermission checks if the set grants the given permission.
// This considers wildcard permissions.
func (ps *PermissionSet) HasPermission(required Permission) bool {
	for _, p := range ps.permissions {
		if p.Implies(required) {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the set grants any of the given permissions.
func (ps *PermissionSet) HasAnyPermission(required ...Permission) bool {
	for _, r := range required {
		if ps.HasPermission(r) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the set grants all of the given permissions.
func (ps *PermissionSet) HasAllPermissions(required ...Permission) bool {
	for _, r := range required {
		if !ps.HasPermission(r) {
			return false
		}
	}
	return true
}

// List returns all permissions as a slice.
func (ps *PermissionSet) List() []Permission {
	result := make([]Permission, 0, len(ps.permissions))
	for _, p := range ps.permissions {
		result = append(result, p)
	}
	return result
}

// Strings returns all permissions as string slice.
func (ps *PermissionSet) Strings() []string {
	result := make([]string, 0, len(ps.permissions))
	for _, p := range ps.permissions {
		result = append(result, p.String())
	}
	return result
}

// Len returns the number of permissions in the set.
func (ps *PermissionSet) Len() int {
	return len(ps.permissions)
}

// IsEmpty returns true if the set has no permissions.
func (ps *PermissionSet) IsEmpty() bool {
	return len(ps.permissions) == 0
}

// Merge merges another permission set into this one.
func (ps *PermissionSet) Merge(other *PermissionSet) {
	for _, p := range other.permissions {
		ps.Add(p)
	}
}

// Clone creates a copy of the permission set.
func (ps *PermissionSet) Clone() *PermissionSet {
	clone := NewPermissionSet()
	for _, p := range ps.permissions {
		clone.Add(p)
	}
	return clone
}

// Predefined permissions for convenience
var (
	// User permissions
	PermissionUsersCreate = MustNewPermission(ResourceUsers, ActionCreate)
	PermissionUsersRead   = MustNewPermission(ResourceUsers, ActionRead)
	PermissionUsersUpdate = MustNewPermission(ResourceUsers, ActionUpdate)
	PermissionUsersDelete = MustNewPermission(ResourceUsers, ActionDelete)
	PermissionUsersList   = MustNewPermission(ResourceUsers, ActionList)
	PermissionUsersAll    = MustNewPermission(ResourceUsers, ActionAll)

	// Role permissions
	PermissionRolesCreate = MustNewPermission(ResourceRoles, ActionCreate)
	PermissionRolesRead   = MustNewPermission(ResourceRoles, ActionRead)
	PermissionRolesUpdate = MustNewPermission(ResourceRoles, ActionUpdate)
	PermissionRolesDelete = MustNewPermission(ResourceRoles, ActionDelete)
	PermissionRolesList   = MustNewPermission(ResourceRoles, ActionList)
	PermissionRolesAll    = MustNewPermission(ResourceRoles, ActionAll)

	// Customer permissions
	PermissionCustomersCreate = MustNewPermission(ResourceCustomers, ActionCreate)
	PermissionCustomersRead   = MustNewPermission(ResourceCustomers, ActionRead)
	PermissionCustomersUpdate = MustNewPermission(ResourceCustomers, ActionUpdate)
	PermissionCustomersDelete = MustNewPermission(ResourceCustomers, ActionDelete)
	PermissionCustomersList   = MustNewPermission(ResourceCustomers, ActionList)
	PermissionCustomersAll    = MustNewPermission(ResourceCustomers, ActionAll)

	// Lead permissions
	PermissionLeadsCreate = MustNewPermission(ResourceLeads, ActionCreate)
	PermissionLeadsRead   = MustNewPermission(ResourceLeads, ActionRead)
	PermissionLeadsUpdate = MustNewPermission(ResourceLeads, ActionUpdate)
	PermissionLeadsDelete = MustNewPermission(ResourceLeads, ActionDelete)
	PermissionLeadsList   = MustNewPermission(ResourceLeads, ActionList)
	PermissionLeadsAll    = MustNewPermission(ResourceLeads, ActionAll)

	// Opportunity permissions
	PermissionOpportunitiesCreate = MustNewPermission(ResourceOpportunities, ActionCreate)
	PermissionOpportunitiesRead   = MustNewPermission(ResourceOpportunities, ActionRead)
	PermissionOpportunitiesUpdate = MustNewPermission(ResourceOpportunities, ActionUpdate)
	PermissionOpportunitiesDelete = MustNewPermission(ResourceOpportunities, ActionDelete)
	PermissionOpportunitiesList   = MustNewPermission(ResourceOpportunities, ActionList)
	PermissionOpportunitiesAll    = MustNewPermission(ResourceOpportunities, ActionAll)

	// Settings permissions
	PermissionSettingsRead   = MustNewPermission(ResourceSettings, ActionRead)
	PermissionSettingsUpdate = MustNewPermission(ResourceSettings, ActionUpdate)
	PermissionSettingsAll    = MustNewPermission(ResourceSettings, ActionAll)

	// Full access permission
	PermissionFullAccess = MustNewPermission(ResourceAll, ActionAll)
)

// Permission errors
var (
	ErrInvalidPermissionResource = fmt.Errorf("invalid permission resource")
	ErrInvalidPermissionAction   = fmt.Errorf("invalid permission action")
	ErrInvalidPermissionFormat   = fmt.Errorf("invalid permission format, expected 'resource:action'")
)
