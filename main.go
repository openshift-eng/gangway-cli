package main

import (
	"fmt"
	"os"

	"github.com/openshift-eng/gangway-cli/cmd"
)

func main() {
	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
