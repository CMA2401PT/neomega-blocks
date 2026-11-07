package geyser

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"neomega-blocks/catalog"
	nbt "neomega-nbt/presets/v1"
)

func inputs(t *testing.T) (catalog.Metadata, []byte, []byte, []byte) {
	t.Helper()
	meta := catalog.Metadata{FormatVersion: 1, JavaVersion: "test", BedrockVersion: "test", DataVersion: 1}
	// Alphabetic name order is deliberately different from registry ID order.
	report := []byte(`{"minecraft:stone":{"states":[{"id":1,"default":true}]},"minecraft:air":{"states":[{"id":0,"default":true}]}}`)
	debug := []byte(`{"minecraft:stone":{"state":{"direction":2,"flag":1}},"minecraft:air":{}}`)
	raw, err := nbt.EncodeDocument(nbt.Document{Value: nbt.Compound{"bedrock_mappings": []nbt.Compound{
		{}, {"state": nbt.Compound{"direction": int32(2), "flag": int8(1)}},
	}}}, nbt.EncodingBigEndian)
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	z := gzip.NewWriter(&buffer)
	z.Write(raw)
	z.Close()
	return meta, report, buffer.Bytes(), debug
}

func TestRegistryIDJoinAndDeterminism(t *testing.T) {
	meta, report, mappings, debug := inputs(t)
	snapshot, err := Build(meta, report, mappings, debug)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.States[0].Java != "minecraft:air []" || snapshot.States[1].Bedrock != `minecraft:stone ["direction"=2,"flag"=true]` {
		t.Fatalf("wrong typed mapping: %+v", snapshot.States)
	}
	a, err := Encode(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Encode(snapshot)
	if err != nil || !bytes.Equal(a, b) {
		t.Fatal("generation is not deterministic")
	}
	if _, err = catalog.Load(a[:len(a)-4]); err == nil {
		t.Fatal("truncated gzip accepted")
	}
}

func TestRejectSourceMisalignment(t *testing.T) {
	meta, report, mappings, debug := inputs(t)
	cases := map[string]struct{ report, mappings, debug []byte }{
		"wrong registry order": {bytes.ReplaceAll(report, []byte(`"id":1`), []byte(`"id":0`)), mappings, debug},
		"missing default":      {bytes.ReplaceAll(report, []byte(`"default":true`), []byte(`"default":false`)), mappings, debug},
		"debug mismatch":       {report, mappings, bytes.ReplaceAll(debug, []byte(`"direction":2`), []byte(`"direction":3`))},
		"missing named map":    {report, mappings, []byte(`{}`)},
		"truncated nbt":        {report, mappings[:len(mappings)/2], debug},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Build(meta, values.report, values.mappings, values.debug); err == nil {
				t.Fatal("invalid sources accepted")
			}
		})
	}
}

func TestLockedInputsAndAtomicFailure(t *testing.T) {
	meta, report, mappings, debug := inputs(t)
	snapshot, err := Build(meta, report, mappings, debug)
	if err != nil {
		t.Fatal(err)
	}
	data, err := Encode(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	path := filepath.Join(root, "catalog.json.gz")
	if err := WriteAtomic(path, data); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, []byte("broken")); err == nil {
		t.Fatal("invalid output published")
	}
	unchanged, _ := os.ReadFile(path)
	if !bytes.Equal(unchanged, data) {
		t.Fatal("failed generation damaged previous output")
	}
	snapshot.Metadata.JavaVersion = "next"
	next, err := Encode(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, next); err == nil {
		t.Fatal("different version replaced existing snapshot")
	}
	unchanged, _ = os.ReadFile(path)
	if !bytes.Equal(unchanged, data) {
		t.Fatal("version mismatch damaged existing snapshot")
	}
	if err := WriteAtomic(path, data); err != nil {
		t.Fatalf("replacing existing snapshot: %v", err)
	}
	sum := sha256.Sum256(data)
	input := Input{Path: "catalog.json.gz", Source: catalog.Source{SHA256: hex.EncodeToString(sum[:])}}
	if _, err := ReadLocked(root, input); err != nil {
		t.Fatal(err)
	}
	input.Source.SHA256 = string(bytes.Repeat([]byte("0"), 64))
	if _, err := ReadLocked(root, input); err == nil {
		t.Fatal("checksum mismatch ignored")
	}
	input.Path = "../outside.json"
	if _, err := ReadLocked(root, input); err == nil {
		t.Fatal("source root escape accepted")
	}
	files, _ := filepath.Glob(filepath.Join(root, ".catalog-*.tmp"))
	if len(files) != 0 {
		t.Fatal("staging files left behind")
	}
}

func TestVersionMismatch(t *testing.T) {
	meta, report, mappings, debug := inputs(t)
	files := map[string][]byte{"blocksReport": report, "blocksNBT": mappings, "blocksDebug": debug,
		"gameVersion": []byte(`{"id":"other","world_version":1}`), "mappingsReadme": []byte("Minecraft: Java Edition test and Minecraft: Bedrock Edition test.")}
	root := t.TempDir()
	lock := Lock{FormatVersion: 1, JavaVersion: meta.JavaVersion, BedrockVersion: meta.BedrockVersion, DataVersion: 1, Inputs: make(map[string]Input)}
	for key, data := range files {
		if err := os.WriteFile(filepath.Join(root, key), data, 0644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		lock.Inputs[key] = Input{Path: key, Source: catalog.Source{SHA256: hex.EncodeToString(sum[:])}}
	}
	if _, err := Generate(lock, root); err == nil {
		t.Fatal("mismatched official version accepted")
	}
	// Lock serialization is part of the future-version workflow.
	if _, err := json.Marshal(lock); err != nil {
		t.Fatal(err)
	}
}
