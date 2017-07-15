package causality

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// IsChanged checks if a directory of file has changed
func IsChanged(path string) bool {
	pinfo, err := os.Stat(path)
	if err != nil {
		return true
	}
	mtime := pinfo.ModTime()
	if mtime.Sub(fileModifyTimeMap[path]) > time.Millisecond*100 { // 100ms
		fileModifyTimeMap[path] = mtime
		return true
	}
	return false
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	pinfo, err := os.Stat(path)
	return err == nil && pinfo.IsDir()
}

// UniqueStrings returns unique mapping of imputed strings
func UniqueStrings(ss []string) []string {
	out := make([]string, 0, len(ss))
	m := make(map[string]bool, len(ss))
	for _, key := range ss {
		if !m[key] {
			out = append(out, key)
			m[key] = true
		}
	}
	return out
}

// ListAllDirectories list all directories in path
func ListAllDirectories(path string, depth int) (dirs []string, err error) {
	baseNumSeps := strings.Count(path, string(os.PathSeparator))
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			base := info.Name()
			if base != "." && strings.HasPrefix(base, ".") { // ignore hidden dir
				return filepath.SkipDir
			}
			if base == "node_modules" {
				return filepath.SkipDir
			}

			pathDepth := strings.Count(path, string(os.PathSeparator)) - baseNumSeps
			if pathDepth > depth {
				return filepath.SkipDir
			}
			dirs = append(dirs, path)
		}
		return nil
	})
	return
}

// ConsolePrintf prints to the terminal with specified format
func ConsolePrintf(ansiColor string, format string, args ...interface{}) {
	if runtime.GOOS != "windows" {
		format = "\033[" + ansiColor + "m" + format + "\033[0m"
	}
	log.Printf(format, args...)
}

func readString(prompt, value string) string {
	fmt.Printf("[?] %s (%s) ", prompt, value)
	var s = value
	fmt.Scanf("%s", &s)
	return s
}

func getShell() ([]string, error) {
	if path, err := exec.LookPath("bash"); err == nil {
		return []string{path, "-c"}, nil
	}
	if path, err := exec.LookPath("sh"); err == nil {
		return []string{path, "-c"}, nil
	}
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c"}, nil
	}
	return nil, fmt.Errorf("Could not find bash or sh on path")
}
