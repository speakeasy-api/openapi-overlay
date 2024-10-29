package overlay_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vnarek/openapi-overlay/pkg/overlay"
)

var expectOverlay = &overlay.Overlay{
	Extensions: map[string]any{
		"x-top-level-extension": true,
	},
	Version: "1.0.0",
	Info: overlay.Info{
		Extensions: map[string]any{
			"x-info-extension": 42,
		},
		Title:   "Drinks Overlay",
		Version: "1.2.3",
	},
	Extends: "https://raw.githubusercontent.com/speakeasy-sdks/template-sdk/main/openapi.yaml",
	Actions: []overlay.Action{
		{
			Extensions: map[string]any{
				"x-action-extension": "foo",
			},
			Target:      `$.paths["/drink/{name}"].get`,
			Description: "Test update",
		},
		{
			Extensions: map[string]any{
				"x-action-extension": "bar",
			},
			Target:      `$.paths["/drinks"].get`,
			Description: "Test remove",
			Remove:      true,
		},
		{
			Target: "$.paths[\"/drinks\"]",
		},
		{
			Target: "$.tags",
		},
	},
}

func TestParse(t *testing.T) {
	err := overlay.Format("testdata/overlay.yaml")
	require.NoError(t, err)
	o, err := overlay.Parse("testdata/overlay.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, o)
	expect, err := os.ReadFile("testdata/overlay.yaml")
	assert.NoError(t, err)

	actual, err := o.ToString()
	assert.NoError(t, err)
	assert.Equal(t, string(expect), actual)

}
