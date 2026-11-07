# Versioned Geyser importer

This pipeline is separate from the legacy PyMCTranslate/NEMC generator. It does not rewrite `presets/java/v1` or `presets/block-converts/v1`.

Prepare the pinned inputs from the module root (Python 3.11+ and Java 25). The preparation script writes a separate candidate lock and refuses to overwrite it. Compare that candidate with the committed lock before generation; the generator verifies every input against the committed hashes.

```sh
python third_party/gen_resource/geyser/prepare-sources.py --java-version 26.3 --mappings-commit 957beb8993a240ce7087b1b1b4266a07a5d10273 --geyser-release 2.1.1-build.1 --sources ../../sources --lock-output ../../sources/26.3.candidate.lock.json
```

Pass `--java /path/to/java` if Java 25 is not on PATH. The script verifies Mojang's server hash, runs its report generator, verifies the Geyser release archive hash, and checks its binary mappings against the pinned mappings commit. The downloaded inputs stay outside the repository.

Then, from the module root:

```sh
go run ./third_party/gen_resource/geyser/cmd/generate -lock third_party/gen_resource/geyser/versions/26.3.lock.json -sources ../../sources -output presets/java/v26_3/catalog.json.gz
go run ./third_party/gen_resource/geyser/cmd/generate -check
```

The lock fixes Java, Bedrock and DataVersion plus five input hashes: the official Mojang block report, official game version, Geyser binary mappings, Geyser named debug mappings, and Geyser version declaration.

`blocks.nbt` is a gzip-compressed big-endian NBT compound. Its `bedrock_mappings` list is indexed by Java state registry ID. A missing `bedrock_identifier` means the Java block's identifier; a missing `state` is an empty compound. NBT type information is preserved. The debug JSON is joined by full state name and checked against every binary mapping, never by JSON iteration order.

All ID ranges, names, properties, defaults, complete state products and hashes are validated. The generated portable JSON is compressed deterministically, reloaded, and written through a temporary file. Different target versions cannot overwrite an existing snapshot. `-check` only compares and writes nothing.

For a future version, prepare a new locked input set and use a new preset path. Existing consumers must opt in explicitly; updating block data alone does not update rendering assets.

See [the source notice](../../geyser-notices/NOTICE.md) for attribution.
