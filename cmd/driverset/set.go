package driverset

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// SetCmd sets a driver setting.
var SetCmd = &cobra.Command{
	Use:   "set SETTINGNAME ACTION",
	Short: "change a driver setting",
	Long: `Change a driver setting by writing the action text to the corresponding file.
`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := Set(args, configFile); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Set(args []string, configFile string) error {
	b, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	cfg := make(Config)
	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	err = cfg.Validate(os.Stderr)
	if err != nil {
		return err
	}

	setting, action, err := cfg.Disambiguate(args)
	if err != nil {
		return err
	}
	if _, ok := cfg[setting]; !ok {
		return ErrNotExist
	}

	err = cfg[setting].Set(action)
	if err != nil {
		return fmt.Errorf("update setting: %w", err)
	}

	fmt.Printf("%s set to '%s'\n", setting, action)

	return nil
}
