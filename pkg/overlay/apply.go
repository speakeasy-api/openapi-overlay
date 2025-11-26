package overlay

import (
	"fmt"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/config"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/token"
	"gopkg.in/yaml.v3"
	"strings"
)

// ApplyTo will take an overlay and apply its changes to the given YAML
// document.
func (o *Overlay) ApplyTo(root *yaml.Node) error {
	for _, action := range o.Actions {
		var err error
		if action.Remove {
			err = o.applyRemoveAction(root, action, nil)
		} else {
			err = o.applyUpdateAction(root, action, &[]string{})
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Overlay) ApplyToStrict(root *yaml.Node) (error, []string) {
	multiError := []string{}
	warnings := []string{}
	hasFilterExpression := false
	for i, action := range o.Actions {
		tokens := token.NewTokenizer(action.Target, config.WithPropertyNameExtension()).Tokenize()
		for _, tok := range tokens {
			if tok.Token == token.FILTER {
				hasFilterExpression = true
			}
		}

		actionWarnings := []string{}
		err := o.validateSelectorHasAtLeastOneTarget(root, action)
		if err != nil {
			multiError = append(multiError, err.Error())
		}
		if action.Remove {
			err = o.applyRemoveAction(root, action, &actionWarnings)
		} else {
			err = o.applyUpdateAction(root, action, &actionWarnings)
		}
		for _, warning := range actionWarnings {
			warnings = append(warnings, fmt.Sprintf("update action (%v / %v) target=%s: %s", i+1, len(o.Actions), action.Target, warning))
		}
	}

	if hasFilterExpression && !o.UsesRFC9535() {
		warnings = append(warnings, "overlay has a filter expression but lacks `x-speakeasy-jsonpath: rfc9535` extension. Deprecated jsonpath behaviour in use. See overlay.speakeasy.com for the implementation playground.")
	}

	if len(multiError) > 0 {
		return fmt.Errorf("error applying overlay (strict): %v", strings.Join(multiError, ",")), warnings
	}
	return nil, warnings
}

func (o *Overlay) validateSelectorHasAtLeastOneTarget(root *yaml.Node, action Action) error {
	if action.Target == "" {
		return nil
	}

	p, err := o.NewPath(action.Target, nil)
	if err != nil {
		return err
	}

	nodes := p.Query(root)

	if len(nodes) == 0 {
		return fmt.Errorf("selector %q did not match any targets", action.Target)
	}

	return nil
}

func (o *Overlay) applyRemoveAction(root *yaml.Node, action Action, warnings *[]string) error {
	if action.Target == "" {
		return nil
	}

	idx := newParentIndex(root)

	p, err := o.NewPath(action.Target, warnings)
	if err != nil {
		return err
	}

	nodes := p.Query(root)
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
				if i%2 == 1 {
					// if we select a value, we should delete the key too
					parent.Content = append(parent.Content[:i-1], parent.Content[i+1:]...)
				} else {
					// if we select a key, we should delete the value
					parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
				}
				return
			case yaml.SequenceNode:
				parent.Content = append(parent.Content[:i], parent.Content[i+1:]...)
				return
			}
		}
	}
}

func (o *Overlay) applyUpdateAction(root *yaml.Node, action Action, warnings *[]string) error {
	if action.Target == "" {
		return nil
	}

	if action.Update.IsZero() {
		return nil
	}

	p, err := o.NewPath(action.Target, warnings)
	if err != nil {
		return err
	}

	nodes := p.Query(root)

	didMakeChange := false
	for _, node := range nodes {
		didMakeChange = updateNode(node, &action.Update) || didMakeChange
	}

	// If no nodes were found, try to add the key to the parent if it exists
	if !didMakeChange && len(nodes) == 0 {
		didMakeChange = o.tryAddKeyToParent(root, action.Target, &action.Update, warnings)
	}

	if !didMakeChange {
		*warnings = append(*warnings, "does nothing")
	}

	return nil
}

