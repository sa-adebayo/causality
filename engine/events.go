package causality

import (
	"strings"
	"sync"
	"time"

	ignore "github.com/codeskyblue/dockerignore"
	"github.com/go-fsnotify/fsnotify"
	"github.com/gobuild/log"
)

// WatchEvent is the causality watcher event
// when use func (event *TriggerEvent) strange things happened, wired
func (event *TriggerEvent) WatchEvent(eventChannel chan Event, wg *sync.WaitGroup) {
	waitC := event.Start()
	for evt := range eventChannel {
		isMatch, err := ignore.Matches(evt.Name, event.Patterns)
		if err != nil {
			log.Fatal(err)
		}
		if !isMatch {
			continue
		}
		if event.Stop(waitC) {
			ConsolePrintf(CGREEN, "changed: %v", evt.Name)
			ConsolePrintf(CGREEN, "delay: %v", event.Delay)
			time.Sleep(event.delayDuration)
			waitC = event.Start()
		}
	}
	event.killSignal = event.exitSignal
	event.Stop(waitC)
	wg.Done()
}

// WatchPathAndChildren watches path and its children (sub-directories)
// visits here for in case of duplicate paths
func WatchPathAndChildren(w *fsnotify.Watcher, paths []string, depth int, visits map[string]bool) error {
	if visits == nil {
		visits = make(map[string]bool)
	}
	watchDir := func(dir string) error {
		if visits[dir] {
			return nil
		}
		if err := w.Add(dir); err != nil {
			if strings.Contains(err.Error(), "too many open files") {
				log.Fatalf("Watch directory(%s) error: %v", dir, err)
			}
			log.Warnf("Watch directory(%s) error: %v", dir, err)
			return err
		}
		log.Debug("Watch directory:", dir)
		visits[dir] = true
		return nil
	}
	var err error
	for _, path := range paths {
		if visits[path] {
			continue
		}
		watchDir(path)
		dirs, er := ListAllDirectories(path, depth)
		if er != nil {
			err = er
			log.Warnf("ERR list dir: %s, depth: %d, %v", path, depth, err)
			continue
		}
		for _, dir := range dirs {
			watchDir(dir)
		}
	}
	return err
}

// DrainEvent drains the events
func DrainEvent(fwc Config) (globalEventC chan Event, wg *sync.WaitGroup, err error) {
	globalEventC = make(chan Event, 1)
	wg = &sync.WaitGroup{}
	evtChannls := make([]chan Event, 0)
	// log.Println(len(fwc.Triggers))
	for _, tg := range fwc.Triggers {
		wg.Add(1)
		evtC := make(chan Event, 1)
		evtChannls = append(evtChannls, evtC)
		go func(tge TriggerEvent) {
			tge.WatchEvent(evtC, wg)
		}(tg)

		// Can't write like this, the next loop tg changed, but go .. is not finished
		// go tg.WatchEvent(evtC, wg)
	}

	go func() {
		for evt := range globalEventC {
			for _, eC := range evtChannls {
				eC <- evt
			}
		}
		for _, eC := range evtChannls {
			close(eC)
		}
	}()
	return
}

// TransformEvent transform the event channel
func TransformEvent(fsw *fsnotify.Watcher, evtC chan Event) {
	go func() {
		for err := range fsw.Errors {
			log.Errorf("Watch error: %v", err)
		}
	}()
	for evt := range fsw.Events {
		if evt.Op == fsnotify.Create && IsDirectory(evt.Name) {
			log.Info("Add watcher", evt.Name)
			fsw.Add(evt.Name)
			continue
		}
		if evt.Op == fsnotify.Remove {
			if err := fsw.Remove(evt.Name); err == nil {
				log.Info("Remove watcher", evt.Name)
			}
			continue
		}
		if !IsChanged(evt.Name) {
			continue
		}
		//log.Printf("Changed: %s", evt.Name)
		evtC <- Event{ // may panic here
			Name: evt.Name,
		}
	}
}
