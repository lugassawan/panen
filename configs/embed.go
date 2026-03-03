package configs

import _ "embed"

//go:embed brokers.json
var BrokersJSON []byte

//go:embed indices.json
var IndicesJSON []byte

//go:embed sectors.json
var SectorsJSON []byte
