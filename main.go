package main

import (
	"fmt"
	"github.com/speakeasy-api/openapi-specedit/pkg/overlay"
	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   "specedit",
		Short: "Work with OpenAPI Overlays",
	}

	validateCmd = &cobra.Command{
		Use:   "validate <overlay>",
		Short: "Given an overlay, it will state whether it appears to be valid or describe the problems found",
		Args:  cobra.ExactArgs(1),
		Run:   RunValidateOverlay,
	}
)

func init() {
	rootCmd.AddCommand(validateCmd)
}

func main() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func Dief(f string, args ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
	os.Exit(1)
}

func RunValidateOverlay(cmd *cobra.Command, args []string) {
	r, err := os.Open(args[0])
	if err != nil {
		Dief("Failed to open overlay file %q: %v", args[0], err)
	}

	o, err := overlay.Parse(r)
	if err != nil {
		Dief("Failed to parse overlay file %q: %v", args[0], err)
	}

	err = o.Validate()
	if err != nil {
		Dief("Overlay file %q failed validation:\n%v", args[0], err)
	}

	fmt.Printf("Overlay file %q is valid.\n", args[0])
}
