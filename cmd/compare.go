package cmd

import (
	"fmt"
	"github.com/speakeasy-api/openapi-overlay/pkg/loader"
	"github.com/speakeasy-api/openapi-overlay/pkg/overlay"
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
	y1, err := loader.LoadSpecification(args[0])
	if err != nil {
		Dief("Failed to load %q: %v", args[0], err)
	}

	y2, err := loader.LoadSpecification(args[1])
	if err != nil {
		Dief("Failed to load %q: %v", args[1], err)
	}

	title := fmt.Sprintf("Overlay %s => %s", args[0], args[1])

	o, err := overlay.Compare(title, args[0], y1, *y2)
	if err != nil {
		Dief("Failed to compare spec files %q and %q: %v", args[0], args[1], err)
	}

	err = o.Format(os.Stdout)
	if err != nil {
		Dief("Failed to format overlay: %v", err)
	}
}
