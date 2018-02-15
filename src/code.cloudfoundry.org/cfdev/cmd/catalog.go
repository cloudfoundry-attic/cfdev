package cmd

import (
	"fmt"
	"os"
	"encoding/json"
)

type Catalog struct{}

func(c *Catalog) Run(args []string) {
	bytes, err := json.MarshalIndent(catalog(), "", "  ")

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to marshal catalog: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(bytes))
}