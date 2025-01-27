package overlay

import (
	"bytes"
	"fmt"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/config"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

// Extensible provides a place for extensions to be added to components of the
// Overlay configuration. These are  a map from x-* extension fields to their values.
type Extensions map[string]any

// Overlay is the top-level configuration for an OpenAPI overlay.
type Overlay struct {
	Extensions `yaml:"-,inline"`

	// Version is the version of the overlay configuration.
	// This should be set to `1.0.1` for compatability with RFC9535 (JSONPath)
	// If set to 1.0.0, the overlay will be evaluated using vmware-yamlpath behaviour.
	Version string `yaml:"overlay"`

	// JSONPathVersion should be set to rfc9535, and is used for backwards compatability purposes
	JSONPathVersion string `yaml:"x-speakeasy-jsonpath,omitempty"`

	// Info describes the metadata for the overlay.
	Info Info `yaml:"info"`

	// Extends is a URL to the OpenAPI specification this overlay applies to.
	Extends string `yaml:"extends,omitempty"`

	// Actions is the list of actions to perform to apply the overlay.
	Actions []Action `yaml:"actions"`
}

func (o *Overlay) ToString() (string, error) {
	buf := bytes.NewBuffer([]byte{})
	decoder := yaml.NewEncoder(buf)
	decoder.SetIndent(2)
	err := decoder.Encode(o)
	return buf.String(), err
}

type Queryable interface {
	Query(root *yaml.Node) []*yaml.Node
}

type yamlPathQueryable struct {
	path *yamlpath.Path
}

func (y yamlPathQueryable) Query(root *yaml.Node) []*yaml.Node {
	if y.path == nil {
		return []*yaml.Node{}
	}
	// errors aren't actually possible from yamlpath.
	result, _ := y.path.Find(root)
	return result
}

func (o *Overlay) NewPath(target string, warnings *[]string) (Queryable, error) {
	rfcJSONPath, rfcJSONPathErr := jsonpath.NewPath(target, config.WithPropertyNameExtension())
	if o.UsesRFC9535() {
		return rfcJSONPath, rfcJSONPathErr
	}
	if rfcJSONPathErr != nil && warnings != nil {
		*warnings = append(*warnings, fmt.Sprintf("invalid rfc9535 jsonpath %s: %s\nThis will be treated as an error in the future. Please fix and opt into the new implementation with `\"x-speakeasy-jsonpath\": rfc9535` in the root of your overlay. See overlay.speakeasy.com for an implementation playground.", target, rfcJSONPathErr.Error()))
	}

	path, err := yamlpath.NewPath(target)
	return mustExecute(path), err
}

func (o *Overlay) UsesRFC9535() bool {
	return o.JSONPathVersion == "rfc9535"
}

func mustExecute(path *yamlpath.Path) yamlPathQueryable {
	return yamlPathQueryable{path}
}

// Info describes the metadata for the overlay.
type Info struct {
	Extensions `yaml:"-,inline"`

	// Title is the title of the overlay.
	Title string `yaml:"title"`

	// Version is the version of the overlay.
	Version string `yaml:"version"`
}

type Action struct {
	Extensions `yaml:"-,inline"`

	// Target is the JSONPath to the target of the action.
	Target string `yaml:"target"`

	// Description is a description of the action.
	Description string `yaml:"description,omitempty"`

	// Update is the sub-document to use to merge or replace in the target. This is
	// ignored if Remove is set.
	Update yaml.Node `yaml:"update,omitempty"`

	// Remove marks the target node for removal rather than update.
	Remove bool `yaml:"remove,omitempty"`
}
