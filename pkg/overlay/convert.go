package overlay

import "gopkg.in/yaml.v3"

// convertToNode will convert any suitable Go data structure into a *yaml.Node by
// marshalling and then unmarshalling the value.
func convertToNode(data any) (*yaml.Node, error) {
	dataYaml, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	var docNode yaml.Node
	err = yaml.Unmarshal(dataYaml, &docNode)
	if err != nil {
		return nil, err
	}

	// docNode will always be a document containing a single node, which is the one
	// we actually want
	dataNode := docNode.Content[0]

	return dataNode, nil
}