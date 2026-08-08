package actuator

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/cloud/core"
)

type mockIndicator struct {
	status core.HealthStatus
}

func (m *mockIndicator) Health() core.HealthStatus {
	return m.status
}

func TestActuator(t *testing.T) {
	a := NewActuator()
	a.AddIndicator(&mockIndicator{status: core.Up})
	
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	
	a.HealthHandler(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"status":"UP"}`, rr.Body.String())
	
	a.AddIndicator(&mockIndicator{status: core.Down})
	
	rr2 := httptest.NewRecorder()
	a.HealthHandler(rr2, req)
	
	assert.Equal(t, http.StatusServiceUnavailable, rr2.Code)
	assert.Equal(t, `{"status":"DOWN"}`, rr2.Body.String())
}