// tryAddKeyToParent attempts to add a new key to a parent mapping node when the
// target path doesn't exist. It parses the JSONPath to extract the parent path
// and the new key, then adds the key/value pair to the parent if it exists.
// Returns true if a change was made.
func (o *Overlay) tryAddKeyToParent(root *yaml.Node, target string, updateNode *yaml.Node, warnings *[]string) bool {
	// Tokenize the path to extract parent and key
	tokens := token.NewTokenizer(target, config.WithPropertyNameExtension()).Tokenize()
	if len(tokens) < 2 {
		return false
	}

	// Find the last meaningful token (property name or string literal in brackets)
	// We need to skip the closing bracket if present
	var propertyName string
	var parentTokens token.Tokens

	// Check if we have a bracket notation property at the end: ['key']
	if len(tokens) >= 3 {
		lastIdx := len(tokens) - 1
		if tokens[lastIdx].Token.String() == "]" &&
			tokens[lastIdx-1].Token.String() == "STRING_LITERAL" {
			// Bracket notation: ...['property']
			propertyName = tokens[lastIdx-1].Literal
			// Parent path excludes the bracket open, literal, and bracket close
			parentTokens = tokens[:lastIdx-2]
		}
	}

	// If not bracket notation, check for dot notation: .property
	if propertyName == "" && len(tokens) >= 2 {
		lastIdx := len(tokens) - 1
		if tokens[lastIdx].Token.String() == "STRING" {
			// Dot notation: ...property
			propertyName = tokens[lastIdx].Literal
			// Parent path excludes the dot and property
			if lastIdx > 0 && tokens[lastIdx-1].Token.String() == "." {
				parentTokens = tokens[:lastIdx-1]
			} else {
				parentTokens = tokens[:lastIdx]
			}
		}
	}

	if propertyName == "" || len(parentTokens) == 0 {
		// Couldn't extract a simple property access
		return false
	}

	// Build the parent path by reconstructing from tokens
	parentPath := buildParentPath(parentTokens)
	if parentPath == "" {
		return false
	}

	// Query for the parent node
	parentQueryable, err := o.NewPath(parentPath, warnings)
	if err != nil {
		return false
	}

	parentNodes := parentQueryable.Query(root)
	if len(parentNodes) == 0 {
		// Parent doesn't exist, can't add the key
		return false
	}

	// Add the key to all matching parent nodes (usually just one)
	didMakeChange := false
	for _, parentNode := range parentNodes {
		if parentNode.Kind != yaml.MappingNode {
			// Can only add keys to mapping nodes
			continue
		}

		// Check if the key already exists
		keyExists := false
		for i := 0; i < len(parentNode.Content); i += 2 {
			if parentNode.Content[i].Value == propertyName {
				keyExists = true
				break
			}
		}

		if !keyExists {
			// Add the new key/value pair
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: propertyName,
			}
			parentNode.Content = append(parentNode.Content, keyNode, clone(updateNode))
			didMakeChange = true
		}
	}

	return didMakeChange
}

// buildParentPath reconstructs a JSONPath string from tokens
func buildParentPath(tokens token.Tokens) string {
	if len(tokens) == 0 {
		return ""
	}

	var result strings.Builder
	for _, tok := range tokens {
		tokStr := tok.Token.String()
		switch tokStr {
		case "$":
			result.WriteString("$")
		case ".":
			result.WriteString(".")
		case "STRING":
			result.WriteString(tok.Literal)
		case "[":
			result.WriteString("[")
		case "]":
			result.WriteString("]")
		case "STRING_LITERAL":
			// For bracket notation properties, we need to include the quotes
			result.WriteString("'")
			result.WriteString(tok.Literal)
			result.WriteString("'")
		case ",":
			result.WriteString(",")
		case "..":
			result.WriteString("..")
		case "*":
			result.WriteString("*")
		case "NUMBER":
			result.WriteString(tok.Literal)
		default:
			// For other tokens, try to preserve the original using Literal
			if tok.Literal != "" {
				result.WriteString(tok.Literal)
			}
		}
	}

	return result.String()
}

func updateNode(node *yaml.Node, updateNode *yaml.Node) bool {
	return mergeNode(node, updateNode)
}

func mergeNode(node *yaml.Node, merge *yaml.Node) bool {
	if node.Kind != merge.Kind {
		*node = *clone(merge)
		return true
	}
	switch node.Kind {
	default:
		isChanged := node.Value != merge.Value
		node.Value = merge.Value
		return isChanged
	case yaml.MappingNode:
		return mergeMappingNode(node, merge)
	case yaml.SequenceNode:
		return mergeSequenceNode(node, merge)
	}
}

// mergeMappingNode will perform a shallow merge of the merge node into the main
// node.
func mergeMappingNode(node *yaml.Node, merge *yaml.Node) bool {
	anyChange := false
NextKey:
	for i := 0; i < len(merge.Content); i += 2 {
		mergeKey := merge.Content[i].Value
		mergeValue := merge.Content[i+1]

		for j := 0; j < len(node.Content); j += 2 {
			nodeKey := node.Content[j].Value
			if nodeKey == mergeKey {
				anyChange = mergeNode(node.Content[j+1], mergeValue) || anyChange
				continue NextKey
			}
		}

		node.Content = append(node.Content, merge.Content[i], clone(mergeValue))
		anyChange = true
	}
	return anyChange
}

// mergeSequenceNode will append the merge node's content to the original node.
func mergeSequenceNode(node *yaml.Node, merge *yaml.Node) bool {
	node.Content = append(node.Content, clone(merge).Content...)
	return true
}

func clone(node *yaml.Node) *yaml.Node {
	newNode := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor,
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
	}
	if node.Alias != nil {
		newNode.Alias = clone(node.Alias)
	}
	if node.Content != nil {
		newNode.Content = make([]*yaml.Node, len(node.Content))
		for i, child := range node.Content {
			newNode.Content[i] = clone(child)
		}
	}
	return newNode
}
