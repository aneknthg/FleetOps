package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/aneknthg/fleetops/tools/fleetctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, cmd.ErrHighFindings) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}
