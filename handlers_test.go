package vent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockSchemaClient is a mock implementation of SchemaClient for testing
type MockSchemaClient struct {
	ListFunc               func(ctx context.Context, opts ListOptions) ([]EntityData, error)
	GetFunc                func(ctx context.Context, id int) (EntityData, error)
	CreateFunc             func(ctx context.Context, data map[string]any) (EntityData, error)
	UpdateFunc             func(ctx context.Context, id int, data map[string]any) error
	DeleteFunc             func(ctx context.Context, id int) error
	GetRelationOptionsFunc func(ctx context.Context, relation *RelationDef) ([]SelectOption, error)
}

func (m *MockSchemaClient) List(ctx context.Context, opts ListOptions) ([]EntityData, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, opts)
	}
	return nil, nil
}

func (m *MockSchemaClient) Get(ctx context.Context, id int) (EntityData, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSchemaClient) Create(ctx context.Context, data map[string]any) (EntityData, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, data)
	}
	return nil, nil
}

func (m *MockSchemaClient) Update(ctx context.Context, id int, data map[string]any) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, data)
	}
	return nil
}

func (m *MockSchemaClient) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSchemaClient) GetRelationOptions(ctx context.Context, relation *RelationDef) ([]SelectOption, error) {
	if m.GetRelationOptionsFunc != nil {
		return m.GetRelationOptionsFunc(ctx, relation)
	}
	return nil, nil
}

// MockSecretProvider implements auth.SecretProvider for testing
type mockSecretProvider struct{}

func (m mockSecretProvider) Secret() []byte {
	return []byte("test-secret")
}

// MockCredentialAuthenticator implements auth.CredentialAuthenticator for testing
type mockCredentialAuthenticator struct {
	shouldSucceed bool
}

func (m mockCredentialAuthenticator) Authenticate(password string, hash string) error {
	if m.shouldSucceed {
		return nil
	}
	return ErrPasswordMismatch
}

func TestEntityData_ID(t *testing.T) {
	entity := EntityData{
		"id":    NewIntFieldValue(42),
		"email": NewStringFieldValue("test@example.com"),
	}

	if entity.ID() != 42 {
		t.Errorf("expected ID 42, got %d", entity.ID())
	}
}

func TestEntityData_GetString(t *testing.T) {
	entity := EntityData{
		"id":    NewIntFieldValue(1),
		"email": NewStringFieldValue("test@example.com"),
		"name":  NewStringFieldValue("Test User"),
	}

	if entity.GetString("email") != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", entity.GetString("email"))
	}

	if entity.GetString("nonexistent") != "" {
		t.Errorf("expected empty string for nonexistent field, got '%s'", entity.GetString("nonexistent"))
	}
}

func TestEntityData_GetBool(t *testing.T) {
	entity := EntityData{
		"id":        NewIntFieldValue(1),
		"is_active": NewBoolFieldValue(true),
		"is_admin":  NewBoolFieldValue(false),
	}

	if !entity.GetBool("is_active") {
		t.Error("expected is_active to be true")
	}

	if entity.GetBool("is_admin") {
		t.Error("expected is_admin to be false")
	}

	if entity.GetBool("nonexistent") {
		t.Error("expected false for nonexistent field")
	}
}

func TestFieldValue_String(t *testing.T) {
	tests := []struct {
		name     string
		field    FieldValue
		expected string
	}{
		{"string field", NewStringFieldValue("hello"), "hello"},
		{"int field", NewIntFieldValue(42), "42"},
		{"bool true", NewBoolFieldValue(true), "true"},
		{"bool false", NewBoolFieldValue(false), "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.String() != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, tt.field.String())
			}
		})
	}
}

func TestFieldValue_ForeignKey(t *testing.T) {
	relation := RelationValue{
		TargetSchema: "User",
		TargetID:     42,
		TargetLabel:  "admin@example.com",
		TargetPath:   "/admin/users/",
	}
	field := NewForeignKeyFieldValue(42, relation)

	if field.Type != TypeForeignKey {
		t.Errorf("expected TypeForeignKey, got %v", field.Type)
	}

	if field.IntValue() != 42 {
		t.Errorf("expected int value 42, got %d", field.IntValue())
	}

	if field.Display != "admin@example.com" {
		t.Errorf("expected display 'admin@example.com', got '%s'", field.Display)
	}

	if field.Relation == nil {
		t.Fatal("expected relation to be set")
	}

	if field.Relation.TargetSchema != "User" {
		t.Errorf("expected target schema 'User', got '%s'", field.Relation.TargetSchema)
	}
}

func TestSchemaConfig_Path(t *testing.T) {
	schema := SchemaConfig{
		Name:      "User",
		AdminPath: "/admin/",
	}

	expected := "/admin/users/"
	if schema.Path() != expected {
		t.Errorf("expected path '%s', got '%s'", expected, schema.Path())
	}
}

