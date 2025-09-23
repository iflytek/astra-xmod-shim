package eventbus

import dto "modserv-shim/internal/dto/deploy"

// ServiceEvent represents a state change event
type ServiceEvent struct {
	ServiceID string
	To        dto.DeployPhase
	EndPoint  string
}
