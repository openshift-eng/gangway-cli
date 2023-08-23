package main

import (
	"fmt"
	"os"

	"github.com/openshift-eng/gangway-cli/cmd"
)

func main() {
	cmd := cmd.NewCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
