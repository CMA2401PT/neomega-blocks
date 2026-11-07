// Package catalog loads versioned Java state registries and Geyser block mappings.
// It does not translate world files, block entities, models, or legacy versions.
package catalog

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"maps"
	"slices"
	"strings"

	"neomega-blocks/ir"
)

const FormatVersion = 1

// Source pins the exact input used by the offline generator.
type Source struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Revision string `json:"revision"`
	SHA256   string `json:"sha256"`
}

type Metadata struct {
	FormatVersion  int      `json:"formatVersion"`
	JavaVersion    string   `json:"javaVersion"`
	BedrockVersion string   `json:"bedrockVersion"`
	DataVersion    int      `json:"dataVersion"`
	Sources        []Source `json:"sources"`
}

// Definition contains Java string-valued properties, in upstream domain order.
type Definition struct {
	Name           string              `json:"name"`
	Properties     map[string][]string `json:"properties"`
	DefaultStateID int                 `json:"defaultStateId"`
}

// State.ID is a Java registry ID valid only within this snapshot.
// Java retains waterlogged. Bedrock is the Geyser primary-layer mapping,
// preserving NBT byte/int/string distinctions in its IR text.
type State struct {
	ID      int        `json:"id"`
	Java    ir.IRBlock `json:"java"`
	Bedrock ir.IRBlock `json:"bedrock"`
}

// Snapshot is the portable on-disk format. It does not contain native Go memory.
type Snapshot struct {
	Metadata Metadata     `json:"metadata"`
	Blocks   []Definition `json:"blocks"`
	States   []State      `json:"states"`
}

// Catalog owns its data; its public accessors return values or copies.
type Catalog struct {
	data      Snapshot
	blocks    map[string]int
	byJava    map[ir.IRBlock]int
	byHash    map[uint32]int
	byBedrock map[ir.IRBlock][]int
}

// Load reads and validates a gzip-compressed JSON snapshot.
func Load(data []byte) (*Catalog, error) {
	z, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("catalog gzip: %w", err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		return nil, fmt.Errorf("catalog gzip payload: %w", err)
	}
	var snapshot Snapshot
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("catalog JSON: %w", err)
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("trailing catalog JSON")
	}
	return New(snapshot)
}

// New validates and takes a copy of a snapshot. It rejects missing/duplicate IDs,
// invalid domains, incomplete state universes, and ambiguous 32-bit hashes.
func New(input Snapshot) (*Catalog, error) {
	if input.Metadata.FormatVersion != FormatVersion {
		return nil, fmt.Errorf("unsupported catalog format %d", input.Metadata.FormatVersion)
	}
	if input.Metadata.JavaVersion == "" || input.Metadata.BedrockVersion == "" || input.Metadata.DataVersion <= 0 {
		return nil, fmt.Errorf("missing catalog version metadata")
	}
	if len(input.Blocks) == 0 || len(input.States) == 0 {
		return nil, fmt.Errorf("empty catalog")
	}
	c := &Catalog{
		data:   Snapshot{Metadata: input.Metadata, States: slices.Clone(input.States)},
		blocks: make(map[string]int), byJava: make(map[ir.IRBlock]int),
		byHash: make(map[uint32]int), byBedrock: make(map[ir.IRBlock][]int),
	}
	c.data.Metadata.Sources = slices.Clone(input.Metadata.Sources)
	for i, block := range input.Blocks {
		if block.Name == "" || !strings.Contains(block.Name, ":") || ir.IRBlock(block.Name+" []").Tidy().NameWithNameSpace() != block.Name {
			return nil, fmt.Errorf("invalid block name %q", block.Name)
		}
		if _, exists := c.blocks[block.Name]; exists {
			return nil, fmt.Errorf("duplicate block %s", block.Name)
		}
		for key, values := range block.Properties {
			if key == "" || len(values) == 0 || len(values) != len(slices.Compact(slices.Sorted(slices.Values(values)))) {
				return nil, fmt.Errorf("invalid property domain %s.%s", block.Name, key)
			}
		}
		c.blocks[block.Name] = i
		c.data.Blocks = append(c.data.Blocks, copyDefinition(block))
	}
	counts := make(map[string]int)
	for i, state := range c.data.States {
		if state.ID != i {
			return nil, fmt.Errorf("state index %d has ID %d", i, state.ID)
		}
		if !state.Java.IsValid() || !state.Java.IsTidy() || !state.Bedrock.IsValid() || !state.Bedrock.IsTidy() {
			return nil, fmt.Errorf("state %d has invalid or untidy IR", i)
		}
		bi, exists := c.blocks[state.Java.NameWithNameSpace()]
		if !exists {
			return nil, fmt.Errorf("state %d has undeclared block", i)
		}
		if err := validateProperties(state.Java, c.data.Blocks[bi], false); err != nil {
			return nil, fmt.Errorf("state %d: %w", i, err)
		}
		if _, exists := c.byJava[state.Java]; exists {
			return nil, fmt.Errorf("duplicate Java state %s", state.Java)
		}
		hash := state.Java.Hash()
		if other, exists := c.byHash[hash]; exists {
			return nil, fmt.Errorf("Java hash collision %08x: %s / %s", hash, c.data.States[other].Java, state.Java)
		}
		c.byJava[state.Java], c.byHash[hash] = i, i
		c.byBedrock[state.Bedrock] = append(c.byBedrock[state.Bedrock], i)
		counts[state.Java.NameWithNameSpace()]++
	}
	for _, block := range c.data.Blocks {
		id := block.DefaultStateID
		if id < 0 || id >= len(c.data.States) || c.data.States[id].Java.NameWithNameSpace() != block.Name {
			return nil, fmt.Errorf("invalid default for %s", block.Name)
		}
		expected := 1
		for _, values := range block.Properties {
			if expected > len(c.data.States)/len(values) {
				return nil, fmt.Errorf("oversized property product for %s", block.Name)
			}
			expected *= len(values)
		}
		if counts[block.Name] != expected {
			return nil, fmt.Errorf("incomplete states for %s: %d, expected %d", block.Name, counts[block.Name], expected)
		}
	}
	return c, nil
}

