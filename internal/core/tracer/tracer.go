// Package tracer provides a service status tracker.
// It does NOT include global state or static methods.
// You control the lifecycle of the Tracer instance.
package tracer

import (
	"context"
	"modserv-shim/internal/core/eventbus"
	"modserv-shim/internal/core/shimlet"
	eventbus2 "modserv-shim/internal/dto/eventbus"
	"modserv-shim/pkg/log"
	"sync"
	"time"
)

// Tracer is a lightweight status tracker for deployed services.
type Tracer struct {
	tasks    map[string]context.CancelFunc // serviceID -> cancel
	eventBus eventbus.EventBus
	mu       sync.RWMutex
}

// New creates a new Tracer instance.
func New(eventbus eventbus.EventBus) *Tracer {
	return &Tracer{
		eventBus: eventbus,
		tasks:    make(map[string]context.CancelFunc),
	}
}

// Init initializes the Tracer and starts tracking existing services.
// You must explicitly call this at startup if you want to recover tracking.
func (t *Tracer) Init(shim shimlet.Shimlet, interval time.Duration) error {
	if shim == nil {
		return nil
	}

	serviceIDs, err := shim.ListDeployedServices()
	if err != nil {
		log.Warn("Tracer.Init: failed to list deployed services: %v", err)
	}

	for _, serviceID := range serviceIDs {
		_ = t.Trace(serviceID, shim, interval)
	}

	log.Info("Tracer.Init: recovered tracking for %d services", len(serviceIDs))
	return nil
}

// Trace starts tracking a service with periodic status checks.
// If already tracked, it returns nil (idempotent).
func (t *Tracer) Trace(serviceID string, shim shimlet.Shimlet, interval time.Duration) error {
	if serviceID == "" || shim == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.tasks[serviceID]; exists {
		log.Debug("Service already tracked, skipping: %s", serviceID)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.tasks[serviceID] = cancel

	go t.trackService(ctx, serviceID, shim, interval)
	log.Info("Started tracking service: %s", serviceID)
	return nil
}

// trackService runs in a goroutine to check service status periodically.
func (t *Tracer) trackService(ctx context.Context, serviceID string, shim shimlet.Shimlet, interval time.Duration) {
	defer func() {
		t.mu.Lock()
		delete(t.tasks, serviceID)
		t.mu.Unlock()
		log.Info("Stopped tracking service: %s", serviceID)
	}()

	if interval < time.Second {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	t.checkStatus(serviceID, shim)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.checkStatus(serviceID, shim)
		}
	}
}

// checkStatus queries and logs the current status of a service.
func (t *Tracer) checkStatus(serviceID string, shim shimlet.Shimlet) {
	status, err := shim.Status(serviceID)
	if err != nil {
		log.Error("Status check failed for %s: %v", serviceID, err)
		return
	}
	// 状态变化，更新缓存并发布事件

	t.eventBus.Publish("service.status", &eventbus2.ServiceEvent{ServiceID: serviceID, To: status.Status, EndPoint: status.EndPoint})
	log.Debug("Status of %s: %+v", serviceID, status)
}

// Stop stops tracking a specific service.
func (t *Tracer) Stop(serviceID string) {
	if serviceID == "" {
		return
	}

	t.mu.Lock()
	cancel, exists := t.tasks[serviceID]
	if exists {
		delete(t.tasks, serviceID)
	}
	t.mu.Unlock()

	if exists {
		cancel()
	}
}

// StopAll stops tracking all services.
func (t *Tracer) StopAll() {
	t.mu.Lock()
	var cancels []context.CancelFunc
	for _, cancel := range t.tasks {
		cancels = append(cancels, cancel)
	}
	t.tasks = make(map[string]context.CancelFunc)
	t.mu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}
}
