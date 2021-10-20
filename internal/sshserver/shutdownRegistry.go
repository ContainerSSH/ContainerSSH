package sshserver

import (
	"context"
	"sync"
)

type shutdownRegistry struct {
	lock      *sync.Mutex
	callbacks map[string]shutdownHandler
}

func (s *shutdownRegistry) Register(key string, handler shutdownHandler) {
	s.lock.Lock()
	s.callbacks[key] = handler
	s.lock.Unlock()
}

func (s *shutdownRegistry) Unregister(key string) {
	s.lock.Lock()
	delete(s.callbacks, key)
	s.lock.Unlock()
}

func (s *shutdownRegistry) Shutdown(shutdownContext context.Context) {
	wg := &sync.WaitGroup{}
	s.lock.Lock()
	wg.Add(len(s.callbacks))
	for _, handler := range s.callbacks {
		h := handler
		go func() {
			defer wg.Done()
			h.OnShutdown(shutdownContext)
		}()
	}
	s.lock.Unlock()
	wg.Wait()
}
