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
	// Note: this is TCR tinkering around: isn't working right now.
	filePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %q: %w", path, err)
	}
	ro, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open overlay file %q: %w", path, err)
	}
	yamlDec := yaml.NewDecoder(ro)
	var yamlNode yaml.Node
	err = yamlDec.Decode(&yamlNode)

	cfg := index.CreateOpenAPIIndexConfig()
	baseDir := filepath.Dir(filePath)
	cfg.BasePath = baseDir

	rolo := index.NewRolodex(cfg)

	localFSConf := index.LocalFSConfig{
		BaseDirectory: baseDir,
		DirFS:         os.DirFS(baseDir),
	}

	fileFS, err := index.NewLocalFSWithConfig(&localFSConf)
	if err != nil {
		return nil, err
	}
	rolo.AddLocalFS(baseDir, fileFS)

	remoteFS, _ := index.NewRemoteFSWithConfig(cfg)
	rolo.AddRemoteFS("default", remoteFS)
	rolo.SetRootNode(&yamlNode)
	//f, err := rolo.Open(path)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to open overlay file %q: %w", path, err)
	//}
	// Note: seems aggressive
	err = rolo.IndexTheRolodex()
	if err != nil {
		return nil, fmt.Errorf("failed to index overlay file %q: %w", path, err)
	}
	f, err := rolo.Open(filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to open overlay file %q: %w", path, err)
	}

	rolo.Resolve()
	errs := rolo.GetCaughtErrors()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to resolve overlay file %q: %v", path, errs)
	}

	var overlay Overlay
	resolver := f.GetIndex().GetResolver()
	resolvingErrs := resolver.Resolve()
	if len(resolvingErrs) > 0 {
		return nil, fmt.Errorf("failed to resolve overlay file %q: %v", path, resolvingErrs)
	}
	rootNode := f.GetIndex().GetRootNode()

	err = rootNode.Decode(&overlay)
	if err != nil {
		return nil, fmt.Errorf("failed to get content of overlay file %q: %w", path, err)
	}
	//err = rolo.GetContentAsYAMLNode().Decode(&overlay)
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
