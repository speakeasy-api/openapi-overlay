package overlay_test

import (
	"fmt"
	"github.com/speakeasy-api/jsonpath/pkg/overlay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func LoadSpecification(path string) (*yaml.Node, error) {
	rs, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema from path %q: %w", path, err)
	}

	var ys yaml.Node
	err = yaml.NewDecoder(rs).Decode(&ys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema at path %q: %w", path, err)
	}

	return &ys, nil
}

func LoadOverlay(path string) (*overlay.Overlay, error) {
	o, err := overlay.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay from path %q: %w", path, err)
	}

	return o, nil
}

func TestCompare(t *testing.T) {
	t.Parallel()

	node, err := LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)
	node2, err := LoadSpecification("testdata/openapi-overlayed.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/overlay-generated.yaml")
	require.NoError(t, err)

	o2, err := overlay.Compare("Drinks Overlay", node, *node2)
	assert.NoError(t, err)

	o1s, err := o.ToString()
	assert.NoError(t, err)
	o2s, err := o2.ToString()
	assert.NoError(t, err)

	// Uncomment this if we've improved the output
	os.WriteFile("testdata/overlay-generated.yaml", []byte(o2s), 0644)
	assert.Equal(t, o1s, o2s)

	// round trip it
	err = o.ApplyTo(node)
	assert.NoError(t, err)
	NodeMatchesFile(t, node, "testdata/openapi-overlayed.yaml")

}
