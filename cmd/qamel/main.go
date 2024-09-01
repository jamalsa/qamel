package main

import (
	"fmt"
	"os"

	"github.com/jamalsa/qamel/internal/cmd"
)

func main() {
	rootCmd := cmd.QamelCmd()
	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
