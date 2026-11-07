package catalog_test

import (
	"testing"

	"neomega-blocks/catalog"
	"neomega-blocks/ir"
)

func fixture() catalog.Snapshot {
	return catalog.Snapshot{
		Metadata: catalog.Metadata{FormatVersion: 1, JavaVersion: "test", BedrockVersion: "test", DataVersion: 1},
		Blocks:   []catalog.Definition{{Name: "minecraft:test", Properties: map[string][]string{"axis": {"x", "y"}}, DefaultStateID: 1}},
		States: []catalog.State{
			{ID: 0, Java: `minecraft:test ["axis"="x"]`, Bedrock: `minecraft:test ["direction"=0]`},
			{ID: 1, Java: `minecraft:test ["axis"="y"]`, Bedrock: `minecraft:test ["direction"=1]`},
		},
	}
}

func TestRejectCorruptRegistry(t *testing.T) {
	cases := map[string]func(*catalog.Snapshot){
		"future format":       func(s *catalog.Snapshot) { s.Metadata.FormatVersion++ },
		"missing version":     func(s *catalog.Snapshot) { s.Metadata.JavaVersion = "" },
		"duplicate ID":        func(s *catalog.Snapshot) { s.States[1].ID = 0 },
		"duplicate text":      func(s *catalog.Snapshot) { s.States[1].Java = s.States[0].Java },
		"invalid value":       func(s *catalog.Snapshot) { s.States[0].Java = `minecraft:test ["axis"="z"]` },
		"invalid type":        func(s *catalog.Snapshot) { s.States[0].Java = `minecraft:test ["axis"=1]` },
		"missing property":    func(s *catalog.Snapshot) { s.States[0].Java = `minecraft:test []` },
		"extra property":      func(s *catalog.Snapshot) { s.States[0].Java = `minecraft:test ["axis"="x","waterlogged"="false"]` },
		"duplicate block":     func(s *catalog.Snapshot) { s.Blocks = append(s.Blocks, s.Blocks[0]) },
		"invalid default":     func(s *catalog.Snapshot) { s.Blocks[0].DefaultStateID = 99 },
		"incomplete universe": func(s *catalog.Snapshot) { s.States = s.States[:1]; s.Blocks[0].DefaultStateID = 0 },
		"empty domain":        func(s *catalog.Snapshot) { s.Blocks[0].Properties["axis"] = nil },
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			snapshot := fixture()
			corrupt(&snapshot)
			if _, err := catalog.New(snapshot); err == nil {
				t.Fatal("corrupt snapshot accepted")
			}
		})
	}
}

func TestStrictNormalizationAndIsolation(t *testing.T) {
	snapshot := fixture()
	c, err := catalog.New(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.States[0].Java = "damaged"
	snapshot.Blocks[0].Properties["axis"][0] = "damaged"
	got, err := c.Normalize("test []")
	if err != nil || got != `minecraft:test ["axis"="y"]` {
		t.Fatalf("default: %q %v", got, err)
	}
	for _, input := range []ir.IRBlock{`test ["axis"="z"]`, `test ["axis"=1]`, `test ["axis"="x","bogus"="true"]`, `test [broken`, `unknown []`, `test ["axis"="x","axis"="y"]`} {
		if _, err := c.Normalize(input); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
	definition, _ := c.Definition("test")
	definition.Properties["axis"][0] = "damaged"
	definition, _ = c.Definition("test")
	if definition.Properties["axis"][0] != "x" {
		t.Fatal("returned definition mutated catalog")
	}
	if _, ok := c.ByID(-1); ok {
		t.Fatal("negative ID accepted")
	}
	if _, ok := c.Lookup("test [broken"); ok {
		t.Fatal("malformed state silently repaired")
	}
}
