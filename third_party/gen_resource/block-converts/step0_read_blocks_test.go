//go:build resourcegen

package blockconverts

import (
	"os"
	"testing"

	blocks_name "neomega-blocks/storage/blocks-name"
)

func TestStep0ReadBedrockStates(t *testing.T) {
	// 状态全集取自目标 specification，网络哈希由保留 NBT 类型的 IR 计算。
	blocks := ReadStates("tmp/bedrock_states.list")
	storage := blocks_name.NewBlockNameStorage[uint16]()
	for _, block := range blocks {
		storage.Set(block)
	}
	if err := os.WriteFile("../../../presets/block-converts/v1/0_bedrock_blocks.bin", storage.Dump(), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("Bedrock states=%d", len(blocks))
}
