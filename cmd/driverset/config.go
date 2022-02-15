// Package driverset provides Cobra commands for various bits of driverset functionality.
package driverset

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Setting defines where a driver setting's file is and the names and values of its possible
// settings.
type Setting struct {
	File    string
	Actions map[string]string
}

// Read returns the contents of the settings file, along with the corresponding action name. If
// there is none, the returned action is the empty string.
func (s Setting) Read() (contents, action string, err error) {
	b, err := os.ReadFile(s.File)

	contents = string(b)

	for action, text := range s.Actions {
		if strings.TrimSpace(contents) == text {
			return contents, action, err
		}
	}

	return contents, "", err
}

// ErrNotExist is returned by Set or Disambiguate if the specified action is not in the config.
var ErrNotExist = errors.New("undefined setting/action")

// Set sets the settings file contents to the text corresponding to the specified action.
func (s Setting) Set(action string) error {
	text, ok := s.Actions[action]
	if !ok {
		return ErrNotExist
	}

	// It's unlikely we're going to create the file, but if we do we can use -rw-r--r-- permissions
	// because that's what these sorts of files look like.
	return os.WriteFile(s.File, []byte(text), 0o644) //nolint:gosec // See above.
}

// Config contains the names and settings that driverset can operate on.
type Config map[string]Setting

// ErrValidation is returned by (c Config) Validate when something is wrong.
var ErrValidation = errors.New("config file failed validation")

// Validate writes any validation errors to w and returns true if there were any.
//
// First, we'll make sure no actions for the same setting have the same text.
//
// Validation also means checking whether any settings names and actions will be ambiguous when
// when combined on the command line. Because we want to allow the user to say "conservation mode
// off", it is an error to have a setting "conservation mode" with action "off" and a setting
// "conservation" with action "mode off".
//
// I know it's unlikely but hey.
func (c Config) Validate(w io.Writer) (err error) {
	// Make sure no actions for the same setting have the same text.
	for setting, contents := range c {
		texts := make(map[string][]string, len(contents.Actions))

		for action, text := range contents.Actions {
			texts[text] = append(texts[text], action)
		}

		for _, actions := range texts {
			if len(actions) > 1 {
				err = ErrValidation

				fmt.Fprintf(w, "actions with same text for setting '%s':\n", setting)
				sort.Strings(actions)
				for _, action := range actions {
					fmt.Fprintf(w, "  - %s\n", action)
				}
			}
		}
	}

	// Make sure there are no ambiguous setting names and actions.
	type pair struct {
		name   string
		action string
	}
	strs := make(map[string][]pair, 2*len(c)) // likely at least two actions per setting

	for setting, contents := range c {
		for action := range contents.Actions {
			str := setting + " " + action

			strs[str] = append(strs[str], pair{setting, action})
		}
	}

	// Sorted order for determinism.
	keys := make([]string, 0, len(strs))
	for str := range strs {
		keys = append(keys, str)
	}
	sort.Strings(keys)
	for _, str := range keys {
		pairs := strs[str]

		if len(pairs) > 1 {
			err = ErrValidation

			fmt.Fprintf(w, "ambiguous settings: '%s' could be any of:\n", str)
			sort.Slice(pairs, func(i, j int) bool {
				if pairs[i].name == pairs[j].name {
					return pairs[i].action < pairs[j].action
				}

				return pairs[i].name < pairs[j].name
			})
			for _, pair := range pairs {
				fmt.Fprintf(w, "  - '%s' + '%s'\n", pair.name, pair.action)
			}
		}
	}

	return err
}

// Disambiguate the args into a setting and action, e.g.:
//
// 	"conservation mode off" -> "conservation mode" "off"
//
// Validation guarantees there's a unique answer.
//
// len(args) must be at least 2.
func (c Config) Disambiguate(args []string) (setting, action string, err error) {
	for i := 1; i < len(args); i++ {
		setting = strings.Join(args[:i], " ")
		action = strings.Join(args[i:], " ")

		s, ok := c[setting]
		if !ok {
			continue
		}
		_, ok = s.Actions[action]
		if ok {
			return setting, action, nil
		}
	}

	return "", "", ErrNotExist
}
