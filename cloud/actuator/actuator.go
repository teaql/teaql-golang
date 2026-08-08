package actuator

import (
	"net/http"

	"github.com/teaql/teaql-golang/cloud/core"
)

type Actuator struct {
	indicators []core.HealthIndicator
}

func NewActuator() *Actuator {
	return &Actuator{
		indicators: make([]core.HealthIndicator, 0),
	}
}

func (a *Actuator) AddIndicator(indicator core.HealthIndicator) {
	a.indicators = append(a.indicators, indicator)
}

func (a *Actuator) HealthHandler(w http.ResponseWriter, r *http.Request) {
	status := core.Up

	for _, indicator := range a.indicators {
		if indicator.Health() == core.Down {
			status = core.Down
			break
		}
	}

	if status == core.Down {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Write([]byte(`{"status":"` + string(status) + `"}`))
}
