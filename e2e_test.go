package containerssh_test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/containerssh/configuration"
	"github.com/containerssh/log"
	"github.com/containerssh/structutils"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Strict: true,
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
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
		config := configuration.AppConfig{}
		structutils.Defaults(&config)

		if modifyConfiguration(t, factors, &config) {
			return
		}

		var startedFactors = &[]TestingFactor{}
		loggerFactory := log.NewLoggerPipelineFactory(os.Stdout)
		defer stopFactors(t, startedFactors, loggerFactory, config)()
		for _, factor := range factors {
			logger, err := loggerFactory.Make(
				log.Config{
					Level:  log.LevelDebug,
					Format: log.FormatLJSON,
				},
				factor.String(),
			)
			if err != nil {
				t.Errorf("failed to create logger for factor %s (%v)", factor.String(), err)
				t.Fail()
				return
			}
			logger.Noticef("Starting backing services for %s=%s...", factor.Aspect().String(), factor.String())
			if err = factor.StartBackingServices(config, logger, loggerFactory); err != nil {
				t.Errorf("failed to start backing services for %s=%s (%v)", factor.Aspect().String(), factor.String(), err)
				t.Fail()
				return
			}
			logger.Noticef("Backing services for %s=%s running.", factor.Aspect().String(), factor.String())
			*startedFactors = append(*startedFactors, factor)
		}

		runTestSuite(t, factors, loggerFactory, config)
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
	loggerFactory log.LoggerFactory,
	config configuration.AppConfig,
) func() {
	return func() {
		if startedFactors == nil {
			return
		}
		for _, factor := range *startedFactors {
			logger, err := loggerFactory.Make(
				log.Config{
					Level:  log.LevelDebug,
					Format: log.FormatLJSON,
				},
				factor.String(),
			)
			if err != nil {
				panic(err)
			}

			logger.Noticef("Stopping backing services for %s=%s...", factor.Aspect().String(), factor.String())
			err = factor.StopBackingServices(config, nil, nil)
			if err != nil {
				t.Errorf("failed to stop backing services for %s=%s (%v)", factor.Aspect().String(), factor.String(), err)
				t.Fail()
				return
			}
			logger.Noticef("Backing services for %s=%s stopped.", factor.Aspect().String(), factor.String())
		}
	}
}

func modifyConfiguration(t *testing.T, factors []TestingFactor, config *configuration.AppConfig) bool {
	for _, factor := range factors {
		err := factor.ModifyConfiguration(config)
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
	loggerFactory log.LoggerFactory,
	config configuration.AppConfig,
) {
	hardError := false
	testSuite := godog.TestSuite{
		Name: t.Name(),
		ScenarioInitializer: scenarioInitializer(t, factors, loggerFactory, config, &hardError),
		Options: &opts,
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
	loggerFactory log.LoggerFactory,
	config configuration.AppConfig,
	hardError *bool,
) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		scenarioWG := map[string]*sync.WaitGroup{}
		testings := map[string]*testing.T{}
		ctx.BeforeScenario(
			func(sc *godog.Scenario) {
				scenarioWG[sc.Name] = &sync.WaitGroup{}
				scenarioWG[sc.Name].Add(1)

				go t.Run(
					sc.Name,
					func(t *testing.T) {
						testings[sc.Name] = t
						scenarioWG[sc.Name].Wait()
					},
				)
			},
		)
		ctx.AfterScenario(
			func(sc *godog.Scenario, err error) {
				if err != nil {
					if errors.Is(err, ErrSkipped) {
						testings[sc.Name].Skipf("%v", err)
					} else {
						*hardError = true
						testings[sc.Name].Errorf("%v", err)
						testings[sc.Name].Fail()
					}
				}
				scenarioWG[sc.Name].Done()
				delete(testings, sc.Name)
				delete(scenarioWG, sc.Name)
			},
		)

		for _, factor := range factors {
			logger, err := loggerFactory.Make(
				log.Config{
					Level:  log.LevelDebug,
					Format: log.FormatText,
				},
				factor.String(),
			)
			if err != nil {
				panic(err)
			}
			for _, step := range factor.GetSteps(config, logger, loggerFactory) {
				ctx.Step(step.Match, step.Method)
			}
		}
	}
}
