package overlay

import (
	"fmt"
	"github.com/pb33f/libopenapi/index"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
)

// Parse will parse the given reader as an overlay file.
func Parse(path string) (*Overlay, error) {
	filePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %q: %w", path, err)
	}

	ro, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open overlay file at path %q: %w", path, err)
	}

	var overlay Overlay
	dec := yaml.NewDecoder(ro)
	var rootNode yaml.Node

	err = dec.Decode(&rootNode)
	if err != nil {
		return nil, err
	}

	cfg := index.CreateOpenAPIIndexConfig()
	cfg.BasePath = filepath.Dir(filePath)
	idx := index.NewSpecIndexWithConfig(&rootNode, cfg)
	referenceErrors := idx.GetReferenceIndexErrors()
	if len(referenceErrors) > 0 {
		msg := ""
		for _, err := range referenceErrors {
			msg += err.Error() + ";"
		}
		return nil, fmt.Errorf("error indexing spec: %s", msg)
	}

	resolverRef := idx.GetResolver()
	resolvingErrors := resolverRef.Resolve()
	// any errors found during resolution? Print them out.
	if len(resolvingErrors) > 0 {
		msg := ""
		for _, err := range resolvingErrors {
			msg += err.Error() + ";"
		}
		return nil, fmt.Errorf("error resolving spec: %s", msg)
	}

	err = idx.GetRootNode().Decode(&overlay)
	if err != nil {
		return nil, err
	}

	return &overlay, err
}

// Format writes the file back out as YAML.
func (o *Overlay) Format(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(o)
}
