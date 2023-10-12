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
			parent.Content = append(parent.Content[:i], parent.Content[i+1:]...)
			return
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

func updateNode(node *yaml.Node, merge any) error {
	mergeNode, err := convertToNode(merge)
	if err != nil {
		return err
	}

	switch mergeNode.Kind {
	case yaml.ScalarNode:
		node.Value = mergeNode.Value
	case yaml.MappingNode:
		if node.Kind != yaml.MappingNode {
			node.Value = mergeNode.Value
		}
		mergeMappingNode(node, mergeNode)
	case yaml.SequenceNode:
		// TODO should sequence nodes be merged too?
		node.Value = mergeNode.Value
	}

	return nil
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
				node.Content[j+1] = mergeValue
				continue NextKey
			}
		}

		node.Content = append(node.Content, merge.Content[i], mergeValue)
	}
}
