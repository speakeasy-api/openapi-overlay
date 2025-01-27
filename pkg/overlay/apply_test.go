package overlay_test

import (
	"bytes"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath"
	"github.com/speakeasy-api/openapi-overlay/pkg/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

// NodeMatchesFile is a test that marshals the YAML file from the given node,
// then compares those bytes to those found in the expected file.
func NodeMatchesFile(
	t *testing.T,
	actual *yaml.Node,
	expectedFile string,
	msgAndArgs ...any,
) {
	variadoc := func(pre ...any) []any { return append(msgAndArgs, pre...) }

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err := enc.Encode(actual)
	require.NoError(t, err, variadoc("failed to marshal node: ")...)

	expectedBytes, err := os.ReadFile(expectedFile)
	require.NoError(t, err, variadoc("failed to read expected file: ")...)

	// lazy redo snapshot
	//os.WriteFile(expectedFile, actualBuf.Bytes(), 0644)

	//t.Log("### EXPECT START ###\n" + string(expectedBytes) + "\n### EXPECT END ###\n")
	//t.Log("### ACTUAL START ###\n" + actualBuf.string() + "\n### ACTUAL END ###\n")

	assert.Equal(t, string(expectedBytes), actualBuf.String(), variadoc("node does not match expected file: ")...)
}

func TestApplyTo(t *testing.T) {
	t.Parallel()

	node, err := loader.LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)

	o, err := loader.LoadOverlay("testdata/overlay.yaml")
	require.NoError(t, err)

	err = o.ApplyTo(node)
	assert.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/openapi-overlayed.yaml")
}

func TestApplyToStrict(t *testing.T) {
	t.Parallel()

	node, err := loader.LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)

	o, err := loader.LoadOverlay("testdata/overlay-mismatched.yaml")
	require.NoError(t, err)

	err, warnings := o.ApplyToStrict(node)
	assert.Error(t, err, "error applying overlay (strict): selector \"$.unknown-attribute\" did not match any targets")
	assert.Len(t, warnings, 2)
	o.Actions = o.Actions[1:]
	node, err = loader.LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)

	err, warnings = o.ApplyToStrict(node)
	assert.NoError(t, err)
	assert.Len(t, warnings, 1)
	assert.Equal(t, "update action (2 / 2) target=$.info.title: does nothing", warnings[0])
	NodeMatchesFile(t, node, "testdata/openapi-strict-onechange.yaml")
}

func TestApplyToOld(t *testing.T) {
	t.Parallel()

	nodeOld, err := loader.LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)

	nodeNew, err := loader.LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)

	o, err := loader.LoadOverlay("testdata/overlay-old.yaml")
	require.NoError(t, err)

	err, warnings := o.ApplyToStrict(nodeOld)
	require.NoError(t, err)
	require.Len(t, warnings, 2)
	require.Contains(t, warnings[0], "invalid rfc9535 jsonpath")
	require.Contains(t, warnings[1], "x-speakeasy-jsonpath: rfc9535")

	path, err := jsonpath.NewPath(`$.paths["/anything/selectGlobalServer"]`)
	require.NoError(t, err)
	result := path.Query(nodeOld)
	require.NoError(t, err)
	require.Equal(t, 0, len(result))
	o.JSONPathVersion = "rfc9535"
	err, warnings = o.ApplyToStrict(nodeNew)
	require.ErrorContains(t, err, "unexpected token") // should error out: invalid nodepath
	// now lets fix it.
	o.Actions[0].Target = "$.paths.*[?(@[\"x-my-ignore\"])]"
	err, warnings = o.ApplyToStrict(nodeNew)
	require.ErrorContains(t, err, "did not match any targets")
	// Now lets fix it.
	o.Actions[0].Target = "$.paths[?(@[\"x-my-ignore\"])]" // @ should always refer to the child node in RFC 9535..
	err, warnings = o.ApplyToStrict(nodeNew)
	require.NoError(t, err)
	result = path.Query(nodeNew)
	require.NoError(t, err)
	require.Equal(t, 0, len(result))
}
