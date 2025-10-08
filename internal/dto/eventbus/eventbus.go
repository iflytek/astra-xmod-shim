package eventbus

import dto "astron-xmod-shim/internal/dto/deploy"

// ServiceEvent represents a state change event
type ServiceEvent struct {
	ServiceID string
	To        dto.DeployPhase
	EndPoint  string
}