func copyDefinition(block Definition) Definition {
	block.Properties = maps.Clone(block.Properties)
	for key, values := range block.Properties {
		block.Properties[key] = slices.Clone(values)
	}
	return block
}

func validateProperties(block ir.IRBlock, definition Definition, allowMissing bool) error {
	seen := make(map[string]bool)
	for key, value := range block.States().KeyValues {
		if seen[key] {
			return fmt.Errorf("duplicate property %q", key)
		}
		seen[key] = true
		domain, known := definition.Properties[key]
		text, isString := value.(string)
		if !known || !isString || !slices.Contains(domain, text) {
			return fmt.Errorf("invalid Java property %s.%s=%v", definition.Name, key, value)
		}
	}
	if !allowMissing && len(seen) != len(definition.Properties) {
		return fmt.Errorf("incomplete Java properties for %s", definition.Name)
	}
	return nil
}

func (c *Catalog) Metadata() Metadata {
	m := c.data.Metadata
	m.Sources = slices.Clone(m.Sources)
	return m
}

func (c *Catalog) BlockCount() int { return len(c.data.Blocks) }
func (c *Catalog) StateCount() int { return len(c.data.States) }

func (c *Catalog) Blocks() iter.Seq[Definition] {
	return func(yield func(Definition) bool) {
		for _, block := range c.data.Blocks {
			if !yield(copyDefinition(block)) {
				return
			}
		}
	}
}

// States includes real waterlogged states, never fabricated variants for stone etc.
func (c *Catalog) States() iter.Seq[State] { return slices.Values(c.data.States) }

func (c *Catalog) Definition(name string) (Definition, bool) {
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	i, ok := c.blocks[name]
	if !ok {
		return Definition{}, false
	}
	return copyDefinition(c.data.Blocks[i]), true
}

func (c *Catalog) ByID(id int) (State, bool) {
	if id < 0 || id >= len(c.data.States) {
		return State{}, false
	}
	return c.data.States[id], true
}

// Lookup requires a full, valid Java state. It never silently repairs invalid input.
func (c *Catalog) Lookup(block ir.IRBlock) (State, bool) {
	if !block.IsValid() {
		return State{}, false
	}
	i, ok := c.byJava[block.Tidy()]
	if !ok {
		return State{}, false
	}
	return c.data.States[i], true
}

func (c *Catalog) ByHash(hash uint32) (State, bool) {
	i, ok := c.byHash[hash]
	if !ok {
		return State{}, false
	}
	return c.data.States[i], true
}

// Normalize fills only missing properties from the official default state.
// Unknown names, extra keys, invalid types and out-of-domain values are errors.
func (c *Catalog) Normalize(block ir.IRBlock) (ir.IRBlock, error) {
	if !block.IsValid() {
		return "", fmt.Errorf("invalid IR syntax: %q", block)
	}
	name := block.Tidy().NameWithNameSpace()
	bi, ok := c.blocks[name]
	if !ok {
		return "", fmt.Errorf("unknown Java block %s", name)
	}
	definition := c.data.Blocks[bi]
	if err := validateProperties(block, definition, true); err != nil {
		return "", err
	}
	defaults := c.data.States[definition.DefaultStateID].Java.States()
	for key, value := range block.States().KeyValues {
		defaults = defaults.Set(key, value)
	}
	result := ir.IRBlock(name + " " + string(defaults)).Tidy()
	if _, exists := c.byJava[result]; !exists {
		return "", fmt.Errorf("state not in registry: %s", result)
	}
	return result, nil
}

// Mapping exposes Geyser's primary block and Java's explicit waterlogged flag.
// This is not a complete world conversion (fluids, entities and context need callers).
type Mapping struct {
	Block       ir.IRBlock
	Waterlogged bool
}

func (c *Catalog) ToBedrock(block ir.IRBlock) (Mapping, error) {
	full, err := c.Normalize(block)
	if err != nil {
		return Mapping{}, err
	}
	state := c.data.States[c.byJava[full]]
	return Mapping{Block: state.Bedrock, Waterlogged: full.States().CheckKey("waterlogged", "true")}, nil
}

// JavaCandidates returns every full Java state mapped to this Bedrock primary
// block, ordered by Java registry ID. Multiple candidates are intentionally retained.
func (c *Catalog) JavaCandidates(block ir.IRBlock) []State {
	if !block.IsValid() {
		return nil
	}
	ids := c.byBedrock[block.Tidy()]
	result := make([]State, 0, len(ids))
	for _, id := range ids {
		result = append(result, c.data.States[id])
	}
	return result
}
