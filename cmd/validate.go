package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vnarek/openapi-overlay/pkg/loader"
)

var (
	validateCmd = &cobra.Command{
		Use:   "validate <overlay>",
		Short: "Given an overlay, it will state whether it appears to be valid or describe the problems found",
		Args:  cobra.ExactArgs(1),
		Run:   RunValidateOverlay,
	}
)

func RunValidateOverlay(cmd *cobra.Command, args []string) {
	o, err := loader.LoadOverlay(args[0])
	if err != nil {
		Die(err)
	}

	err = o.Validate()
	if err != nil {
		Dief("Overlay file %q failed validation:\n%v", args[0], err)
	}

	fmt.Printf("Overlay file %q is valid.\n", args[0])
}
