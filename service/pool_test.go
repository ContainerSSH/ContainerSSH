package service_test

import (
	"context"
	"sync"
	"testing"

    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/service"
	"github.com/stretchr/testify/assert"
)

func TestEmptyPool(t *testing.T) {
	pool := service.NewPool(service.NewLifecycleFactory(), log.NewTestLogger(t))
	lifecycle := service.NewLifecycle(pool)
	started := make(chan bool)
	stopped := make(chan bool)
	lifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		started <- true
	})
	go func() {
		err := lifecycle.Run()
		if err != nil {
			t.Fail()
		}
		stopped <- true
	}()
	<-started
	lifecycle.Stop(context.Background())
	<-stopped
}

func TestOneService(t *testing.T) {
	pool := service.NewPool(service.NewLifecycleFactory(), log.NewTestLogger(t))
	poolLifecycle := service.NewLifecycle(pool)
	poolStarted := make(chan bool)
	poolStopped := make(chan bool)
	var poolStates []service.State
	var serviceStates []service.State
	poolLifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		poolStarted <- true
	})
	poolLifecycle.OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		poolStates = append(poolStates, state)
	})

	s := newTestService("Test service")
	pool.Add(s).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		serviceStates = append(serviceStates, state)
	})

	go func() {
		err := poolLifecycle.Run()
		if err != nil {
			t.Fail()
		}
		poolStopped <- true
	}()

	<-poolStarted
	poolLifecycle.Stop(context.Background())
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateStopping,
		service.StateStopped,
	}, serviceStates)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateStopping,
		service.StateStopped,
	}, poolStates)
}

func TestOneServiceCrash(t *testing.T) {
	testLock := &sync.Mutex{}
	pool := service.NewPool(service.NewLifecycleFactory(), log.NewTestLogger(t))
	poolLifecycle := service.NewLifecycle(pool)
	poolStarted := make(chan bool)
	poolStopped := make(chan bool)
	var poolStates []service.State
	var serviceStates []service.State
	poolLifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		poolStarted <- true
	})
	poolLifecycle.OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		poolStates = append(poolStates, state)
	})

	s := newTestService("Test service")
	pool.Add(s).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		serviceStates = append(serviceStates, state)
	})

	go func() {
		err := poolLifecycle.Run()
		if err == nil {
			t.Fail()
		}
		poolStopped <- true
	}()

	<-poolStarted
	s.Crash()
	err := poolLifecycle.Wait()
	testLock.Lock()
	assert.NotNil(t, err)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateCrashed,
	}, serviceStates)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateCrashed,
	}, poolStates)
	testLock.Unlock()
}

func TestOneServiceStartupCrash(t *testing.T) {
	testLock := &sync.Mutex{}
	pool := service.NewPool(service.NewLifecycleFactory(), log.NewTestLogger(t))
	poolLifecycle := service.NewLifecycle(pool)
	var poolStates []service.State
	var serviceStates []service.State
	startup := make(chan struct{})
	poolLifecycle.OnStarting(func(s service.Service, l service.Lifecycle) {
		startup <- struct{}{}
	})
	poolLifecycle.OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		poolStates = append(poolStates, state)
	})

	s := newTestService("Test service")
	s.CrashStartup()
	pool.Add(s).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		serviceStates = append(serviceStates, state)
	})

	go func() {
		err := poolLifecycle.Run()
		if err == nil {
			t.Fail()
		}
	}()

	<-startup
	err := poolLifecycle.Wait()
	testLock.Lock()
	assert.NotNil(t, err)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateCrashed,
	}, serviceStates)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateStopping,
		service.StateCrashed,
	}, poolStates)
	testLock.Unlock()
}

func TestTwoServices(t *testing.T) {
	testLock := &sync.Mutex{}
	pool, poolLifecycle, poolStarted, poolStopped, poolStates, serviceStates1, serviceStates2 := setupPoolForTwoServiceTest(
		t,
		testLock,
	)

	s1 := newTestService("Test service 1")
	pool.Add(s1).OnStateChange(
		func(s service.Service, l service.Lifecycle, state service.State) {
			testLock.Lock()
			defer testLock.Unlock()
			*serviceStates1 = append(*serviceStates1, state)
		},
	)

	s2 := newTestService("Test service 2")
	pool.Add(s2).OnStateChange(
		func(s service.Service, l service.Lifecycle, state service.State) {
			testLock.Lock()
			defer testLock.Unlock()
			*serviceStates2 = append(*serviceStates2, state)
		},
	)

	go func() {
		err := poolLifecycle.Run()
		if err != nil {
			t.Fail()
		}
		poolStopped <- true
	}()

	<-poolStarted
	poolLifecycle.Stop(context.Background())

	verifyTwoServiceExecutionOrder(t, testLock, serviceStates1, serviceStates2, poolStates)
}

