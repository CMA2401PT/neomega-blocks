package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"neomega-blocks/third_party/gen_resource/geometry/solid_block_table/internal/tablegen"
)

func main() {
	// Anchor defaults to this source directory so go generate is independent of cwd.
	_, source, _, _ := runtime.Caller(0)
	inputDir := filepath.Dir(source)
	moduleRoot := filepath.Clean(filepath.Join(inputDir, "../../../.."))
	assets := flag.String("assets", inputDir, "local directory containing the six matching assets files")
	output := flag.String("output", filepath.Join(moduleRoot, "presets/java/v1/geometry/geometry.nvtables"), "output geometry table")
	solid := flag.String("solid", filepath.Join(moduleRoot, "presets/java/v1/solidtable/geometry.nvsolid"), "output opaque boundary table")
	flag.Parse()
	if err := generate(*assets, *output, *solid); err != nil {
		log.Fatal(err)
	}
}

func generate(assetsDir, output, solidPath string) error {
	builder, err := tablegen.Open(tablegen.OpenOptions{AssetsDir: assetsDir})
	if err != nil {
		return err
	}
	tables, err := builder.Build()
	if err != nil {
		return err
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	if err := tables.Encode(file); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	solidFile, err := os.Create(solidPath)
	if err != nil {
		return err
	}
	if err := tables.Solid.Encode(solidFile); err != nil {
		solidFile.Close()
		return err
	}
	if err := solidFile.Close(); err != nil {
		return err
	}
	metadata, err := json.MarshalIndent(tables.Metadata, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(output+".java.json", append(metadata, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("%s: indexed=%d shapes=%d skipped=%d\n", output, len(tables.Index), len(tables.Shapes), len(tables.Skipped))
	fmt.Printf("%s: entries=%d\n", solidPath, len(tables.Solid))
	return nil
}
