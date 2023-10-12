# OpenAPI SpecEdit

This is an implementation of the [OpenAPI Overlay
Specification](https://github.com/OAI/Overlay-Specification/blob/3f398c6/versions/1.0.0.md)
(2023-10-12). This specification defines a means of editing a OpenAPI
Specificaiton file by applying a list of actions. Each action is either a remove
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
