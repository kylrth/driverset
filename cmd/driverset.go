package main

import (
	"fmt"
	"os"

	"github.com/kylrth/driverset/cmd/driverset"
)

func main() {
	if err := driverset.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
