package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-fsnotify/fsnotify"
	"github.com/gobuild/log"
	causality "github.com/sa-adebayo/causality/engine"
)

func main() {
	version := flag.Bool("version", false, "Show version")
	configfile := flag.String("config", causality.ConfigYAML, "Specify config file")
	flag.Parse()

	if *version {
		fmt.Println(causality.Version)
		return
	}

	subCmd := flag.Arg(0)
	var config causality.Config
	var err error
	if subCmd == "" {
		config, err = causality.ReadConfig(*configfile, causality.ConfigJSON)
		if err == nil {
			subCmd = "start"
		} else {
			subCmd = "init"
		}
	}

	switch subCmd {
	case "init":
		causality.InitializeConfig()
	case "start":
		visits := make(map[string]bool)
		fsw, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}

		err = causality.WatchPathAndChildren(fsw, config.WatchPaths, config.WatchDepth, visits)
		if err != nil {
			log.Println(err)
		}

		evtC, wg, err := causality.DrainEvent(config)
		if err != nil {
			log.Fatal(err)
		}

		sigOS := make(chan os.Signal, 1)
		signal.Notify(sigOS, syscall.SIGINT)
		signal.Notify(sigOS, syscall.SIGTERM)

		go func() {
			sig := <-sigOS
			causality.ConsolePrintf(causality.CPURPLE, "Catch signal %v!", sig)
			close(evtC)
		}()
		go causality.TransformEvent(fsw, evtC)
		wg.Wait()
		causality.ConsolePrintf(causality.CPURPLE, "Kill all running ... Done")
	}
}
