// Package geyser imports pinned Mojang reports and Geyser mappings.
// Generation is offline: callers supply files whose SHA-256 values match a lock.
package geyser

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"neomega-blocks/catalog"
	"neomega-blocks/ir"
	nbt "neomega-nbt/presets/v1"
)

type Input struct {
	Path   string         `json:"path"`
	Source catalog.Source `json:"source"`
}

type Lock struct {
	FormatVersion  int              `json:"formatVersion"`
	JavaVersion    string           `json:"javaVersion"`
	BedrockVersion string           `json:"bedrockVersion"`
	DataVersion    int              `json:"dataVersion"`
	Inputs         map[string]Input `json:"inputs"`
}

type reportBlock struct {
	Properties map[string][]string `json:"properties"`
	States     []reportState       `json:"states"`
}

type reportState struct {
	ID         int               `json:"id"`
	Default    bool              `json:"default"`
	Properties map[string]string `json:"properties"`
}

type debugMapping struct {
	Identifier string         `json:"bedrock_identifier"`
	State      map[string]any `json:"state"`
}

// ReadLocked rejects changed inputs and source paths escaping the supplied root.
func ReadLocked(root string, input Input) ([]byte, error) {
	if !filepath.IsLocal(input.Path) {
		return nil, fmt.Errorf("source path must stay inside sources root: %s", input.Path)
	}
	if hash, err := hex.DecodeString(input.Source.SHA256); err != nil || len(hash) != sha256.Size {
		return nil, fmt.Errorf("invalid SHA-256 for %s", input.Path)
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(input.Path)))
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != input.Source.SHA256 {
		return nil, fmt.Errorf("SHA-256 mismatch: %s", input.Path)
	}
	return data, nil
}

// Generate joins the NBT list by official registry ID, then independently checks
// every result against Geyser's named debug mapping. JSON debug order is ignored.
func Generate(lock Lock, root string) (catalog.Snapshot, error) {
	if lock.FormatVersion != 1 {
		return catalog.Snapshot{}, fmt.Errorf("unsupported lock format %d", lock.FormatVersion)
	}
	keys := []string{"blocksReport", "blocksNBT", "blocksDebug", "gameVersion", "mappingsReadme"}
	inputs := make(map[string][]byte)
	meta := catalog.Metadata{FormatVersion: catalog.FormatVersion, JavaVersion: lock.JavaVersion, BedrockVersion: lock.BedrockVersion, DataVersion: lock.DataVersion}
	for _, key := range keys {
		input, ok := lock.Inputs[key]
		if !ok {
			return catalog.Snapshot{}, fmt.Errorf("lock missing %s", key)
		}
		data, err := ReadLocked(root, input)
		if err != nil {
			return catalog.Snapshot{}, fmt.Errorf("%s: %w", key, err)
		}
		inputs[key] = data
		meta.Sources = append(meta.Sources, input.Source)
	}
	var version struct {
		ID          string `json:"id"`
		DataVersion int    `json:"world_version"`
	}
	if err := json.Unmarshal(inputs["gameVersion"], &version); err != nil {
		return catalog.Snapshot{}, err
	}
	if version.ID != meta.JavaVersion || version.DataVersion != meta.DataVersion {
		return catalog.Snapshot{}, fmt.Errorf("official game version differs from lock")
	}
	declaration := "Minecraft: Java Edition " + meta.JavaVersion + " and Minecraft: Bedrock Edition " + meta.BedrockVersion + "."
	if !strings.Contains(string(inputs["mappingsReadme"]), declaration) {
		return catalog.Snapshot{}, fmt.Errorf("Geyser version declaration differs from lock")
	}
	return Build(meta, inputs["blocksReport"], inputs["blocksNBT"], inputs["blocksDebug"])
}

