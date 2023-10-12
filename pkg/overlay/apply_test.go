package overlay_test

import (
	"github.com/speakeasy-api/openapi-specedit/pkg/overlay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestApplyTo(t *testing.T) {
	openApi, err := os.Open("testdata/openapi.yaml")
	require.NoError(t, err)
	defer openApi.Close()

	var node yaml.Node
	err = yaml.NewDecoder(openApi).Decode(&node)
	require.NoError(t, err)

	oasOverlay, err := os.Open("testdata/overlay.yaml")
	require.NoError(t, err)
	defer oasOverlay.Close()

	o, err := overlay.Parse(oasOverlay)
	require.NoError(t, err)

	err = o.ApplyTo(&node)
	assert.NoError(t, err)

	actualBytes, err := yaml.Marshal(&node)
	require.NoError(t, err)

	var actual any
	err = yaml.Unmarshal(actualBytes, &actual)
	require.NoError(t, err)

	expectedIn, err := os.ReadFile("testdata/openapi-overlayed.yaml")
	require.NoError(t, err)

	var expected any
	err = yaml.Unmarshal(expectedIn, &expected)

	assert.Equal(t, expected, actual)
}
