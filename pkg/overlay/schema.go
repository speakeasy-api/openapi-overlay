package overlay

// Extensible provides a place for extensions to be added to components of the
// Overlay configuration. These are  a map from x-* extension fields to their values.
type Extensions map[string]any

// Overlay is the top-level configuration for an OpenAPI overlay.
type Overlay struct {
	Extensions `yaml:"-,inline"`

	// Version is the version of the overlay configuration. As the RFC was never
	// really ratifies, this value does not mean much.
	Version string `yaml:"overlay"`

	// Info describes the metadata for the overlay.
	Info Info `yaml:"info"`

	// Extends is a URL to the OpenAPI specification this overlay applies to.
	Extends string `yaml:"extends"`

	// Actions is the list of actions to perform to apply the overlay.
	Actions []Action `yaml:"actions"`
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
	Description string `yaml:"description"'`

	// Update is the sub-document to use to merge or replace in the target. This is
	// ignored if Remove is set.
	Update any `yaml:"update"`

	// Remove marks the target node for removal rather than update.
	Remove bool `yaml:"remove"`
}
