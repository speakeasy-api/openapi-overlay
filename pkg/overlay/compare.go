package overlay

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"path/filepath"
	"strings"
)

type namedReader interface {
	Name() string
}

func fileName(r io.Reader, fallback string) string {
	if n, isNamed := r.(namedReader); isNamed {
		return n.Name()
	}
	return fallback
}

func loadSpec(r io.Reader) (*yaml.Node, error) {
	c, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var y yaml.Node
	err = yaml.Unmarshal(c, &y)
	return &y, err
}

// Compare compares input specifications from two files and returns an overlay
// that will convert the first into the second.
func Compare(r1, r2 io.Reader) (*Overlay, error) {
	t1 := fileName(r1, "first YAML file")
	hasExtends := t1 != "first YAML file"
	y1, err := loadSpec(r1)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", t1, err)
	}

	t2 := fileName(r2, "second YAML file")
	y2, err := loadSpec(r2)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", t2, err)
	}

	actions, err := walkTreesAndCollectActions(simplePath{}, y1, y2)
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("Overlay %s => %s", t1, t2)

	extends := ""
	if hasExtends {
		abs, err := filepath.Abs(t1)
		if err == nil {
			extends = "file://" + abs
		}
	}

	return &Overlay{
		Version: "1.0.0",
		Info: Info{
			Title:   title,
			Version: "0.0.0",
		},
		Extends: extends,
		Actions: actions,
	}, nil
}

type simplePart struct {
	isKey bool
	key   string
	index int
}

func intPart(index int) simplePart {
	return simplePart{
		index: index,
	}
}

func keyPart(key string) simplePart {
	return simplePart{
		isKey: true,
		key:   key,
	}
}

func (p simplePart) String() string {
	if p.isKey {
		return fmt.Sprintf("[%q]", p.key)
	}
	return fmt.Sprintf("[%d]", p.index)
}

func (p simplePart) KeyString() string {
	if p.isKey {
		return p.key
	}
	panic("FIXME: Bug detected in overlay comparison algorithm: attempt to use non key part as key")
}

type simplePath []simplePart

func (p simplePath) WithIndex(index int) simplePath {
	return append(p, intPart(index))
}

func (p simplePath) WithKey(key string) simplePath {
	return append(p, keyPart(key))
}

func (p simplePath) ToJSONPath() string {
	out := &strings.Builder{}
	out.WriteString("$")
	for _, part := range p {
		out.WriteString(part.String())
	}
	return out.String()
}

func (p simplePath) Dir() simplePath {
	return p[:len(p)-1]
}

func (p simplePath) Base() simplePart {
	return p[len(p)-1]
}

func walkTreesAndCollectActions(path simplePath, y1, y2 *yaml.Node) ([]Action, error) {
	if y1 == nil {
		update, err := convertFromNode(y2)
		if err != nil {
			return nil, err
		}

		return []Action{{
			Target: path.Dir().ToJSONPath(),
			Update: map[string]any{
				path.Base().KeyString(): update,
			},
		}}, nil
	}

	if y2 == nil {
		return []Action{{
			Target: path.ToJSONPath(),
			Remove: true,
		}}, nil
	}

	switch y1.Kind {
	case yaml.DocumentNode:
		return walkTreesAndCollectActions(path, y1.Content[0], y2.Content[0])
	case yaml.SequenceNode:
		if y2.Kind == yaml.SequenceNode && len(y2.Content) == len(y1.Content) {
			return walkSequenceNode(path, y1, y2)
		}

		update, err := convertFromNode(y2)
		if err != nil {
			return nil, err
		}

		return []Action{{
			Target: path.ToJSONPath(),
			Update: update,
		}}, nil
	case yaml.MappingNode:
		if y2.Kind == yaml.MappingNode {
			return walkMappingNode(path, y1, y2)
		}

		update, err := convertFromNode(y2)
		if err != nil {
			return nil, err
		}

		return []Action{{
			Target: path.ToJSONPath(),
			Update: update,
		}}, nil
	case yaml.ScalarNode:
		if y1.Value != y2.Value {
			update, err := convertFromNode(y2)
			if err != nil {
				return nil, err
			}

			return []Action{{
				Target: path.ToJSONPath(),
				Update: update,
			}}, nil
		}
	case yaml.AliasNode:
		log.Println("YAML alias nodes are not yet supported for compare.")
	}
	return nil, nil
}

func walkSequenceNode(path simplePath, y1, y2 *yaml.Node) ([]Action, error) {
	nodeLen := max(len(y1.Content), len(y2.Content))
	var actions []Action
	for i := 0; i < nodeLen; i++ {
		var c1, c2 *yaml.Node
		if i < len(y1.Content) {
			c1 = y1.Content[i]
		}
		if i < len(y2.Content) {
			c2 = y2.Content[i]
		}

		newActions, err := walkTreesAndCollectActions(
			path.WithIndex(i),
			c1, c2)
		if err != nil {
			return nil, err
		}

		actions = append(actions, newActions...)
	}

	return actions, nil
}

func walkMappingNode(path simplePath, y1, y2 *yaml.Node) ([]Action, error) {
	var actions []Action
	foundKeys := map[string]struct{}{}

	// Add or update keys in y2 that differ/missing from y1
Outer:
	for i := 0; i < len(y2.Content); i += 2 {
		k2 := y2.Content[i]
		v2 := y2.Content[i+1]

		foundKeys[k2.Value] = struct{}{}

		// find keys in y1 to update
		for j := 0; j < len(y1.Content); j += 2 {
			k1 := y1.Content[j]
			v1 := y1.Content[j+1]

			if k1.Value == k2.Value {
				newActions, err := walkTreesAndCollectActions(
					path.WithKey(k2.Value),
					v1, v2)
				if err != nil {
					return nil, err
				}
				actions = append(actions, newActions...)
				continue Outer
			}
		}

		// key not found in y1, so add it
		newActions, err := walkTreesAndCollectActions(
			path.WithKey(k2.Value),
			nil, v2)
		if err != nil {
			return nil, err
		}

		actions = append(actions, newActions...)
	}

	// look for keys in y1 that are not in y2: remove them
	for i := 0; i < len(y1.Content); i += 2 {
		k1 := y1.Content[i]

		if _, alreadySeen := foundKeys[k1.Value]; alreadySeen {
			continue
		}

		actions = append(actions, Action{
			Target: path.WithKey(k1.Value).ToJSONPath(),
			Remove: true,
		})
	}

	return actions, nil
}
