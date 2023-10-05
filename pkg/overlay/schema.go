package overlay

type Extensible struct {
	Extensions map[string]any
}

type Overlay struct {
	Extensible `yaml:"-"`

	Version string   `yaml:"overlay"`
	Info    Info     `yaml:"info"`
	Extends string   `yaml:"extends"`
	Actions []Action `yaml:"actions"`
}

type Info struct {
	Extensible `yaml:"-"`

	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

type Action struct {
	Extensible `yaml:"-"`

	Target      string `yaml:"target"`
	Description string `yaml:"description"'`
	Update      any    `yaml:"update"`
	Remove      bool   `yaml:"remove"`
}
