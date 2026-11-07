package v1_test

import (
	"neomega-blocks/ir"
	blocks "neomega-blocks/presets/block-converts/v1"
	"testing"
)

func TestMissingPropertiesUseCanonicalDefaults(t *testing.T) {
	cases := []struct {
		name    string
		input   ir.IRBlock
		canon   func(ir.IRBlock) (ir.IRBlock, bool)
		convert func(ir.IRBlock) (ir.IRBlock, bool)
	}{
		{"java", `minecraft:jungle_fence_gate ["open"="true"]`, blocks.JavaCanon, blocks.ConvertFromJava},
		{"bedrock", `minecraft:dropper ["facing_direction"=0]`, blocks.BedrockCanon, blocks.ConvertFromBedrock},
		{"to java", `minecraft:orange_glazed_terracotta []`, blocks.BedrockCanon, blocks.ConvertToJava},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			full, ok := tc.canon(tc.input)
			if !ok {
				t.Fatal("missing defaults")
			}
			got, ok := tc.convert(tc.input)
			if !ok {
				t.Fatal("missing-property input rejected")
			}
			want, ok := tc.convert(full)
			if !ok || got != want {
				t.Fatalf("partial: %s; complete: %s (%v)", got, want, ok)
			}
		})
	}
}

func TestWallPostRoundTrip(t *testing.T) {
	for _, post := range []string{"true", "false"} {
		source := ir.IRBlock(`minecraft:cobblestone_wall ["wall_post_bit"=` + post + `]`)
		java, ok := blocks.ConvertToJava(source)
		if !ok {
			t.Fatal("wall conversion failed")
		}
		if got := java.States().GetDefault("up", nil); got != post {
			t.Fatalf("got up=%v, want %s", got, post)
		}
		back, ok := blocks.ConvertFromJava(java)
		if !ok {
			t.Fatal("wall reverse conversion failed")
		}
		want, _ := blocks.BedrockCanon(source)
		if back != want {
			t.Fatalf("got %s, want %s", back, want)
		}
	}
}
