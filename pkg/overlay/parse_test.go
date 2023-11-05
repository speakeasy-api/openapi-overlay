package overlay_test

import (
	"github.com/speakeasy-api/openapi-overlay/pkg/overlay"
	"github.com/stretchr/testify/assert"
	"testing"
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
		{
			Target: "$",
			Update: map[string]any{
				"info": map[string]any{
					"description": "A merged description",
				},
			},
		},
	},
}

func TestParse(t *testing.T) {
	o, err := overlay.Parse("testdata/overlay.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, expectOverlay, o)
}
