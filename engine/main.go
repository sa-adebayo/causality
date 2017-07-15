package causality

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/codeskyblue/kexec"
	"github.com/gobuild/log"
)

func init() {
	for key, val := range signalMaps {
		signalMaps["SIG"+key] = val
		signalMaps[fmt.Sprintf("%d", val)] = val
	}
	log.SetFlags(0)
	if runtime.GOOS == "windows" {
		log.SetPrefix("causality >>> ")
	} else {
		log.SetPrefix("\033[32mcausality\033[0m >>> ")
	}
}

// Start is the public function that starts causality's event monitoring
func (event *TriggerEvent) Start() (waitC chan error) {
	ConsolePrintf(CGREEN, fmt.Sprintf("[%s] exec start: %v", event.Name, event.cmdArgs))
	startTime := time.Now()
	cmd := kexec.Command(event.cmdArgs[0], event.cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := os.Environ()
	for key, val := range event.Environ {
		env = append(env, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = env
	event.kcmd = cmd
	waitC = make(chan error, 1)
	if err := cmd.Start(); err != nil {
		waitC <- err
		return
	}
	go func() {
		err := cmd.Wait()
		waitC <- err
		log.Infof("[%s] finish in %s", event.Name, time.Since(startTime))
	}()
	return waitC
}

// Stop is the public function that stops causality's event monitoring
func (event *TriggerEvent) Stop(waitC chan error) bool {
	if event.kcmd != nil {
		if event.kcmd.ProcessState != nil && event.kcmd.ProcessState.Exited() {
			event.kcmd = nil
			return true
		}
		event.kcmd.Terminate(event.killSignal)
		var done bool
		select {
		case err := <-waitC:
			if err != nil {
				ConsolePrintf(CRED, "[%s] program exited: %v", event.Name, err)
			}
			done = true
		case <-time.After(event.stopTimeoutDuration):
			done = false
		}
		if !done {
			ConsolePrintf(CYELLOW, "[%s] program still alive", event.Name)
			event.kcmd.Terminate(syscall.SIGKILL)
		} else {
			event.kcmd = nil
		}
		return done
	}
	return true
}
