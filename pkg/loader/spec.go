package loader

import (
	"bytes"
	"fmt"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/utils"
	"github.com/speakeasy-api/openapi-overlay/pkg/overlay"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
)

// GetOverlayExtendsPath returns the path to file if the extends URL is a file
// URL. Otherwise, returns an empty string and an error. The error may occur if
// no extends URL is present or if the URL is not a file URL or if the URL is
// malformed.
func GetOverlayExtendsPath(o *overlay.Overlay) (string, error) {
	if o.Extends == "" {
		return "", fmt.Errorf("overlay does not specify an extends URL")
	}

	specUrl, err := url.Parse(o.Extends)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %q: %w", o.Extends, err)
	}

	if specUrl.Scheme != "file" {
		return "", fmt.Errorf("only file:// extends URLs are supported, not %q", o.Extends)
	}

	return specUrl.Path, nil
}

// LoadExtendsSpecification will load and parse a YAML or JSON file as specified
// in the extends parameter of the overlay. Currently, this only supports file
// URLs.
func LoadExtendsSpecification(o *overlay.Overlay) (*yaml.Node, error) {
	path, err := GetOverlayExtendsPath(o)
	if err != nil {
		return nil, err
	}

	return LoadSpecification(path)
}

// LoadSpecification will load and parse a YAML or JSON file from the given path.
func LoadSpecification(path string) (*yaml.Node, error) {
	rs, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema from path %q: %w", path, err)
	}

	if utils.IsYAML(path) {
		var node yaml.Node
		if err := yaml.NewDecoder(rs).Decode(&node); err != nil {
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}
		return &node, err
	} else {
		// We use libopenapi in order to preserve JSON key ordering
		var fileBytes bytes.Buffer
		if _, err := io.Copy(&fileBytes, rs); err != nil {
			return nil, err
		}

		doc, err := libopenapi.NewDocument(fileBytes.Bytes())
		if err != nil {
			return nil, err
		}

		v3Model, errs := doc.BuildV3Model()
		if len(errs) > 0 {
			return nil, fmt.Errorf("failed to build V3 model: %v", errs)
		}

		return v3Model.Index.GetRootNode(), nil
	}
}

// LoadEitherSpecification is a convenience function that will load a
// specification from the given file path if it is non-empty. Otherwise, it will
// attempt to load the path from the overlay's extends URL. Also returns the name
// of the file loaded.
func LoadEitherSpecification(path string, o *overlay.Overlay) (*yaml.Node, string, error) {
	var (
		y   *yaml.Node
		err error
	)

	if path != "" {
		y, err = LoadSpecification(path)
	} else {
		path, _ = GetOverlayExtendsPath(o)
		y, err = LoadExtendsSpecification(o)
	}

	return y, path, err
}
