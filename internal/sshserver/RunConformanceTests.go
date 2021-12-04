package sshserver

import (
	"testing"
)

// RunConformanceTests runs a suite of conformance tests against the provided backends supporting a standard
// Linux shell.
//goland:noinspection GoUnusedExportedFunction
func RunConformanceTests(t *testing.T, backendFactories map[string]ConformanceTestBackendFactory) {
	//t.Parallel()()

	for name, factory := range backendFactories {
		n := name
		f := factory
		t.Run(n, func(t *testing.T) {
			//t.Parallel()()
			testSuite := &conformanceTestSuite{
				backendFactory: f,
			}
			t.Run("singleProgramShouldRun", testSuite.singleProgramShouldRun)
			t.Run("settingEnvVariablesShouldWork", testSuite.settingEnvVariablesShouldWork)
			t.Run("runningInteractiveShellShouldWork", testSuite.runningInteractiveShellShouldWork)
			t.Run("reportingExitCodeShouldWork", testSuite.reportingExitCodeShouldWork)
			t.Run("sendingSignalsShouldWork", testSuite.sendingSignalsShouldWork)
		})
	}
}
