[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/speakeasy-api/openapi-overlay)](https://goreportcard.com/report/github.com/speakeasy-api/openapi-overlay)
[![GoDoc](https://godoc.org/github.com/speakeasy-api/openapi-overlay?status.svg)](https://godoc.org/github.com/speakeasy-api/openapi-overlay)


# OpenAPI Overlay

> ⚠️ This an alpha implementation. If you'd like to discuss a production use case please join the Speakeasy [slack](https://join.slack.com/t/speakeasy-dev/shared_invite/zt-1df0lalk5-HCAlpcQiqPw8vGukQWhexw).

This is an implementation of the [OpenAPI Overlay
Specification](https://github.com/OAI/Overlay-Specification/blob/3f398c6/versions/1.0.0.md)
(2023-10-12). This specification defines a means of editing a OpenAPI
Specification file by applying a list of actions. Each action is either a remove
action that prunes nodes or an update that merges a value into nodes. The nodes
impacted are selected by a target expression which uses JSONPath.

The specification itself says very little about the input file to be modified or
the output file. The presumed intention is that the input and output be an
OpenAPI Specification, but that is not required.

In many ways, this is similar to [JSONPatch](https://jsonpatch.com/), but
without the requirement to use a single explicit path for each operation. This
allows the creator of an overlay file to apply a single modification to a large
number of nodes in the file within a single operation.

This tool uses [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) to parse
the input, which implements YAML v1.2 parsing. YAML v1.2 is a superset of JSON,
so it should be able to parse either YAML or JSON with the same parser.

# Installation

Install it with the `go install` command:

```sh
go install github.com/speakeasy-api/openapi-overlay@latest
```

# Usage

The tool provides the following sub-commands for working with overlay files. It can also be accessed through Speakeasy CLI command `speakeasy overlay`. Please see [here](https://www.speakeasyapi.dev/docs/speakeasy-cli/overlay/README) for CLI installation and usage documentation. 

For more examples of usage, see [here](https://www.speakeasyapi.dev/docs/openapi/overlays)

## Apply

The most obvious use-case for this command is applying an overlay to a specification file.

```sh
openapi-overlay apply overlay.yaml spec.yaml
```

If the overlay file has the `extends` key set to a `file://` URL, then the `spec.yaml` file may be omitted.

## Validate

A command is provided to perform basic validation of the overlay file itself. It will not tell you whether it will apply correctly or whether the application will generate a valid OpenAPI specification. Rather, it is limited to just telling you when the spec follows the OpenAPI Overlay Specification correctly: all required fields are present and have valid values.

```sh
openapi-overlay validate overlay.yaml
```

## Compare

Finally, a tool is provided that will generate an OpenAPI Overlay specification from two input files.

```sh
openapi-overlay compare spec1.yaml spec2.yaml
```

The overlay file will be written to stdout.

# Other Notes

This tool works with either YAML or JSON input files, but always outputs YAML at this time.
