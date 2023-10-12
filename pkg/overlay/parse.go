package overlay

import (
	"gopkg.in/yaml.v3"
	"io"
)

// Parse will parse the given reader as an overlay file.
func Parse(r io.Reader) (*Overlay, error) {
	var overlay Overlay
	dec := yaml.NewDecoder(r)
	err := dec.Decode(&overlay)
	return &overlay, err
}
