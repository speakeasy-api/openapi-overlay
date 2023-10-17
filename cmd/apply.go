package cmd

import (
	"github.com/speakeasy-api/openapi-specedit/pkg/overlay"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
)

var (
	applyCmd = &cobra.Command{
		Use:   "apply <overlay> [ <spec> ]",
		Short: "Given an overlay, it will apply it to the spec. If omitted, spec will be loaded via extends (only from local file system).",
		Args:  cobra.RangeArgs(1, 2),
		Run:   RunApply,
	}
)

func RunApply(cmd *cobra.Command, args []string) {
	overlayFile := args[0]

	ro, err := os.Open(overlayFile)
	if err != nil {
		Dief("Failed to open overlay file %q: %v", args[0], err)
	}

	o, err := overlay.Parse(ro)
	if err != nil {
		Dief("Failed to parse overlay file %q: %v", args[0], err)
	}

	specFile := ""
	if len(args) > 1 {
		specFile = args[1]
	} else {
		specUrl, err := url.Parse(o.Extends)
		if err != nil {
			Dief("Failed to parse extends URL %q: %v", o.Extends, err)
		}

		if specUrl.Scheme != "file" {
			Dief("Only file:// extends URLs are supported, not %q", o.Extends)
		}

		specFile = specUrl.Path
	}

	rs, err := os.Open(specFile)
	if err != nil {
		Dief("Failed to open spec file %q: %v", specFile, err)
	}

	var ys yaml.Node
	err = yaml.NewDecoder(rs).Decode(&ys)
	if err != nil {
		Dief("Failed to parse spec file %q: %v", specFile, err)
	}

	err = o.ApplyTo(&ys)
	if err != nil {
		Dief("Failed to apply overlay to spec file %q: %v", specFile, err)
	}

	err = yaml.NewEncoder(os.Stdout).Encode(&ys)
	if err != nil {
		Dief("Failed to encode spec file %q: %v", specFile, err)
	}
}
