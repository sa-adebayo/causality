package causality

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"

	"regexp"

	ignore "github.com/codeskyblue/dockerignore"
	"github.com/google/shlex"
)

// InitializeConfig initializes configuration
func InitializeConfig() {
	fwc := generateConfig()
	format := readString("Save format .fsw.(json|yml)", "yml")
	var data []byte
	var cfg string
	if strings.ToLower(format) == "json" {
		data, _ = json.MarshalIndent(fwc, "", "  ")
		cfg = ConfigJSON
		ioutil.WriteFile(ConfigJSON, data, 0644)
	} else {
		cfg = ConfigYAML
		data, _ = yaml.Marshal(fwc)
		ioutil.WriteFile(ConfigYAML, data, 0644)
	}
	fmt.Printf("Saved to %s\n", strconv.Quote(cfg))
}

// generateConfig generates config returns the config generated
func generateConfig() Config {
	var (
		name    string
		command string
		pattern string
	)
	cwd, _ := os.Getwd()
	name = filepath.Base(cwd)
	name = readString("name:", name)

	for command == "" {
		fmt.Print("[?] command (yarn test): ")
		reader := bufio.NewReader(os.Stdin)
		command, _ = reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "" {
			command = "yarn test"
		}
	}

	for pattern == "" {
		fmt.Print(`[?] patterns (**/*.go, **/*.py, **/.casuality): `)
		reader := bufio.NewReader(os.Stdin)
		pattern, _ = reader.ReadString('\n')
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			pattern = `**/*.go, **/*.py, **/.casuality`
		}
	}

	patterns := regexp.MustCompile(`,\s*`).Split(pattern, -1)

	config := Config{
		Description: fmt.Sprintf("Auto generated by causality [%s]", name),
		Triggers: []TriggerEvent{{
			Name:     name,
			Patterns: patterns,
			Environ: map[string]string{
				"DEBUG": "1",
			},
			Shell:   true,
			Command: command,
		}},
	}
	out, _ := parseConfig(config)
	return out
}

// ReadConfig reads the configuration file from the slice of path passed to it.
func ReadConfig(paths ...string) (config Config, err error) {
	for _, configurationPath := range paths {
		data, err := ioutil.ReadFile(configurationPath)
		if err != nil {
			continue
		}
		ext := filepath.Ext(configurationPath)
		switch ext {
		case ".yml":
			if er := yaml.Unmarshal(data, &config); er != nil {
				return config, er
			}
		case ".json":
			if er := json.Unmarshal(data, &config); er != nil {
				return config, er
			}
		default:
			err = fmt.Errorf("Unknown format config file: %s", configurationPath)
			return config, err
		}
		return parseConfig(config)
	}
	return config, errors.New("Config file not exists")
}

// parseConfig parses current config and add more to it.
// returns both an extended config and error object
func parseConfig(in Config) (out Config, err error) {
	out = in
	for index, trigger := range in.Triggers {
		outTrigger := &out.Triggers[index]
		if trigger.Delay == "" {
			outTrigger.Delay = "100ms"
		}
		outTrigger.delayDuration, err = time.ParseDuration(outTrigger.Delay)
		if err != nil {
			return
		}
		if trigger.StopTimeout == "" {
			outTrigger.StopTimeout = "500ms"
		}
		outTrigger.stopTimeoutDuration, err = time.ParseDuration(outTrigger.StopTimeout)
		if err != nil {
			return
		}
		if outTrigger.Signal == "" {
			outTrigger.Signal = "KILL"
		}
		outTrigger.killSignal = signalMaps[outTrigger.Signal]
		if outTrigger.KillSignal == "" {
			outTrigger.exitSignal = syscall.SIGKILL
		} else {
			outTrigger.exitSignal = signalMaps[outTrigger.KillSignal]
		}
		readCloser := ioutil.NopCloser(bytes.NewBufferString(strings.Join(outTrigger.Patterns, "\n")))
		patterns, er := ignore.ReadIgnore(readCloser)
		if er != nil {
			err = er
			return
		}
		outTrigger.matchPatterns = patterns
		if outTrigger.Shell {
			sh, er := getShell()
			if er != nil {
				err = er
				return
			}
			outTrigger.cmdArgs = append(sh, outTrigger.Command)
		} else {
			outTrigger.cmdArgs, err = shlex.Split(outTrigger.Command)
			if err != nil {
				return
			}
			if len(outTrigger.cmdArgs) == 0 {
				err = errors.New("No command defined")
				return
			}
		}
	}
	if len(out.WatchPaths) == 0 {
		out.WatchPaths = append(out.WatchPaths, ".")
	}
	if out.WatchDepth < 0 {
		out.WatchDepth = 0
	}

	return
}
