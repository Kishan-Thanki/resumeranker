package docs

import _ "embed"

//go:embed internal-api.yaml
var InternalAPIYAML []byte

//go:embed public-api-analysis.yaml
var PublicAPIYAML []byte
