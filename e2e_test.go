package containerssh_test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
	"github.com/cucumber/godog"
)

var opts = godog.Options{
	Output: os.Stdout,
	Strict: true,
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

var factors = []TestingAspect{
	NewBackendTestingAspect(),
	NewAuthTestingAspect(),
	NewMetricsTestingAspect(),
	NewSSHTestingAspect(),
	NewAuditLogTestingAspect(),
}

func TestE2E(t *testing.T) {
	flag.Parse()
	opts.Paths = flag.Args()

	processTestingAspect(t, factors, []TestingFactor{})
}

var ErrSkipped = errors.New("SKIPPED")

func processTestingAspect(t *testing.T, aspects []TestingAspect, factors []TestingFactor) {
	if len(aspects) == 0 {
		cfg := config.AppConfig{}
		structutils.Defaults(&cfg)

		if modifyConfiguration(t, factors, &cfg) {
			return
		}

		var startedFactors = &[]TestingFactor{}
		defer stopFactors(t, startedFactors, cfg)()
		for _, factor := range factors {
			logger := log.NewTestLogger(t)
			logger.Notice(message.NewMessage(message.MTest, "Starting backing services for %s=%s...", factor.Aspect().String(), factor.String()))
			if err := factor.StartBackingServices(cfg, logger); err != nil {
				t.Errorf("failed to start backing services for %s=%s (%v)", factor.Aspect().String(), factor.String(), err)
				t.Fail()

				_ = factor.StopBackingServices(cfg, logger)
				return
			}
			logger.Notice(message.NewMessage(message.MTest, "Backing services for %s=%s running.", factor.Aspect().String(), factor.String()))
			*startedFactors = append(*startedFactors, factor)
		}

		runTestSuite(t, factors, cfg)
		return
	}

	aspect := aspects[0]
	aspectFactors := aspect.Factors()
	if len(aspectFactors) == 1 {
		processTestingAspect(t, aspects[1:], append(factors, aspectFactors[0]))
	} else {
		for _, factor := range aspectFactors {
			t.Run(
				fmt.Sprintf("%s=%s", aspect.String(), factor.String()),
				func(t *testing.T) {
					processTestingAspect(t, aspects[1:], append(factors, factor))
				},
			)
		}
	}
}

func stopFactors(
	t *testing.T,
	startedFactors *[]TestingFactor,
	cfg config.AppConfig,
) func() {
	return func() {
		if startedFactors == nil {
			return
		}
		for _, factor := range *startedFactors {
			logger := log.NewTestLogger(t)

			logger.Notice(message.NewMessage(message.MTest, "Stopping backing services for %s=%s...", factor.Aspect().String(), factor.String()))
			err := factor.StopBackingServices(cfg, logger)
			if err != nil {
				t.Errorf("failed to stop backing services for %s=%s (%v)", factor.Aspect().String(), factor.String(), err)
				t.Fail()
				return
			}
			logger.Notice(message.NewMessage(message.MTest, "Backing services for %s=%s stopped.", factor.Aspect().String(), factor.String()))
		}
	}
}

func modifyConfiguration(t *testing.T, factors []TestingFactor, cfg *config.AppConfig) bool {
	for _, factor := range factors {
		err := factor.ModifyConfiguration(cfg)
		if err != nil {
			t.Errorf("failed to apply factor %s (%v)", factor.String(), err)
			t.Fail()
			return true
		}
	}
	return false
}

func runTestSuite(
	t *testing.T,
	factors []TestingFactor,
	cfg config.AppConfig,
) {
	hardError := false
	testSuite := godog.TestSuite{
		Name:                t.Name(),
		ScenarioInitializer: scenarioInitializer(t, factors, cfg, &hardError),
		Options:             &opts,
	}
	if testSuite.Run() != 0 {
		if hardError {
			t.Fail()
		}
	}
}
func scenarioInitializer(
	t *testing.T,
	factors []TestingFactor,
	cfg config.AppConfig,
	hardError *bool,
) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		scenarioWG := map[string]*sync.WaitGroup{}
		testings := map[string]*testing.T{}
		mu := &sync.Mutex{}
		beforeScenario := createBeforeContext(mu, scenarioWG, t, testings)
		afterScenario := createAfterScenario(mu, testings, hardError, scenarioWG)
		ctx.BeforeScenario(beforeScenario)
		ctx.AfterScenario(afterScenario)

		for _, factor := range factors {
			logger := log.NewTestLogger(t)
			for _, step := range factor.GetSteps(cfg, logger) {
				ctx.Step(step.Match, step.Method)
			}
		}
	}
}

func createAfterScenario(
	mu *sync.Mutex,
	testings map[string]*testing.T,
	hardError *bool,
	scenarioWG map[string]*sync.WaitGroup,
) func(sc *godog.Scenario, err error) {
	afterScenario := func(sc *godog.Scenario, err error) {
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			if errors.Is(err, ErrSkipped) {
				if _, ok := testings[sc.Name]; ok {
					testings[sc.Name].Skipf("%v", err)
				}
			} else {
				*hardError = true
				if _, ok := testings[sc.Name]; ok {
					testings[sc.Name].Errorf("%v", err)
					testings[sc.Name].Fail()
				}
			}
		}
		if _, ok := scenarioWG[sc.Name]; ok {
			scenarioWG[sc.Name].Done()
		}
		delete(testings, sc.Name)
		delete(scenarioWG, sc.Name)
	}
	return afterScenario
}

func createBeforeContext(
	mu *sync.Mutex,
	scenarioWG map[string]*sync.WaitGroup,
	t *testing.T,
	testings map[string]*testing.T,
) func(sc *godog.Scenario) {
	beforeScenario := func(sc *godog.Scenario) {
		mu.Lock()
		defer mu.Unlock()
		if _, ok := scenarioWG[sc.Name]; !ok {
			scenarioWG[sc.Name] = &sync.WaitGroup{}
			scenarioWG[sc.Name].Add(1)
		}

		go t.Run(
			sc.Name,
			func(t *testing.T) {
				testings[sc.Name] = t
				mu.Lock()
				var ok bool
				var wg *sync.WaitGroup
				if wg, ok = scenarioWG[sc.Name]; ok {
					mu.Unlock()
					wg.Wait()
				} else {
					mu.Unlock()
				}
			},
		)
	}
	return beforeScenario
}
