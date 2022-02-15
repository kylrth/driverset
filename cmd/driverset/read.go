package driverset

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// ReadCmd prints a driver setting.
var ReadCmd = &cobra.Command{
	Use:   "read SETTINGNAME",
	Short: "print a driver setting",
	Long: `Read a driver setting file and print the action that matches its contents.

If the contents of the file don't match any of the actions in the config, the
contents are printed instead.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := Read(strings.Join(args, " "), configFile); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Read(setting, configFile string) error {
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

	if _, ok := cfg[setting]; !ok {
		return ErrNotExist
	}

	contents, action, err := cfg[setting].Read()
	if err != nil {
		return fmt.Errorf("read setting: %w", err)
	}

	if action != "" {
		fmt.Printf("%s is currently '%s'\n", setting, action)

		return nil
	}

	// The file contents don't match any of our actions. We'll report that.
	fmt.Printf("%s is currently set with unknown contents:\n%s\n", setting, contents)

	return nil
}
