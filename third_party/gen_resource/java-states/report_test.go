package javastates

import (
	"slices"
	"strings"
	"testing"

	"neomega-blocks/ir"
)

func TestParseReport(t *testing.T) {
	// 报告的对象顺序不等于注册顺序；含水变体应合并，属性域仍保持上游次序。
	registry, err := ParseReport([]byte(`{
		"minecraft:slab": {
			"properties": {"waterlogged":["true","false"],"type":["top","bottom"]},
			"states": [
				{"id":3,"properties":{"waterlogged":"true","type":"top"}},
				{"id":2,"properties":{"waterlogged":"false","type":"top"}},
				{"id":4,"default":true,"properties":{"waterlogged":"false","type":"bottom"}}
			]
		},
		"minecraft:air": {"states":[{"id":0,"default":true}]}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	want := []ir.IRBlock{"minecraft:air []", `minecraft:slab ["type"="top"]`, `minecraft:slab ["type"="bottom"]`}
	if !slices.Equal(registry.States, want) {
		t.Fatalf("状态 = %v，期望 %v", registry.States, want)
	}
	if got := strings.Join(registry.Schema, "\n"); got != "minecraft:air\nminecraft:slab\ttype=\"bottom\"|\"top\",\"bottom\"" {
		t.Fatal(got)
	}
}

func TestMigrate(t *testing.T) {
	// 改名与属性替换必须先于补默认值，否则旧状态的含义会被默认状态吞掉。
	registry, err := ParseReport([]byte(`{
		"minecraft:iron_chain": {
			"properties":{"axis":["x","y"]},
			"states":[
				{"id":0,"default":true,"properties":{"axis":"y"}},
				{"id":1,"properties":{"axis":"x"}}
			]
		},
		"minecraft:creaking_heart": {
			"properties":{"creaking_heart_state":["uprooted","awake"]},
			"states":[
				{"id":2,"default":true,"properties":{"creaking_heart_state":"uprooted"}},
				{"id":3,"properties":{"creaking_heart_state":"awake"}}
			]
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	for source, want := range map[ir.IRBlock]ir.IRBlock{
		`minecraft:chain ["axis"="x","waterlogged"="true"]`: `minecraft:iron_chain ["axis"="x"]`,
		"minecraft:chain []":                          `minecraft:iron_chain ["axis"="y"]`,
		`minecraft:creaking_heart ["active"="true"]`:  `minecraft:creaking_heart ["creaking_heart_state"="awake"]`,
		`minecraft:creaking_heart ["active"="false"]`: `minecraft:creaking_heart ["creaking_heart_state"="uprooted"]`,
	} {
		if got, err := registry.Migrate(source); err != nil || got != want {
			t.Fatalf("%s → %s (%v)，期望 %s", source, got, err, want)
		}
	}
	source := ir.IRBlock(`minecraft:chain ["axis"="z"]`)
	if _, err := registry.Migrate(source); err == nil || !strings.Contains(err.Error(), string(source)) {
		t.Fatalf("缺失迁移规则应报告原状态：%v", err)
	}
}
