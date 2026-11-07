//go:build resourcegen

package javastates

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"neomega-blocks/ir"
	java "neomega-blocks/presets/java/v1"
	blockconverts "neomega-blocks/third_party/gen_resource/block-converts"
)

func TestGenerateJavaStates(t *testing.T) {
	// 把旧状态先解释为转换中心的状态，再迁移到目标全集，避免跨两套语义直接猜测。
	versionData, err := os.ReadFile("tmp/version.json")
	if err != nil {
		t.Fatal(err)
	}
	var version struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(versionData, &version); err != nil {
		t.Fatal(err)
	}
	if version.ID != java.JavaTargetVersion {
		t.Fatalf("报告版本 %s 与 JavaTargetVersion %s 不一致", version.ID, java.JavaTargetVersion)
	}
	data, err := os.ReadFile("tmp/blocks.json")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := ParseReport(data)
	if err != nil {
		t.Fatal(err)
	}
	compatibility := make(map[ir.IRBlock]ir.IRBlock)
	migration := make(map[ir.IRBlock]ir.IRBlock)
	for _, source := range blockconverts.ReadJavaStates("../block-converts/tmp/hub_java_states.list") {
		target, err := registry.Migrate(source)
		if err != nil {
			t.Fatal(err)
		}
		if source != target {
			migration[source] = target
			compatibility[source] = target
		}
	}
	for source, hub := range blockconverts.ReadJavaMappings("../block-converts/tmp/hub_java_compatibility.list") {
		if registry.Known[source] {
			continue
		}
		target, err := registry.Migrate(hub)
		if err != nil {
			t.Fatal(err)
		}
		compatibility[source] = target
	}
	var states, aliases, changes strings.Builder
	for _, state := range registry.States {
		fmt.Fprintln(&states, state)
	}
	for _, source := range slices.Sorted(maps.Keys(compatibility)) {
		fmt.Fprintf(&aliases, "%s#%s\n", source, compatibility[source])
	}
	for _, source := range slices.Sorted(maps.Keys(migration)) {
		fmt.Fprintf(&changes, "%s#%s\n", source, migration[source])
	}
	for path, content := range map[string]string{
		"../../../presets/java/v1/java_states.list":        states.String(),
		"../../../presets/java/v1/java_block_schema.list":  strings.Join(registry.Schema, "\n") + "\n",
		"../../../presets/java/v1/java_compatibility.list": aliases.String(),
		"tmp/hub_to_target.mapping":                        changes.String(),
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf(
		"Java %s：%d 种方块，%d 个单层状态，%d 条兼容映射，%d 条中心迁移",
		version.ID,
		len(registry.Schema),
		len(registry.States),
		len(compatibility),
		len(migration),
	)
}
