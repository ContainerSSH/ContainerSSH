package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

    "go.containerssh.io/libcontainerssh/config"
)

var testLoggerActive = false

// RunTests is a helper function that transforms the test output depending on the CI system running.
// This is most prominently used with GitHub Actions to produce a nice-looking output.
func RunTests(m *testing.M) {
	var result int
	switch {
	case os.Getenv("GITHUB_ACTIONS") != "":
		result = runWithGitHubActions(m)
	default:
		result = m.Run()
	}
	os.Exit(result)
}

func runWithGitHubActions(m *testing.M) int {
	w := &gitHubActionsWriter{
		os.Stdout,
		&sync.Mutex{},
		map[string]*testCase{},
		"",
		"",
	}
	oldStdout := os.Stdout
	tmpFile, err := os.CreateTemp(os.TempDir(), "test-")
	if err != nil {
		panic(err)
	}
	reader, err := os.Open(tmpFile.Name())
	if err != nil {
		panic(err)
	}
	var result int
	done := make(chan struct{})
	go func() {
		os.Stdout = tmpFile
		testLoggerActive = true
		result = m.Run()
		testLoggerActive = false
		os.Stdout = oldStdout
		done <- struct{}{}
	}()
	<-done
	_, _ = io.Copy(w, reader)
	_ = reader.Close()
	_ = tmpFile.Close()
	_ = os.Remove(tmpFile.Name())

	var testCases []string
	for testCaseName := range w.testCases {
		testCases = append(testCases, testCaseName)
	}
	sort.Strings(testCases)
	for _, testCaseName := range testCases {
		writeTestcase(w.testCases[testCaseName])
	}

	return result
}

type logOutputFormat struct {
	symbol      string
	color       string
	symbolColor string
}

var logLevelConfig = map[config.LogLevelString]logOutputFormat{
	config.LevelDebugString: {
		symbol:      "⚙️",
		color:       "",
		symbolColor: "",
	},
	config.LevelInfoString: {
		symbol:      "ⓘ️",
		color:       "\033[34m",
		symbolColor: "\033[34m",
	},
	config.LevelNoticeString: {
		symbol:      "🏷️",
		color:       "\033[33m",
		symbolColor: "\033[33m",
	},
	config.LevelWarningString: {
		symbol:      "⚠️",
		color:       "\033[33m",
		symbolColor: "\033[33m",
	},
	config.LevelErrorString: {
		symbol:      "❌",
		color:       "\033[31m",
		symbolColor: "\033[31m",
	},
	config.LevelCriticalString: {
		symbol:      "🛑",
		color:       "\033[31m",
		symbolColor: "\033[31m",
	},
	config.LevelAlertString: {
		symbol:      "🔔",
		color:       "\033[31m",
		symbolColor: "\033[31m",
	},
	config.LevelEmergencyString: {
		symbol:      "💣",
		color:       "\033[31m",
		symbolColor: "\033[31m",
	},
}

func writeTestcase(c *testCase) {
	fmt.Printf("::group::")
	if c.pass {
		fmt.Printf("\033[0;32m✅ %s\033[0m (%s)\n", c.name, c.time)
	} else {
		fmt.Printf("\033[0;31m❌ %s\033[0m (%s)\n", c.name, c.time)
	}
	for _, line := range c.lines {
		format := logLevelConfig[line.level]
		fmt.Printf(
			"%s%s\033[0m %s%s \033[0;37m(%s:%d)\033[0m\n",
			format.symbolColor,
			format.symbol,
			format.color,
			line.message,
			path.Base(line.file),
			line.line,
		)
	}
	fmt.Printf("::endgroup::\n")
}

type gitHubActionsWriter struct {
	backend      io.Writer
	lock         *sync.Mutex
	testCases    map[string]*testCase
	lastTestCase string
	lastLine     string
}

