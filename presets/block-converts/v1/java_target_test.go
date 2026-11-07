package v1_test

import (
	"testing"

	"neomega-blocks/ir"
	converts "neomega-blocks/presets/block-converts/v1"
	java "neomega-blocks/presets/java/v1"
)

func TestPaletteJavaOutputNormalizesToTarget(t *testing.T) {
	// 存档写出保持中心版本，浏览和修复则必须能解释每一个写出的状态。
	for _, source := range converts.BedrockBlocks() {
		output, ok := converts.ConvertToJava(source)
		if !ok {
			t.Fatalf("palette 状态无法转出 Java：%s", source)
		}
		if _, ok := java.NormalizeSingleLayer(output); !ok {
			t.Fatalf("目标版本无法解释中心输出：%s → %s", source, output)
		}
	}
}

func TestRenamedJavaInput(t *testing.T) {
	// 当前目标的改名状态与中心状态必须转入同一方块。
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
	for active, state := range map[string]string{"true": "awake", "false": "uprooted"} {
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

func TestIsJavaBlockUsesConversionDefaults(t *testing.T) {
	// 历史名称仍属于转换中心；只在目标状态全集出现的方块不扩大存档转换范围。
	for _, name := range []string{"chain", "minecraft:chain", "grass", "minecraft:stone"} {
		if !converts.IsJavaBlock(name) {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"minecraft:copper_chest", "example:unknown"} {
		if converts.IsJavaBlock(name) {
			t.Fatal(name)
		}
	}
}
