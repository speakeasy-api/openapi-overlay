package loader

import (
	"fmt"
	"github.com/speakeasy-api/openapi-specedit/pkg/overlay"
	"os"
)

// LoadOverlay is a tool for loading and parsing an overlay file from the file
// system.
func LoadOverlay(path string) (*overlay.Overlay, error) {
	ro, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open overlay file at path %q: %w", path, err)
	}

	o, err := overlay.Parse(ro)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay from path %q: %w", path, err)
	}

	return o, nil
}
