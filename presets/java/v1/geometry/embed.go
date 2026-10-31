package geometry

import "embed"

// Generate static geometry and matching Java metadata from the six local inputs.
// See third_party/gen_resource/geometry/solid_block_table/README.md.
//go:generate go run ../../../../third_party/gen_resource/geometry/solid_block_table

//go:embed geometry.nvtables geometry.nvtables.java.json
var Files embed.FS
