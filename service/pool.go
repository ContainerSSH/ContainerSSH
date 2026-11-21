package service

// Pool is a handler for multiple services at once. It will run services in parallel in goroutines and terminate all
//      services once a single one has exited.
type Pool interface {
	Service

	// Add adds a service to the service pool.
	Add(s Service) Lifecycle
}
