package service

import (
	"context"
	"fmt"
	"sync"
)

type lifecycle struct {
	service           Service
	state             State
	mutex             *sync.Mutex
	runningContext    context.Context
	cancelRun         func()
	shutdownContext   context.Context
	lastError         error
	waitContext       context.Context
	cancelWaitContext func()

	onStateChange []func(s Service, l Lifecycle, state State)
	onStarting    []func(s Service, l Lifecycle)
	onRunning     []func(s Service, l Lifecycle)
	onStopping    []func(s Service, l Lifecycle, shutdownContext context.Context)
	onStopped     []func(s Service, l Lifecycle)
	onCrashed     []func(s Service, l Lifecycle, err error)
}

func (l *lifecycle) Context() context.Context {
	return l.runningContext
}

func (l *lifecycle) ShouldStop() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	select {
	case <-l.runningContext.Done():
		return true
	default:
		return false
	}
}

func (l *lifecycle) ShutdownContext() context.Context {
	return l.shutdownContext
}

func (l *lifecycle) State() State {
	return l.state
}

func (l *lifecycle) Wait() error {
	l.mutex.Lock()
	if l.state == StateCrashed {
		return l.lastError
	}
	if l.state == StateStopped {
		return nil
	}
	waitContext := l.waitContext
	l.mutex.Unlock()
	if waitContext != nil {
		<-waitContext.Done()
	}
	return l.lastError
}

func (l *lifecycle) Error() error {
	return l.lastError
}

func (l *lifecycle) Stop(shutdownContext context.Context) {
	l.mutex.Lock()
	if l.state == StateStopping || l.state == StateStopped || l.state == StateCrashed {
		l.mutex.Unlock()
		return
	}
	l.shutdownContext = shutdownContext
	l.mutex.Unlock()
	l.cancelRun()
	_ = l.Wait()
}

func (l *lifecycle) Run() error {
	defer func() {
		if crash := recover(); crash != nil {
			l.crashed(fmt.Errorf("service paniced (%v)", crash))
		}
		l.cancelWaitContext()
	}()

	l.lastError = nil
	l.waitContext, l.cancelWaitContext = context.WithCancel(context.Background())
	l.starting()
	err := l.service.RunWithLifecycle(l)
	if err != nil {
		l.crashed(err)
		return err
	}
	l.stopped()
	return nil
}

func (l *lifecycle) callSimpleHook(hooks []func(s Service, l Lifecycle)) {
	wg := &sync.WaitGroup{}
	wg.Add(len(hooks))
	for _, hook := range hooks {
		handler := hook
		go func() {
			defer wg.Done()
			handler(l.service, l)
		}()
	}
	wg.Wait()
}

func (l *lifecycle) stateChange(state State) {
	wg := &sync.WaitGroup{}
	stateChangeHandlers := l.onStateChange
	wg.Add(len(stateChangeHandlers))
	for _, hook := range stateChangeHandlers {
		handler := hook
		go func() {
			defer wg.Done()
			handler(l.service, l, state)
		}()
	}
	wg.Wait()
}

func (l *lifecycle) starting() {
	l.mutex.Lock()

	l.state = StateStarting
	l.mutex.Unlock()
	l.stateChange(StateStarting)
	l.callSimpleHook(l.onStarting)
}

func (l *lifecycle) Running() {
	l.mutex.Lock()
	l.state = StateRunning
	l.mutex.Unlock()
	l.stateChange(StateRunning)
	l.callSimpleHook(l.onRunning)
}

func (l *lifecycle) Stopping() context.Context {
	l.mutex.Lock()
	if l.shutdownContext == nil {
		l.shutdownContext = context.Background()
	}
	shutdownContext := l.shutdownContext

	l.state = StateStopping
	handlers := l.onStopping
	l.mutex.Unlock()

	l.stateChange(StateStopping)
	wg := &sync.WaitGroup{}
	wg.Add(len(handlers))
	for _, onShutdown := range handlers {
		shutdownHandler := onShutdown
		go func() {
			defer wg.Done()
			shutdownHandler(l.service, l, shutdownContext)
		}()
	}
	wg.Wait()
	return shutdownContext
}

func (l *lifecycle) stopped() {
	l.mutex.Lock()
	l.state = StateStopped
	l.mutex.Unlock()

	l.stateChange(StateStopped)
	l.callSimpleHook(l.onStopped)
}

func (l *lifecycle) crashed(err error) {
	l.mutex.Lock()
	l.lastError = err
	l.state = StateCrashed
	l.mutex.Unlock()

	l.stateChange(StateCrashed)
	wg := &sync.WaitGroup{}
	wg.Add(len(l.onCrashed))
	for _, onCrashed := range l.onCrashed {
		crashHandler := onCrashed
		go func() {
			defer wg.Done()
			crashHandler(l.service, l, err)
		}()
	}
	wg.Wait()
}

func (l *lifecycle) OnStateChange(f func(s Service, l Lifecycle, state State)) Lifecycle {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onStateChange = append(l.onStateChange, f)
	return l
}

func (l *lifecycle) OnStarting(f func(s Service, l Lifecycle)) Lifecycle {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onStarting = append(l.onStarting, f)
	return l
}

func (l *lifecycle) OnRunning(f func(s Service, l Lifecycle)) Lifecycle {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onRunning = append(l.onRunning, f)
	return l
}

func (l *lifecycle) OnStopping(f func(s Service, l Lifecycle, shutdownContext context.Context)) Lifecycle {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onStopping = append(l.onStopping, f)
	return l
}

func (l *lifecycle) OnStopped(f func(s Service, l Lifecycle)) Lifecycle {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onStopped = append(l.onStopped, f)
	return l
}

func (l *lifecycle) OnCrashed(f func(s Service, l Lifecycle, err error)) Lifecycle {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.onCrashed = append(l.onCrashed, f)
	return l
}
