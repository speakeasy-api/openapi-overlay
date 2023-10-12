package overlay_test

import (
	"github.com/speakeasy-api/openapi-specedit/pkg/overlay"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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
			Target:      `$.paths["/drinks/{name}"].get`,
			Description: "Test update",
			Update: map[string]any{
				"parameters": []any{
					map[string]any{
						"x-parameter-extension": "foo",
						"name":                  "test",
						"description":           "Test parameter",
						"in":                    "query",
						"schema": map[string]any{
							"type": "string",
						},
					},
				},
				"responses": map[string]any{
					"200": map[string]any{
						"x-response-extension": "foo",
						"description":          "Test response",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "string",
								},
							},
						},
					},
				},
			},
		},
		{
			Extensions: map[string]any{
				"x-action-extension": "bar",
			},
			Target:      `$.paths["/drinks"].get`,
			Description: "Test remove",
			Remove:      true,
		},
	},
}

func TestParse(t *testing.T) {
	f, err := os.Open("testdata/overlay.yaml")
	require.NoError(t, err)

	o, err := overlay.Parse(f)
	assert.NoError(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, expectOverlay, o)
}
