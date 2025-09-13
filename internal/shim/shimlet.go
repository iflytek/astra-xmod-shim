package shim

// Shimlet is the interface that must be implemented by all shimlets.
type Shimlet interface {
	Init()
	CreateResource()
	UpdateResource()
	DeleteResource()
	GetResourceStatusById()
}
