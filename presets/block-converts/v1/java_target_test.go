package v1_test

import (
	"testing"

	"neomega-blocks/ir"
	converts "neomega-blocks/presets/block-converts/v1"
	java "neomega-blocks/presets/java/v1"
)

func TestBedrockJavaOutputsBelongToTarget(t *testing.T) {
	// 转出结果必须直接属于目标单层全集，不能靠兼容映射再次解释。
	native := map[ir.IRBlock]bool{}
	for state := range java.JavaFullStates() {
		native[state] = true
	}
	for _, source := range converts.BedrockBlocks() {
		output, ok := converts.ConvertToJava(source)
		if !ok {
			t.Fatalf("palette 状态无法转出 Java：%s", source)
		}
		if !native[output.Tidy()] {
			t.Fatalf("目标版本无法解释中心输出：%s → %s", source, output)
		}
	}
}

func TestRenamedJavaInput(t *testing.T) {
	// 历史名称与当前名称应转入同一方块。
	for _, axis := range []string{"x", "y", "z"} {
		legacy := ir.IRBlock(`minecraft:chain ["axis"="` + axis + `"]`)
		current := ir.IRBlock(`minecraft:iron_chain ["axis"="` + axis + `"]`)
		want, ok := converts.ConvertFromJava(legacy)
		if !ok {
			t.Fatal(legacy)
		}
		if got, ok := converts.ConvertFromJava(current); !ok || got != want {
			t.Fatalf("%s → %s (%v)，期望 %s", current, got, ok, want)
		}
	}
}

func TestRewrittenJavaInput(t *testing.T) {
	// active 的两个含义必须通过生成表转入，避免 JavaCanon 删除旧属性后混为同一状态。
	for active, state := range map[string]string{"true": "awake", "false": "dormant"} {
		legacy := ir.IRBlock(`minecraft:creaking_heart ["active"="` + active + `","axis"="y","natural"="true"]`)
		current := ir.IRBlock(`minecraft:creaking_heart ["axis"="y","creaking_heart_state"="` + state + `","natural"="true"]`)
		want, ok := converts.ConvertFromJava(legacy)
		if !ok {
			t.Fatal(legacy)
		}
		if got, ok := converts.ConvertFromJava(current); !ok || got != want {
			t.Fatalf("%s → %s (%v)，期望 %s", current, got, ok, want)
		}
	}
}

func TestLegacyBlockstateInput(t *testing.T) {
	// 数字 ID 时代的 schema 也按 blockstate 解释，才能保留颜色、朝向与供电属性。
	for source, want := range map[ir.IRBlock]ir.IRBlock{
		`minecraft:carpet ["color"="black"]`:                    `minecraft:black_carpet []`,
		`minecraft:observer ["facing"="east","powered"="true"]`: `minecraft:observer ["minecraft:facing_direction"="east","powered_bit"=true]`,
	} {
		got, ok := converts.ConvertFromJava(source)
		if !ok || got != want {
			t.Fatalf("%s → %s (%v)，期望 %s", source, got, ok, want)
		}
	}
}