func Build(meta catalog.Metadata, reportJSON, mappingNBT, debugJSON []byte) (catalog.Snapshot, error) {
	var report map[string]reportBlock
	if err := json.Unmarshal(reportJSON, &report); err != nil {
		return catalog.Snapshot{}, err
	}
	var debug map[string]debugMapping
	if err := json.Unmarshal(debugJSON, &debug); err != nil {
		return catalog.Snapshot{}, err
	}
	z, err := gzip.NewReader(bytes.NewReader(mappingNBT))
	if err != nil {
		return catalog.Snapshot{}, err
	}
	raw, err := io.ReadAll(z)
	z.Close()
	if err != nil {
		return catalog.Snapshot{}, err
	}
	doc, err := nbt.DecodeDocument(raw, nbt.EncodingBigEndian)
	if err != nil {
		return catalog.Snapshot{}, err
	}
	mappings, ok := nbt.Compounds(doc.Value["bedrock_mappings"])
	if !ok || len(mappings) == 0 {
		return catalog.Snapshot{}, fmt.Errorf("missing bedrock_mappings compound list")
	}
	result := catalog.Snapshot{Metadata: meta, States: make([]catalog.State, len(mappings))}
	seen := make([]bool, len(mappings))
	usedDebug := make(map[string]bool)
	names := make([]string, 0, len(report))
	for name := range report {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		block := report[name]
		definition := catalog.Definition{Name: name, Properties: block.Properties, DefaultStateID: -1}
		defaults := 0
		for _, state := range block.States {
			if state.ID < 0 || state.ID >= len(mappings) || seen[state.ID] {
				return catalog.Snapshot{}, fmt.Errorf("duplicate/out-of-range official state ID %d", state.ID)
			}
			seen[state.ID] = true
			if state.Default {
				defaults++
				definition.DefaultStateID = state.ID
			}
			properties := make(map[string]any, len(state.Properties))
			for key, value := range state.Properties {
				properties[key] = value
			}
			java, err := ir.NewIRBlockFromGo(name, properties)
			if err != nil {
				return catalog.Snapshot{}, err
			}
			java = java.Tidy()
			entry := mappings[state.ID]
			bedrockName := name
			if identifier, present := entry["bedrock_identifier"]; present {
				text, ok := identifier.(string)
				if !ok || text == "" {
					return catalog.Snapshot{}, fmt.Errorf("invalid Bedrock identifier at %d", state.ID)
				}
				bedrockName = namespaced(text)
			}
			bedrockProps := map[string]any{}
			if value, present := entry["state"]; present {
				bedrockProps, ok = value.(map[string]any)
				if !ok {
					return catalog.Snapshot{}, fmt.Errorf("invalid Bedrock compound at %d", state.ID)
				}
			}
			for key, value := range bedrockProps {
				switch value.(type) {
				case int8, int32, string:
				default:
					return catalog.Snapshot{}, fmt.Errorf("unsupported NBT property %s (%T)", key, value)
				}
			}
			bedrock, err := ir.NewIRBlockFromGo(bedrockName, bedrockProps)
			if err != nil {
				return catalog.Snapshot{}, err
			}
			debugKey := javaKey(name, state.Properties)
			check, exists := debug[debugKey]
			if !exists {
				return catalog.Snapshot{}, fmt.Errorf("missing Geyser named state %s", debugKey)
			}
			usedDebug[debugKey] = true
			expectedName := name
			if check.Identifier != "" {
				expectedName = namespaced(check.Identifier)
			}
			if expectedName != bedrockName || len(check.State) != len(bedrockProps) {
				return catalog.Snapshot{}, fmt.Errorf("Geyser ID/name alignment failed: %d %s", state.ID, debugKey)
			}
			for key, value := range bedrockProps {
				if !sameDebugValue(check.State[key], value) {
					return catalog.Snapshot{}, fmt.Errorf("Geyser property alignment failed: %d %s.%s", state.ID, debugKey, key)
				}
			}
			result.States[state.ID] = catalog.State{ID: state.ID, Java: java, Bedrock: bedrock.Tidy()}
		}
		if defaults != 1 {
			return catalog.Snapshot{}, fmt.Errorf("%s has %d default states", name, defaults)
		}
		result.Blocks = append(result.Blocks, definition)
	}
	for id, exists := range seen {
		if !exists {
			return catalog.Snapshot{}, fmt.Errorf("official registry is missing ID %d", id)
		}
	}
	if len(usedDebug) != len(debug) {
		return catalog.Snapshot{}, fmt.Errorf("unmatched named Geyser states: %d", len(debug)-len(usedDebug))
	}
	if _, err := catalog.New(result); err != nil {
		return catalog.Snapshot{}, err
	}
	return result, nil
}

func sameDebugValue(debug, typed any) bool {
	switch value := typed.(type) {
	case string:
		text, ok := debug.(string)
		return ok && text == value
	case int8:
		number, ok := debug.(float64)
		return ok && number == float64(value)
	case int32:
		number, ok := debug.(float64)
		return ok && number == float64(value)
	default:
		return false
	}
}

func namespaced(name string) string {
	if !strings.Contains(name, ":") {
		return "minecraft:" + name
	}
	return name
}

func javaKey(name string, props map[string]string) string {
	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	fields := make([]string, 0, len(keys))
	for _, key := range keys {
		fields = append(fields, key+"="+props[key])
	}
	if len(fields) == 0 {
		return name
	}
	return name + "[" + strings.Join(fields, ",") + "]"
}

// Encode creates reproducible bytes: sorted JSON map keys and zero gzip timestamp.
func Encode(snapshot catalog.Snapshot) ([]byte, error) {
	if _, err := catalog.New(snapshot); err != nil {
		return nil, err
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	z, err := gzip.NewWriterLevel(&buffer, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err = z.Write(data); err != nil {
		return nil, err
	}
	if err = z.Close(); err != nil {
		return nil, err
	}
	if _, err = catalog.Load(buffer.Bytes()); err != nil {
		return nil, fmt.Errorf("generated snapshot reload failed: %w", err)
	}
	return buffer.Bytes(), nil
}

// WriteAtomic validates before replacing a single complete snapshot. On failure,
// the previous file is retained and the temporary file is removed.
func WriteAtomic(path string, data []byte) error {
	next, err := catalog.Load(data)
	if err != nil {
		return err
	}
	if previousData, readErr := os.ReadFile(path); readErr == nil {
		previous, loadErr := catalog.Load(previousData)
		if loadErr != nil {
			return fmt.Errorf("existing snapshot cannot be validated: %w", loadErr)
		}
		a, b := previous.Metadata(), next.Metadata()
		if a.JavaVersion != b.JavaVersion || a.BedrockVersion != b.BedrockVersion || a.DataVersion != b.DataVersion {
			return fmt.Errorf("refusing to replace another version; use a new snapshot path")
		}
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".catalog-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
