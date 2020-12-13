package containerssh_test

import (
	"flag"
	"os"
	"testing"

	"github.com/containerssh/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/containerssh/containerssh/test/steps"
)

var opts = godog.Options{Output: colors.Colored(os.Stdout)}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = flag.Args()

	status := godog.TestSuite{
		Name:                 "godogs",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	loggerFactory := log.NewFactory(os.Stdout)
	logger, err := loggerFactory.Make(
		log.Config{
			Level:  log.LevelDebug,
			Format: log.FormatText,
		},
		"test",
	)
	if err != nil {
		panic(err)
	}
	scenario := &steps.Scenario{
		LoggerFactory: loggerFactory,
		Logger:        logger,
	}

	ctx.AfterScenario(
		func(*godog.Scenario, error) {
			_ = scenario.StopAuthServer()
			_ = scenario.StopConfigServer()
			_ = scenario.StopSshServer()
		},
	)

	ctx.Step(`^I start(?:|ed) the SSH server$`, scenario.StartSSHServer)
	ctx.Step(`^I stop(?:|ed) the SSH server$`, scenario.StopSshServer)

	ctx.Step(`^I start(?:|ed) the authentication server$`, scenario.StartAuthServer)
	ctx.Step(`^I stop(?:|ed) the authentication server$`, scenario.StopAuthServer)
	ctx.Step(`^I create(?:|d) the user "(.*)" with the password "(.*)"$`, scenario.CreateUser)

	ctx.Step(`^I start(?:|ed) the configuration server$`, scenario.StartConfigServer)
	ctx.Step(`^I stop(?:|ed) the configuration server$`, scenario.StopConfigServer)

	ctx.Step(
		`^authentication with user "(.*)" and password "(.*)" (?:should fail|should have failed)$`,
		scenario.AuthenticationShouldFail,
	)
	ctx.Step(
		`^authentication with user "(.*)" and password "(.*)" (?:should succeed|should have succeeded)$`,
		scenario.AuthenticationShouldSucceed,
	)

	ctx.Step(
		`^I should (?:be able to|should have been able to) execute a command with user "(.*)" and password "(.*)"$`,
		scenario.RunCommand,
	)

	ctx.Step(`^I configure the user "(.*)" to use Kubernetes`, scenario.ConfigureKubernetes)
	ctx.Step(`^I configure the user "(.*)" to use Docker`, scenario.ConfigureDocker)
}
