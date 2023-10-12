package overlay

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestConvertToNodeScalar(t *testing.T) {
	scalar := "foo"
	node, err := convertToNode(scalar)
	assert.NoError(t, err)
	assert.Equal(t, node.Kind, yaml.ScalarNode)
	assert.Equal(t, "foo", node.Value)
}

func TestConvertToNodeMap(t *testing.T) {
	m := map[string]string{"foo": "bar"}
	node, err := convertToNode(m)
	assert.NoError(t, err)
	assert.Equal(t, node.Kind, yaml.MappingNode)
	assert.Len(t, node.Content, 2)
	assert.Equal(t, "foo", node.Content[0].Value)
	assert.Equal(t, "bar", node.Content[1].Value)
}

func TestConvertToNodeSlice(t *testing.T) {
	s := []string{"foo", "bar"}
	node, err := convertToNode(s)
	assert.NoError(t, err)
	assert.Equal(t, node.Kind, yaml.SequenceNode)
	assert.Len(t, node.Content, 2)
	assert.Equal(t, "foo", node.Content[0].Value)
	assert.Equal(t, "bar", node.Content[1].Value)
}
