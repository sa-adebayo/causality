package causality

import (
	"os"
	"syscall"
	"time"

	"github.com/codeskyblue/kexec"
)

// CausalityConfigYAML is the YAML configuration file
// CausalityConfigJSON is the JSON configuration file
const (
	ConfigYAML = ".causality.yml"
	ConfigJSON = ".causality.json"
)

// Version is the current version number of the application
var (
	Version = "1.0"
)

var fileModifyTimeMap = make(map[string]time.Time)

var signalMaps = map[string]os.Signal{
	"INT":  syscall.SIGINT,
	"HUP":  syscall.SIGHUP,
	"QUIT": syscall.SIGQUIT,
	"TRAP": syscall.SIGTRAP,
	"TERM": syscall.SIGTERM,
	"KILL": syscall.SIGKILL,
}

// Color codes
const (
	CBLACK   = "30"
	CRED     = "31"
	CGREEN   = "32"
	CYELLOW  = "33"
	CBLUE    = "34"
	CMAGENTA = "35"
	CPURPLE  = "36"
)

// Event is the event struct
type Event struct {
	Name string
}

// Config is the configuration struct
type Config struct {
	Description string         `yaml:"desc" json:"desc"`
	Triggers    []TriggerEvent `yaml:"triggers" json:"triggers"`
	WatchPaths  []string       `yaml:"watch_paths" json:"watch_paths"`
	WatchDepth  int            `yaml:"watch_depth" json:"watch_depth"`
}

// TriggerEvent is the event trigger struct
type TriggerEvent struct {
	Name                string            `yaml:"name" json:"name"`
	Pattens             []string          `yaml:"pattens" json:"pattens"`
	matchPattens        []string          ``
	Environ             map[string]string `yaml:"env" json:"env"`
	Command             string            `yaml:"cmd" json:"cmd"`
	Shell               bool              `yaml:"shell" json:"shell"`
	cmdArgs             []string          ``
	Delay               string            `yaml:"delay" json:"delay"`
	delayDuration       time.Duration     ``
	StopTimeout         string            `yaml:"stop_timeout" json:"stop_timeout"`
	stopTimeoutDuration time.Duration     ``
	Signal              string            `yaml:"signal" json:"signal"`
	killSignal          os.Signal         ``
	KillSignal          string            `yaml:"kill_signal" json:"kill_signal"`
	exitSignal          os.Signal
	kcmd                *kexec.KCommand
}
