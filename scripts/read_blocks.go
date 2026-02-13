package scripts

import (
	"encoding/json"
	"fmt"
	"os"

	"neomega-blocks/ir"
)

type Block struct {
	BlockID   uint32
	Data      int32
	NameHash  uint64
	NetworkID uint32
	IRBlock   ir.IRBlock
}

type paletteJSON struct {
	Blocks []paletteBlock `json:"blocks"`
}

type paletteBlock struct {
	BlockID   uint32         `json:"block_id"`
	Data      int32          `json:"data"`
	Name      string         `json:"name"`
	NameHash  uint64         `json:"name_hash"`
	NetworkID uint32         `json:"network_id"`
	States    []paletteState `json:"states"`
}

type paletteState struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

func ReadBlocks(path string) []Block {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("open block palette %q: %w", path, err))
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber()

	var palette paletteJSON
	if err := dec.Decode(&palette); err != nil {
		panic(fmt.Errorf("decode block palette %q: %w", path, err))
	}

	out := make([]Block, 0, len(palette.Blocks))
	for _, b := range palette.Blocks {
		states := ir.IRStates("[]")
		for _, s := range b.States {
			switch s.Type {
			case "byte":
				var v int8
				if err := json.Unmarshal(s.Value, &v); err != nil {
					panic(fmt.Errorf("decode state value(byte) for block %q key %q: %w", b.Name, s.Name, err))
				}
				states = states.Set(s.Name, v)
			case "int":
				var v int32
				if err := json.Unmarshal(s.Value, &v); err != nil {
					panic(fmt.Errorf("decode state value(int) for block %q key %q: %w", b.Name, s.Name, err))
				}
				states = states.Set(s.Name, v)
			case "string":
				var v string
				if err := json.Unmarshal(s.Value, &v); err != nil {
					panic(fmt.Errorf("decode state value(string) for block %q key %q: %w", b.Name, s.Name, err))
				}
				states = states.Set(s.Name, v)
			default:
				panic(fmt.Errorf("unsupported state type %q for block %q key %q", s.Type, b.Name, s.Name))
			}
		}
		irB := ir.IRBlock(b.Name + " " + string(states))
		if irB.Hash() != b.NetworkID {
			panic(fmt.Errorf("hash of %v mismatch %v!=%v", irB, irB.Hash(), b.NetworkID))
		}
		out = append(out, Block{
			BlockID:   b.BlockID,
			Data:      b.Data,
			NameHash:  b.NameHash,
			NetworkID: b.NetworkID,
			IRBlock:   irB,
		})
	}
	return out
}
