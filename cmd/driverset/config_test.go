package driverset_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/kylrth/driverset/cmd/driverset"
)

var defaultActions = map[string]string{
	"on":  "1",
	"off": "0",
}

func makeDefaultSetting(file string) driverset.Setting {
	return driverset.Setting{
		File:    file,
		Actions: defaultActions,
	}
}

const nilErr = "<nil>"

func TestSetting_Read(t *testing.T) {
	t.Parallel()

	// Write some sample files.
	dir := t.TempDir()
	onFile := filepath.Join(dir, "conservation_mode")
	offFile := filepath.Join(dir, "other_mode")
	nopeFile := filepath.Join(dir, "nope_mode")
	missingFile := filepath.Join(dir, "missing")
	err := os.WriteFile(onFile, []byte("1\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(offFile, []byte("0\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(nopeFile, []byte("sandwich\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		s driverset.Setting

		contents string
		action   string
		err      string
	}{
		"on": {
			s:        makeDefaultSetting(onFile),
			contents: "1\n",
			action:   "on",
			err:      nilErr,
		},
		"off": {
			s:        makeDefaultSetting(offFile),
			contents: "0\n",
			action:   "off",
			err:      nilErr,
		},
		"missing": {
			s:        makeDefaultSetting(missingFile),
			contents: "",
			action:   "",
			err:      fmt.Sprintf("open %s: no such file or directory", missingFile),
		},
		"unnamed": {
			s:        makeDefaultSetting(nopeFile),
			contents: "sandwich\n",
			action:   "",
			err:      nilErr,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			contents, action, err := tc.s.Read()

			errStr := nilErr
			if err != nil {
				errStr = err.Error()
			}
			diff := cmp.Diff(tc.err, errStr)
			if diff != "" {
				t.Error("unexpected error (-want +got):\n" + diff)
			}

			diff = cmp.Diff(tc.contents, contents)
			if diff != "" {
				t.Error("unexpected contents (-want +got):\n" + diff)
			}

			diff = cmp.Diff(tc.action, action)
			if diff != "" {
				t.Error("unexpected action (-want +got):\n" + diff)
			}
		})
	}
}

func TestSetting_Set(t *testing.T) {
	t.Parallel()

	// Written files go here.
	dir := t.TempDir()
	// We'll make sure we can overwrite.
	err := os.WriteFile(filepath.Join(dir, "present"), nil, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		action string

		err          string
		fileContents string
	}{
		"present": {
			action:       "on",
			err:          nilErr,
			fileContents: "1",
		},
		"notpresent": {
			action:       "off",
			err:          nilErr,
			fileContents: "0",
		},
		"unnamed": {
			action: "sandwich",
			err:    "undefined setting/action",
		},
	}

	for name, tc := range tests {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file := filepath.Join(dir, name)
			s := makeDefaultSetting(file)
			err := s.Set(tc.action)

			errStr := nilErr
			if err != nil {
				errStr = err.Error()
			}
			diff := cmp.Diff(tc.err, errStr)
			if diff != "" {
				t.Error("unexpected error (-want +got):\n" + diff)
			}
			if err != nil && diff == "" {
				// We got the error we wanted.
				return
			}

			fileContents, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("file '%s' was not created", file)
			}
			diff = cmp.Diff(tc.fileContents, string(fileContents))
			if diff != "" {
				t.Error("unexpected file contents (-want +got):\n" + diff)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		c    driverset.Config
		want string
	}{
		"ok": {
			c: driverset.Config{
				"conservation mode": driverset.Setting{
					File: "path/to/conservation_mode",
					Actions: map[string]string{
						"on":  "1",
						"off": "0",
					},
				},
				"conservation NOT": driverset.Setting{
					File: "other/setting_path",
					Actions: map[string]string{
						"on":  "1",
						"off": "0",
					},
				},
			},
			want: "",
		},
		"duplicate": {
			c: driverset.Config{
				"conservation mode": driverset.Setting{
					File: "path/to/conservation_mode",
					Actions: map[string]string{
						"on":    "1",
						"off":   "0",
						"other": "1",
					},
				},
			},
			want: `actions with same text for setting 'conservation mode':
  - on
  - other
`,
		},
		"ambiguous": {
			c: driverset.Config{
				"conservation mode": driverset.Setting{
					File: "path/to/conservation_mode",
					Actions: map[string]string{
						"on":  "1",
						"off": "0",
					},
				},
				"conservation": driverset.Setting{
					File: "other/setting_path",
					Actions: map[string]string{
						"mode on":  "1",
						"mode off": "0",
					},
				},
			},
			want: `ambiguous settings: 'conservation mode off' could be any of:
  - 'conservation' + 'mode off'
  - 'conservation mode' + 'off'
ambiguous settings: 'conservation mode on' could be any of:
  - 'conservation' + 'mode on'
  - 'conservation mode' + 'on'
`,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got bytes.Buffer
			err := tc.c.Validate(&got)
			if tc.want == "" && err != nil {
				t.Errorf("got err '%v'; expected '<nil>'", err)
			} else if tc.want != "" && err == nil {
				t.Errorf("got err '<nil>'; expected '%v'", err)
			}

			diff := cmp.Diff(tc.want, got.String())
			if diff != "" {
				t.Error("unexpected output (-want +got):\n" + diff)
			}
		})
	}
}

func TestConfig_Disambiguate(t *testing.T) {
	t.Parallel()

	c := driverset.Config{
		"conservation mode": {
			Actions: map[string]string{
				"on":  "1",
				"off": "0",
			},
		},
		"conservation": {
			Actions: map[string]string{
				"turned on":  "12",
				"turned off": "13",
			},
		},
		"conservation mode on": {
			Actions: map[string]string{
				"on":  "1",
				"off": "0",
			},
		},
	}

	// Test every combination plus one that doesn't exist.
	type test struct {
		args []string

		setting string
		action  string
		err     string
	}
	tests := make(map[string]test, 1+2*len(c))
	tests["garbage"] = test{
		args: []string{"spoons", "conservation", "mode"},
		err:  "undefined setting/action",
	}
	for setting, s := range c {
		for action := range s.Actions {
			str := setting + " " + action
			if _, ok := tests[str]; ok {
				t.Fatal("config would not pass validation")
			}

			args := strings.Split(setting, " ")
			args = append(args, strings.Split(action, " ")...)

			tests[str] = test{
				args:    args,
				setting: setting,
				action:  action,
				err:     nilErr,
			}
		}
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			setting, action, err := c.Disambiguate(tc.args)

			errStr := nilErr
			if err != nil {
				errStr = err.Error()
			}
			diff := cmp.Diff(tc.err, errStr)
			if diff != "" {
				t.Error("unexpected error (-want +got):\n" + diff)
			}

			diff = cmp.Diff(tc.setting, setting)
			if diff != "" {
				t.Error("unexpected setting (-want +got):\n" + diff)
			}

			diff = cmp.Diff(tc.action, action)
			if diff != "" {
				t.Error("unexpected action (-want +got):\n" + diff)
			}
		})
	}
}
