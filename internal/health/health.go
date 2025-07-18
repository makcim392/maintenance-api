package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/makcim392/maintenance-api/internal/logger"
)

// HealthChecker manages health checks for the application
type HealthChecker struct {
	db     *sql.DB
	logger *logger.Logger
	checks map[string]CheckFunc
	mu     sync.RWMutex
}

// CheckFunc defines the signature for health check functions
type CheckFunc func(context.Context) error

// Status represents the overall health status
type Status struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]Check `json:"checks"`
}

// Check represents the result of a single health check
type Check struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// New creates a new HealthChecker instance
func New(db *sql.DB, logger *logger.Logger) *HealthChecker {
	hc := &HealthChecker{
		db:     db,
		logger: logger,
		checks: make(map[string]CheckFunc),
	}

	// Register default health checks
	hc.RegisterCheck("database", hc.checkDatabase)
	hc.RegisterCheck("readiness", hc.checkReadiness)

	return hc
}

// RegisterCheck adds a new health check
func (hc *HealthChecker) RegisterCheck(name string, check CheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// HealthHandler handles health check requests
func (hc *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	status := hc.performChecks(ctx)
	
	w.Header().Set("Content-Type", "application/json")
	
	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(status)
}

// ReadinessHandler handles readiness probe requests
func (hc *HealthChecker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Only check critical dependencies for readiness
	checks := map[string]CheckFunc{
		"database": hc.checkDatabase,
	}
	
	status := hc.performSpecificChecks(ctx, checks)
	
	w.Header().Set("Content-Type", "application/json")
	
	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(status)
}

// LivenessHandler handles liveness probe requests
func (hc *HealthChecker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Status:    "alive",
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]Check),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// performChecks runs all registered health checks
func (hc *HealthChecker) performChecks(ctx context.Context) Status {
	hc.mu.RLock()
	checks := make(map[string]CheckFunc)
	for k, v := range hc.checks {
		checks[k] = v
	}
	hc.mu.RUnlock()
	
	return hc.performSpecificChecks(ctx, checks)
}

// performSpecificChecks runs specific health checks
func (hc *HealthChecker) performSpecificChecks(ctx context.Context, checks map[string]CheckFunc) Status {
	status := Status{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]Check),
	}
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for name, check := range checks {
		wg.Add(1)
		go func(n string, c CheckFunc) {
			defer wg.Done()
			
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			
			err := c(checkCtx)
			
			mu.Lock()
			defer mu.Unlock()
			
			checkResult := Check{
				Timestamp: time.Now().UTC(),
			}
			
			if err != nil {
				checkResult.Status = "unhealthy"
				checkResult.Error = err.Error()
				checkResult.Message = "Check failed"
				status.Status = "unhealthy"
			} else {
				checkResult.Status = "healthy"
				checkResult.Message = "OK"
			}
			
			status.Checks[n] = checkResult
		}(name, check)
	}
	
	wg.Wait()
	
	return status
}

// checkDatabase checks database connectivity
func (hc *HealthChecker) checkDatabase(ctx context.Context) error {
	if hc.db == nil {
		return nil // Skip if no database configured
	}
	
	return hc.db.PingContext(ctx)
}

// checkReadiness checks if the application is ready to serve requests
func (hc *HealthChecker) checkReadiness(ctx context.Context) error {
	// Add any readiness checks here
	// For example: check if required services are available
	return nil
}

// StartBackgroundChecks starts periodic health checks in the background
func (hc *HealthChecker) StartBackgroundChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.performChecks(ctx)
		}
	}
}