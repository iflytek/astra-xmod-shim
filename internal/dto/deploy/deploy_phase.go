package dto

// DeployPhase Deployment status enumeration (recommended to place in a separate status.go file)
// DeployPhase defines all possible states (phases) of deployment
type DeployPhase string

const (
	PhasePending     DeployPhase = "PENDING"     // Initializing, waiting for scheduling
	PhaseCreating    DeployPhase = "CREATING"    // Resource creation in progress
	PhaseRunning     DeployPhase = "RUNNING"     // Deployment successful, running normally
	PhaseUpdating    DeployPhase = "UPDATING"    // Configuration update in progress
	PhaseFailed      DeployPhase = "FAILED"      // Deployment failed
	PhaseTerminating DeployPhase = "TERMINATING" // Terminating/deleting in progress
	PhaseTerminated  DeployPhase = "TERMINATED"  // Terminated/deleted
)