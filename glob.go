package main

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/papertrail/remote_syslog2/utils"
)

func glob(
	patterns []string,
	exclusions []*regexp.Regexp,
	wr *WorkerRegistry,
	warn bool,
) (map[string]bool, error) {
	m := map[string]bool{}
	for _, glob := range patterns {
		files, err := filepath.Glob(utils.ResolvePath(glob))
		if err != nil {
			log.Errorf("Failed to glob %s: %s", glob, err)
			continue
		}
		if files == nil && warn {
			log.Errorf("Cannot forward %s, it may not exist", glob)
			continue
		}
		for _, file := range files {
			if wr.Exists(file) {
				log.Debugf("Skipping %s because it is already tailed", file)
				continue
			}
			if match(file, exclusions) {
				log.Debugf("Skipping %s because it is excluded by regular expression", file)
				continue
			}
			// skip directories
			fi, err := os.Stat(file)
			if err != nil {
				log.Debugf("Cannot stat file '%s': %v", file, err)
				continue
			}
			if fi.IsDir() {
				log.Debugf("Skipping '%s', use '<dir>/*' to tail files", file)
				continue
			}
			m[file] = true
		}
	}
	return m, nil
}

// Evaluates each regex against the string. If any one is a match
// the function returns true, otherwise it returns false
func match(value string, expressions []*regexp.Regexp) bool {
	for _, exp := range expressions {
		if exp.MatchString(value) {
			return true
		}
	}
	return false
}
