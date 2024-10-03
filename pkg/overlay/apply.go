package overlay

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"strings"
)

// ApplyTo will take an overlay and apply its changes to the given YAML
// document.
func (o *Overlay) ApplyTo(root *ast.Node) error {
	for _, action := range o.Actions {
		var err error
		if action.Remove {
			err = applyRemoveAction(root, action)
		} else {
			err = applyUpdateAction(root, action, &[]string{})
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Overlay) ApplyToStrict(root *ast.Node) (error, []string) {
	multiError := []string{}
	warnings := []string{}
	for i, action := range o.Actions {
		err := validateSelectorHasAtLeastOneTarget(root, action)
		if err != nil {
			multiError = append(multiError, err.Error())
		}
		if action.Remove {
			err = applyRemoveAction(root, action)
		} else {
			actionWarnings := []string{}
			err = applyUpdateAction(root, action, &actionWarnings)
			for _, warning := range actionWarnings {
				warnings = append(warnings, fmt.Sprintf("update action (%v / %v) target=%s: %s", i+1, len(o.Actions), action.Target, warning))
			}
		}
	}
	if len(multiError) > 0 {
		return fmt.Errorf("error applying overlay (strict): %v", strings.Join(multiError, ",")), warnings
	}
	return nil, warnings
}

func validateSelectorHasAtLeastOneTarget(root *ast.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	p, err := yaml.PathString(action.Target)
	if err != nil {
		return err
	}

	_, err = p.FilterNode(*root)
	if err != nil {
		return fmt.Errorf("selector %q did not match any targets: %w", action.Target, err)
	}

	return nil
}

func applyRemoveAction(root ast.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	idx := newParentIndex(root)

	p, err := yaml.PathString(action.Target)
	if err != nil {
		return err
	}

	filtered, err := p.FilterNode(root)
	if err != nil {
		return err
	}

	removeNode(idx, filtered)

	return nil
}

func removeNode(idx parentIndex, node ast.Node) error {
	parent := idx.getParent(node)
	if parent == nil {
		return fmt.Errorf("parent not found for node")
	}

	switch p := parent.(type) {
	case *ast.MappingNode:
		for i, value := range p.Values {
			if value.Key == node || value.Value == node {
				// we have to delete the key too
				p.Values = append(p.Values[:i-1], p.Values[i+1:]...)
				return nil
			}
		}
	case *ast.SequenceNode:
		for i, value := range p.Values {
			if value == node {
				p.Values = append(p.Values[:i], p.Values[i+1:]...)
				return nil
			}
		}
	case *ast.DocumentNode:
		if p.Body == node {
			p.Body = nil
			return nil
		}
	// Add more cases as needed for other node types
	default:
		return fmt.Errorf("unsupported parent node type: %T", parent)
	}

	return nil
}

func applyUpdateAction(root *ast.Node, action Action, warnings *[]string) error {
	if action.Target == "" {
		return nil
	}

	if action.Update.IsZero() {
		return nil
	}

	p, err := yamlpath.NewPath(action.Target)
	if err != nil {
		return err
	}

	nodes, err := p.Find(root)
	if err != nil {
		return err
	}

	prior, err := yaml.Marshal(root)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if err := updateNode(node, action.Update); err != nil {
			return err
		}
	}
	post, err := yaml.Marshal(root)
	if err != nil {
		return err
	}
	if warnings != nil && string(prior) == string(post) {
		*warnings = append(*warnings, "does nothing")
	}

	return nil
}

func updateNode(node *ast.Node, updateNode ast.Node) error {
	mergeNode(node, updateNode)
	return nil
}

func mergeNode(node *ast.Node, merge ast.Node) {
	if node.Kind != merge.Kind {
		*node = merge
		return
	}
	switch node.Kind {
	default:
		node.Value = merge.Value
	case yaml.MappingNode:
		mergeMappingNode(node, merge)
	case yaml.SequenceNode:
		mergeSequenceNode(node, merge)
	}
}

// mergeMappingNode will perform a shallow merge of the merge node into the main
// node.
func mergeMappingNode(node *ast.Node, merge ast.Node) {
NextKey:
	for i := 0; i < len(merge.Content); i += 2 {
		mergeKey := merge.Content[i].Value
		mergeValue := merge.Content[i+1]

		for j := 0; j < len(node.Content); j += 2 {
			nodeKey := node.Content[j].Value
			if nodeKey == mergeKey {
				mergeNode(node.Content[j+1], *mergeValue)
				continue NextKey
			}
		}

		node.Content = append(node.Content, merge.Content[i], mergeValue)
	}
}

// mergeSequenceNode will append the merge node's content to the original node.
func mergeSequenceNode(node *ast.Node, merge ast.Node) {
	node.Content = append(node.Content, merge.Content...)
}