func (g *gitHubActionsWriter) Write(p []byte) (n int, err error) {
	g.lock.Lock()
	defer g.lock.Unlock()
	lines := strings.Split(fmt.Sprintf("%s%s", g.lastLine, string(p)), "\n")
	for _, line := range lines[:len(lines)-1] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		switch {
		case strings.HasPrefix(strings.TrimSpace(line), "=== RUN "):
			g.lastTestCase = g.processRun(line)
		case strings.HasPrefix(strings.TrimSpace(line), "=== CONT "):
			g.lastTestCase = g.processCont(line)
		case strings.HasPrefix(strings.TrimSpace(line), "=== PAUSE "):
			g.lastTestCase = g.processPause(line)
		case strings.HasPrefix(strings.TrimSpace(line), "--- PASS:"):
			g.lastTestCase = g.processPass(line)
		case strings.HasPrefix(strings.TrimSpace(line), "--- FAIL:"):
			g.lastTestCase = g.processFail(line)
		case line == "PASS":
		case line == "FAIL":
		case line == "":
		default:
			g.processDefault(line)
		}
	}
	g.lastLine = lines[len(lines)-1]
	return len(p), nil
}

func (g *gitHubActionsWriter) processCont(line string) string {
	return strings.TrimSpace(strings.Replace(line, "=== CONT ", "", 1))
}

func (g *gitHubActionsWriter) processPause(line string) string {
	return strings.TrimSpace(strings.Replace(line, "=== PAUSE ", "", 1))
}

func (g *gitHubActionsWriter) processPass(line string) string {
	parts := strings.SplitN(strings.TrimSpace(line), " ", 4)
	lastTestCase := parts[2]
	g.testCases[lastTestCase].pass = true
	lastTestCase = ""
	return lastTestCase
}

func (g *gitHubActionsWriter) processFail(line string) string {
	parts := strings.SplitN(strings.TrimSpace(line), " ", 4)
	lastTestCase := parts[2]
	t, err := time.ParseDuration(strings.Trim(parts[3], "()"))
	if err != nil {
		panic(err)
	}
	g.testCases[lastTestCase].time = t
	lastTestCase = ""
	return lastTestCase
}

func (g *gitHubActionsWriter) processDefault(line string) {
	parts := strings.SplitN(strings.TrimSpace(line), "\t", 6)
	if len(parts) == 6 {
		lineNumber, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			panic(err)
		}
		if g.lastTestCase != "" {
			g.testCases[g.lastTestCase].lines = append(
				g.testCases[g.lastTestCase].lines,
				testCaseLine{
					file:    parts[1],
					line:    lineNumber,
					level:   config.LogLevelString(parts[3]),
					code:    parts[4],
					message: strings.TrimSpace(parts[5]),
				},
			)
		} else {
			panic(fmt.Errorf("no test case for %s, line: %s", g.lastTestCase, line))
		}
	} else {
		if g.lastTestCase != "" {
			if _, ok := g.testCases[g.lastTestCase]; !ok {
				panic(fmt.Errorf("no test case for %s, line: %s", g.lastTestCase, line))
			}
			g.testCases[g.lastTestCase].lines = append(
				g.testCases[g.lastTestCase].lines,
				testCaseLine{
					file:    "",
					line:    0,
					level:   config.LevelDebugString,
					code:    "",
					message: strings.TrimSpace(line),
				},
			)
		}
	}
}

func (g *gitHubActionsWriter) processRun(line string) string {
	lastTestCase := strings.TrimSpace(strings.Replace(line, "=== RUN ", "", 1))
	if _, ok := g.testCases[lastTestCase]; !ok {
		g.testCases[lastTestCase] = &testCase{
			name: lastTestCase,
		}
	}
	return lastTestCase
}

type testCase struct {
	name  string
	pass  bool
	time  time.Duration
	lines []testCaseLine
}

type testCaseLine struct {
	file    string
	line    uint64
	level   config.LogLevelString
	code    string
	message string
}
