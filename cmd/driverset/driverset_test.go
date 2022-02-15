package driverset_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kylrth/driverset/cmd/driverset"
)

func Example() {
	dir, _ := os.MkdirTemp("", "")
	cfgFile := filepath.Join(dir, "driverset_test_config.yml")
	_ = os.WriteFile(cfgFile, []byte(fmt.Sprintf(`conservation mode:
  file: >-
    %s/driverset_test_conservation_mode
  actions:
    'on': '1'
    'off': '0'
`, dir)), 0o600)

	_ = driverset.Set([]string{"conservation", "mode", "on"}, cfgFile)
	_ = driverset.Read("conservation mode", cfgFile)
	_ = driverset.Set([]string{"conservation", "mode", "off"}, cfgFile)
	// Output:
	// conservation mode set to 'on'
	// conservation mode is currently 'on'
	// conservation mode set to 'off'
}