func TestHandler_GetLoginPage(t *testing.T) {
	// Create mock auth user schema
	mockClient := &MockSchemaClient{}
	authUserSchema := SchemaConfig{
		Name:   "User",
		Client: mockClient,
		Columns: []ColumnConfig{
			{Name: "id", Label: "ID", Type: TypeInt},
			{Name: "email", Label: "Email", Type: TypeString},
		},
	}

	handler := NewHandler(HandlerConfig{
		SecretProvider:          mockSecretProvider{},
		CredentialAuthenticator: mockCredentialAuthenticator{},
		AuthUserSchema:          authUserSchema,
		AuthPasswordField:       "password_hash",
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/login/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Check that the response contains login form elements
	body := rec.Body.String()
	if !strings.Contains(body, "email") {
		t.Error("expected login page to contain email field")
	}
}

func TestHandler_RedirectToLoginWhenUnauthorized(t *testing.T) {
	mockClient := &MockSchemaClient{}
	authUserSchema := SchemaConfig{
		Name:   "User",
		Client: mockClient,
	}

	handler := NewHandler(HandlerConfig{
		SecretProvider:          mockSecretProvider{},
		CredentialAuthenticator: mockCredentialAuthenticator{},
		AuthUserSchema:          authUserSchema,
		AuthPasswordField:       "password_hash",
	})

	// Try to access admin without auth
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should redirect to login
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect status 303, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/admin/login/" {
		t.Errorf("expected redirect to /admin/login/, got %s", location)
	}
}

func TestMockSchemaClient_List(t *testing.T) {
	mockClient := &MockSchemaClient{
		ListFunc: func(ctx context.Context, opts ListOptions) ([]EntityData, error) {
			return []EntityData{
				{
					"id":    NewIntFieldValue(1),
					"name":  NewStringFieldValue("Alice"),
					"email": NewStringFieldValue("alice@example.com"),
				},
				{
					"id":    NewIntFieldValue(2),
					"name":  NewStringFieldValue("Bob"),
					"email": NewStringFieldValue("bob@example.com"),
				},
			}, nil
		},
	}

	entities, err := mockClient.List(context.Background(), ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(entities))
	}

	if entities[0].GetString("name") != "Alice" {
		t.Errorf("expected first entity name 'Alice', got '%s'", entities[0].GetString("name"))
	}

	if entities[1].ID() != 2 {
		t.Errorf("expected second entity ID 2, got %d", entities[1].ID())
	}
}

func TestMockSchemaClient_Get(t *testing.T) {
	mockClient := &MockSchemaClient{
		GetFunc: func(ctx context.Context, id int) (EntityData, error) {
			return EntityData{
				"id":        NewIntFieldValue(id),
				"name":      NewStringFieldValue("Test User"),
				"is_active": NewBoolFieldValue(true),
			}, nil
		},
	}

	entity, err := mockClient.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.ID() != 42 {
		t.Errorf("expected ID 42, got %d", entity.ID())
	}

	if entity.GetString("name") != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", entity.GetString("name"))
	}
}

func TestMockSchemaClient_Update(t *testing.T) {
	var capturedID int
	var capturedData map[string]any

	mockClient := &MockSchemaClient{
		UpdateFunc: func(ctx context.Context, id int, data map[string]any) error {
			capturedID = id
			capturedData = data
			return nil
		},
	}

	err := mockClient.Update(context.Background(), 42, map[string]any{
		"name":      "Updated Name",
		"is_active": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedID != 42 {
		t.Errorf("expected captured ID 42, got %d", capturedID)
	}

	if capturedData["name"] != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%v'", capturedData["name"])
	}
}

func TestMockSchemaClient_Delete(t *testing.T) {
	var deletedID int

	mockClient := &MockSchemaClient{
		DeleteFunc: func(ctx context.Context, id int) error {
			deletedID = id
			return nil
		},
	}

	err := mockClient.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deletedID != 42 {
		t.Errorf("expected deleted ID 42, got %d", deletedID)
	}
}

func TestMockSchemaClient_GetRelationOptions(t *testing.T) {
	mockClient := &MockSchemaClient{
		GetRelationOptionsFunc: func(ctx context.Context, relation *RelationDef) ([]SelectOption, error) {
			return []SelectOption{
				{Value: 1, Label: "Option 1"},
				{Value: 2, Label: "Option 2"},
				{Value: 3, Label: "Option 3"},
			}, nil
		},
	}

	options, err := mockClient.GetRelationOptions(context.Background(), &RelationDef{
		TargetSchema:  "User",
		TargetDisplay: "email",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(options) != 3 {
		t.Errorf("expected 3 options, got %d", len(options))
	}

	if options[0].Label != "Option 1" {
		t.Errorf("expected first option label 'Option 1', got '%s'", options[0].Label)
	}
}

func TestListOptions(t *testing.T) {
	var capturedOpts ListOptions

	mockClient := &MockSchemaClient{
		ListFunc: func(ctx context.Context, opts ListOptions) ([]EntityData, error) {
			capturedOpts = opts
			return nil, nil
		},
	}

	_, _ = mockClient.List(context.Background(), ListOptions{
		OrderBy:   "created_at",
		OrderDesc: true,
		Limit:     10,
		Offset:    20,
		Filters: map[string]any{
			"is_active": true,
			"name":      "test",
		},
	})

	if capturedOpts.OrderBy != "created_at" {
		t.Errorf("expected OrderBy 'created_at', got '%s'", capturedOpts.OrderBy)
	}

	if !capturedOpts.OrderDesc {
		t.Error("expected OrderDesc to be true")
	}

	if capturedOpts.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", capturedOpts.Limit)
	}

	if capturedOpts.Offset != 20 {
		t.Errorf("expected Offset 20, got %d", capturedOpts.Offset)
	}

	if capturedOpts.Filters["is_active"] != true {
		t.Error("expected is_active filter to be true")
	}
}
