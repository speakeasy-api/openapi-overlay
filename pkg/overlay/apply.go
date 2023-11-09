package overlay

import (
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

// ApplyTo will take an overlay and apply its changes to the given YAML
// document.
func (o *Overlay) ApplyTo(root *yaml.Node) error {
	for _, action := range o.Actions {
		var err error
		if action.Remove {
			err = applyRemoveAction(root, action)
		} else {
			err = applyUpdateAction(root, action)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func applyRemoveAction(root *yaml.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	idx := newParentIndex(root)

	p, err := yamlpath.NewPath(action.Target)
	if err != nil {
		return err
	}

	nodes, err := p.Find(root)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		removeNode(idx, node)
	}

	return nil
}

func removeNode(idx parentIndex, node *yaml.Node) {
	parent := idx.getParent(node)
	if parent == nil {
		return
	}

	for i, child := range parent.Content {
		if child == node {
			switch parent.Kind {
			case yaml.MappingNode:
				// we have to delete the key too
				parent.Content = append(parent.Content[:i-1], parent.Content[i+1:]...)
				return
			case yaml.SequenceNode:
				parent.Content = append(parent.Content[:i], parent.Content[i+1:]...)
				return
			}
		}
	}
}

func applyUpdateAction(root *yaml.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	if action.Update == nil {
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

	for _, node := range nodes {
		if err := updateNode(node, action.Update); err != nil {
			return err
		}
	}

	return nil
}

func updateNode(node *yaml.Node, update any) error {
	updateNode, err := convertToNode(update)
	if err != nil {
		return err
	}

	mergeNode(node, updateNode)
	return nil
}

func mergeNode(node, merge *yaml.Node) {
	switch node.Kind {
	case yaml.ScalarNode:
		node.Value = merge.Value
	case yaml.MappingNode:
		if node.Kind != yaml.MappingNode {
			node.Value = merge.Value
		}
		mergeMappingNode(node, merge)
	case yaml.SequenceNode:
		mergeSequenceNode(node, merge)
	}
}

// mergeMappingNode will perform a shallow merge of the merge node into the main
// node.
func mergeMappingNode(node *yaml.Node, merge *yaml.Node) {
NextKey:
	for i := 0; i < len(merge.Content); i += 2 {
		mergeKey := merge.Content[i].Value
		mergeValue := merge.Content[i+1]

		for j := 0; j < len(node.Content); j += 2 {
			nodeKey := node.Content[j].Value
			if nodeKey == mergeKey {
				mergeNode(node.Content[j+1], mergeValue)
				continue NextKey
			}
		}

		node.Content = append(node.Content, merge.Content[i], mergeValue)
	}
}

// mergeSequenceNode will append the merge node's content to the original node.
func mergeSequenceNode(node, merge *yaml.Node) {
	node.Content = append(node.Content, merge.Content...)
}
