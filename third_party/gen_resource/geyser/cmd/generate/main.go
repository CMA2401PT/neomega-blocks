package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"neomega-blocks/third_party/gen_resource/geyser"
)

func run() error {
	lockPath := flag.String("lock", "third_party/gen_resource/geyser/versions/26.3.lock.json", "pinned input lock")
	sources := flag.String("sources", "../../sources", "local input root")
	output := flag.String("output", "presets/java/v26_3/catalog.json.gz", "snapshot destination")
	check := flag.Bool("check", false, "verify deterministic output without writing")
	flag.Parse()
	lockJSON, err := os.ReadFile(*lockPath)
	if err != nil {
		return err
	}
	var lock geyser.Lock
	decoder := json.NewDecoder(bytes.NewReader(lockJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&lock); err != nil {
		return err
	}
	snapshot, err := geyser.Generate(lock, *sources)
	if err != nil {
		return err
	}
	data, err := geyser.Encode(snapshot)
	if err != nil {
		return err
	}
	if *check {
		current, err := os.ReadFile(*output)
		if err != nil {
			return err
		}
		if !bytes.Equal(current, data) {
			return fmt.Errorf("snapshot differs: %s", *output)
		}
	} else if err := geyser.WriteAtomic(*output, data); err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"javaVersion": lock.JavaVersion, "bedrockVersion": lock.BedrockVersion,
		"blocks": len(snapshot.Blocks), "states": len(snapshot.States), "bytes": len(data),
		"sha256": hex.EncodeToString(sum[:]), "check": *check,
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
