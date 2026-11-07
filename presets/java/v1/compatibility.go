package v1

import (
	_ "embed"
	"iter"
	"strings"
	"sync"

	"neomega-blocks/ir"
)

// Sparse, complete single-layer states: source#target. Target identities are
// supplied by JavaFullStates. Generation visits source versions newest first.
//
// 由 third_party/gen_resource/block-converts/gen_java_compat.py 直接生成到目标版本。
//
//go:embed java_compatibility.list
var dataCompatibility string

type StateMapping struct {
	Source ir.IRBlock
	Target ir.IRBlock
}

var loadCompatibility = sync.OnceValue(func() []StateMapping {
	var result []StateMapping
	for line := range strings.Lines(dataCompatibility) {
		source, target, _ := strings.Cut(strings.TrimSpace(line), "#")
		result = append(result, StateMapping{Source: ir.IRBlock(source).Tidy(), Target: ir.IRBlock(target).Tidy()})
	}
	return result
})

// JavaCompatibility enumerates additional single-layer input states and their
// target-version interpretation. Shared source texts use the newest definition.
func JavaCompatibility() iter.Seq[StateMapping] {
	return func(yield func(StateMapping) bool) {
		for _, mapping := range loadCompatibility() {
			if !yield(mapping) {
				return
			}
		}
	}
}

// JavaCompatibleStates includes the target universe and the accepted legacy
// inputs. Like JavaFullStates, it omits waterlogged; callers expand that separately.
func JavaCompatibleStates() iter.Seq[ir.IRBlock] {
	return func(yield func(ir.IRBlock) bool) {
		for block := range JavaFullStates() {
			if !yield(block) {
				return
			}
		}
		for mapping := range JavaCompatibility() {
			if !yield(mapping.Source) {
				return
			}
		}
	}
}

var compatibilityTargets = sync.OnceValue(func() map[uint32]ir.IRBlock {
	result := make(map[uint32]ir.IRBlock)
	for block := range JavaFullStates() {
		result[block.Hash()] = block
	}
	for mapping := range JavaCompatibility() {
		result[mapping.Source.Hash()] = mapping.Target
	}
	return result
})

// NormalizeSingleLayer interprets a complete state without a source version.
// It ignores waterlogged, matching geometry's single-layer contract.
func NormalizeSingleLayer(block ir.IRBlock) (ir.IRBlock, bool) {
	block = ir.IRBlock(block.NameWithNameSpace() + " " + string(block.States().Delete("waterlogged"))).Tidy()
	result, ok := compatibilityTargets()[block.Hash()]
	return result, ok
}

var compatibilityByHash = sync.OnceValue(func() map[uint32]ir.IRBlock {
	result := make(map[uint32]ir.IRBlock)
	for mapping := range JavaCompatibility() {
		block := mapping.Source
		result[block.Hash()] = block
		for _, value := range []string{"true", "false"} {
			variant := ir.IRBlock(block.NameWithNameSpace() + " " + string(block.States().Set("waterlogged", value))).Tidy()
			result[variant.Hash()] = variant
		}
	}
	return result
})