func verifyTwoServiceExecutionOrder(
	t *testing.T,
	testLock *sync.Mutex,
	serviceStates1 *[]service.State,
	serviceStates2 *[]service.State,
	poolStates *[]service.State,
) {
	testLock.Lock()
	defer testLock.Unlock()
	assert.Equal(
		t, []service.State{
			service.StateStarting,
			service.StateRunning,
			service.StateStopping,
			service.StateStopped,
		}, *serviceStates1,
	)
	assert.Equal(
		t, []service.State{
			service.StateStarting,
			service.StateRunning,
			service.StateStopping,
			service.StateStopped,
		}, *serviceStates2,
	)
	assert.Equal(
		t, []service.State{
			service.StateStarting,
			service.StateRunning,
			service.StateStopping,
			service.StateStopped,
		}, *poolStates,
	)
}

func setupPoolForTwoServiceTest(t *testing.T, testLock *sync.Mutex) (
	service.Pool,
	service.Lifecycle,
	chan bool,
	chan bool,
	*[]service.State,
	*[]service.State,
	*[]service.State,
) {
	pool := service.NewPool(service.NewLifecycleFactory(), log.NewTestLogger(t))
	poolLifecycle := service.NewLifecycle(pool)
	poolStarted := make(chan bool)
	poolStopped := make(chan bool)
	poolStates := &[]service.State{}
	serviceStates1 := &[]service.State{}
	serviceStates2 := &[]service.State{}
	poolLifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			poolStarted <- true
		},
	)
	poolLifecycle.OnStateChange(
		func(s service.Service, l service.Lifecycle, state service.State) {
			testLock.Lock()
			defer testLock.Unlock()

			*poolStates = append(*poolStates, state)
		},
	)
	return pool, poolLifecycle, poolStarted, poolStopped, poolStates, serviceStates1, serviceStates2
}

func TestTwoServicesOneCrashed(t *testing.T) {
	testLock := &sync.Mutex{}
	pool := service.NewPool(service.NewLifecycleFactory(), log.NewTestLogger(t))
	poolLifecycle := service.NewLifecycle(pool)
	poolStarted := make(chan bool)
	poolStopped := make(chan bool)
	var poolStates []service.State
	var serviceStates1 []service.State
	var serviceStates2 []service.State
	poolLifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		poolStarted <- true
	})
	poolLifecycle.OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		poolStates = append(poolStates, state)
	})

	s1 := newTestService("Test service 1")
	pool.Add(s1).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		serviceStates1 = append(serviceStates1, state)
	})

	s2 := newTestService("Test service 2")
	pool.Add(s2).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		serviceStates2 = append(serviceStates2, state)
	})

	go func() {
		err := poolLifecycle.Run()
		if err == nil {
			t.Fail()
		}
		poolStopped <- true
	}()

	<-poolStarted
	s1.Crash()
	<-poolStopped
	testLock.Lock()
	defer testLock.Unlock()
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateCrashed,
	}, serviceStates1)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateStopping,
		service.StateStopped,
	}, serviceStates2)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateCrashed,
	}, poolStates)
}

func TestTwoServicesOneTerminatedDuringStartup(t *testing.T) {
	testLock := &sync.Mutex{}
	pool := service.NewPool(service.NewLifecycleFactory(), log.NewTestLogger(t))
	poolLifecycle := service.NewLifecycle(pool)
	poolStarted := make(chan bool)
	poolStopped := make(chan bool)
	var poolStates []service.State
	var serviceStates1 []service.State
	var serviceStates2 []service.State
	poolLifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		poolStarted <- true
	})
	poolLifecycle.OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		poolStates = append(poolStates, state)
	})

	s1 := newTestService("Test service 1")
	s1.CrashStartup()
	pool.Add(s1).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		serviceStates1 = append(serviceStates1, state)
	})

	s2 := newTestService("Test service 2")
	pool.Add(s2).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		testLock.Lock()
		defer testLock.Unlock()
		serviceStates2 = append(serviceStates2, state)
	})

	go func() {
		err := poolLifecycle.Run()
		if err == nil {
			t.Fail()
		}
		poolStopped <- true
	}()

	<-poolStopped
	testLock.Lock()
	defer testLock.Unlock()
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateCrashed,
	}, serviceStates1)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateStopping,
		service.StateStopped,
	}, serviceStates2)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateStopping,
		service.StateCrashed,
	}, poolStates)
}
