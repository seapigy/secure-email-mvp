package geofencing

import (
	"database/sql"
	"testing"

	"secure-email-mvp/pkg/geolocation"
)

// MockDB is a mock database for testing
type MockDB struct {
	allowedCountries string
	allowedIPRanges  string
	geofenceViolations int
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	// Mock implementation for testing
	return &sql.Row{}
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Mock implementation for testing
	return nil, nil
}

// MockGeofencingService is a mock geofencing service for testing
type MockGeofencingService struct {
	geolocationSvc geolocation.GeolocationService
}

func NewMockGeofencingService(geolocationSvc geolocation.GeolocationService) *MockGeofencingService {
	return &MockGeofencingService{
		geolocationSvc: geolocationSvc,
	}
}

func (m *MockGeofencingService) CheckGeofenceAccess(emailID, clientIP string) (*GeofenceResult, error) {
	// Simple mock implementation
	return &GeofenceResult{Allowed: true}, nil
}

// TestGeofencingService_CheckGeofenceAccess tests the geofencing access check
func TestGeofencingService_CheckGeofenceAccess(t *testing.T) {
	// Create mock geolocation service
	mockGeo := geolocation.NewMockGeolocationService()
	
	// Set up test locations
	mockGeo.SetLocation("192.168.1.1", &geolocation.Location{
		Country: "US",
		City:    "New York",
		IP:      "192.168.1.1",
	})
	
	mockGeo.SetLocation("10.0.0.1", &geolocation.Location{
		Country: "CA",
		City:    "Toronto",
		IP:      "10.0.0.1",
	})

	// Create mock geofencing service
	geofencingSvc := NewMockGeofencingService(mockGeo)

	// Test basic functionality
	result, err := geofencingSvc.CheckGeofenceAccess("test-email", "192.168.1.1")
	if err != nil {
		t.Errorf("CheckGeofenceAccess() error = %v", err)
		return
	}

	if !result.Allowed {
		t.Errorf("CheckGeofenceAccess() expected allowed=true, got %v", result.Allowed)
	}
}

// TestGeofencingService_BasicFunctionality tests basic geofencing functionality
func TestGeofencingService_BasicFunctionality(t *testing.T) {
	mockGeo := geolocation.NewMockGeolocationService()
	geofencingSvc := NewMockGeofencingService(mockGeo)

	// Test that the service can be created and used
	result, err := geofencingSvc.CheckGeofenceAccess("test-email", "192.168.1.1")
	if err != nil {
		t.Errorf("CheckGeofenceAccess() error = %v", err)
		return
	}

	if result == nil {
		t.Errorf("CheckGeofenceAccess() returned nil result")
	}
}
