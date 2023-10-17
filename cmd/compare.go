package cmd

import (
	"github.com/speakeasy-api/openapi-specedit/pkg/overlay"
	"github.com/spf13/cobra"
	"os"
)

var (
	compareCmd = &cobra.Command{
		Use:   "compare <spec1> <spec2>",
		Short: "Given two specs, it will output an overlay that describes the differences between them",
		Args:  cobra.ExactArgs(2),
		Run:   RunCompare,
	}
)

func RunCompare(cmd *cobra.Command, args []string) {
	r1, err := os.Open(args[0])
	if err != nil {
		Dief("Failed to open spec file %q: %v", args[0], err)
	}

	r2, err := os.Open(args[1])
	if err != nil {
		Dief("Failed to open spec file %q: %v", args[1], err)
	}

	o, err := overlay.Compare(r1, r2)
	if err != nil {
		Dief("Failed to compare spec files %q and %q: %v", args[0], args[1], err)
	}

	err = o.Format(os.Stdout)
	if err != nil {
		Dief("Failed to format overlay: %v", err)
	}
}
