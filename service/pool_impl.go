package service

import (
	"context"
	"sync"

    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
)

type pool struct {
	mutex            *sync.Mutex
	services         []Service
	lifecycles       map[Service]Lifecycle
	serviceStates    map[Service]State
	lifecycleFactory LifecycleFactory
	running          bool
	startupComplete  chan struct{}
	stopComplete     chan struct{}
	lastError        error
	stopping         bool
	logger           log.Logger
}

func (p *pool) String() string {
	return "Service Pool"
}

func (p *pool) Add(s Service) Lifecycle {
	p.mutex.Lock()
	if p.running {
		panic("bug: pool already running, cannot add service")
	}
	defer p.mutex.Unlock()
	l := p.lifecycleFactory.Make(s)
	l.OnStateChange(p.onServiceStateChange)
	p.serviceStates[s] = StateStopped
	p.services = append(p.services, s)
	p.lifecycles[s] = l
	return l
}

func (p *pool) RunWithLifecycle(lifecycle Lifecycle) error {
	p.mutex.Lock()
	if p.running {
		p.mutex.Unlock()
		panic("bug: pool already running, cannot run again")
	}
	p.logger.Info(message.NewMessage(message.MServicesStarting, "Services are starting..."))
	p.stopComplete = make(chan struct{}, len(p.services))
	p.running = true
	p.stopping = false
	p.mutex.Unlock()
	defer func() {
		p.mutex.Lock()
		p.running = false
		p.mutex.Unlock()
	}()

	for _, service := range p.services {
		p.runService(service)
	}

	stopped := false
	startedServices := len(p.services)
	finished := false
	for i := 0; i < len(p.services); i++ {
		select {
		case <-p.startupComplete:
		case <-p.stopComplete:
			stopped = true
			startedServices--
			finished = true
		}
		if finished {
			break
		}
	}

	startedServices = p.processRunning(lifecycle, stopped, startedServices)

	for i := 0; i < startedServices; i++ {
		<-p.stopComplete
	}
	p.logger.Info(message.NewMessage(message.MServicesStopped, "All services have stopped."))
	return p.lastError
}

func (p *pool) processRunning(lifecycle Lifecycle, stopped bool, startedServices int) int {
	if !stopped {
		p.logger.Info(message.NewMessage(message.MServicesRunning, "All services are now running."))

		lifecycle.Running()

		select {
		case <-p.stopComplete:
			// One service stopped, initiate shutdown
			p.logger.Info(message.NewMessage(message.MServicesStopping, "Services are now stopping..."))
			startedServices--
		case <-lifecycle.Context().Done():
			p.logger.Info(message.NewMessage(message.MServicesStopping, "Services are now stopping..."))
			lifecycle.Stopping()
			p.triggerStop(lifecycle.ShutdownContext())
		}
	} else {
		p.logger.Info(message.NewMessage(message.MServicesStopping, "Services are now stopping..."))
		lifecycle.Stopping()
		p.triggerStop(context.Background())
	}
	return startedServices
}

func (p *pool) runService(service Service) {
	go func() {
		_ = p.lifecycles[service].Run()
	}()
}

func (p *pool) onServiceStateChange(s Service, l Lifecycle, newState State) {
	if s == p {
		return
	}

	p.mutex.Lock()
	oldState := p.serviceStates[s]
	p.serviceStates[s] = newState
	p.mutex.Unlock()

	if oldState == newState {
		return
	}

	switch newState {
	case StateStarting:
		p.logger.Info(message.NewMessage(message.MServiceStarting, "%s is starting...", s.String()).Label("service", s.String()))
		return
	case StateRunning:
		p.logger.Info(message.NewMessage(message.MServiceRunning, "%s is running.", s.String()).Label("service", s.String()))
		select {
		case p.startupComplete <- struct{}{}:
		default:
		}
	case StateStopping:
		p.logger.Info(message.NewMessage(message.MServiceStopping, "%s is stopping...", s.String()).Label("service", s.String()))
		p.triggerStop(context.Background())
	case StateStopped:
		p.logger.Info(message.NewMessage(message.MServiceStopped, "%s has stopped.", s.String()).Label("service", s.String()))
		p.triggerStop(context.Background())
		p.stopComplete <- struct{}{}
	case StateCrashed:
		p.logger.Error(message.NewMessage(message.EServiceCrashed, "%s has crashed.", s.String()).Label("service", s.String()))
		p.lastError = l.Error()
		p.triggerStop(context.Background())
		p.stopComplete <- struct{}{}
	}
}

func (p *pool) triggerStop(shutdownContext context.Context) {
	p.mutex.Lock()
	if p.stopping {
		p.mutex.Unlock()
		return
	}
	p.stopping = true
	svc := p.services
	p.mutex.Unlock()

	wg := &sync.WaitGroup{}
	wg.Add(len(svc))
	for _, s := range svc {
		service := s
		l := p.lifecycles[service]
		go func() {
			defer wg.Done()
			l.Stop(shutdownContext)
		}()
	}
	wg.Wait()
}
