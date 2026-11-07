// 枚举 Java 完整状态，只读分析 Java → NEMC → Java 的差异。
package main

import (
	"cmp"
	"fmt"
	"maps"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"neomega-blocks/ir"
	converts "neomega-blocks/presets/block-converts/v1"
	blockconverts "neomega-blocks/third_party/gen_resource/block-converts"
)

type propertyStats struct {
	changed int
	missing int
	added   int
}

func (p propertyStats) total() int {
	return p.changed + p.missing + p.added
}

type blockStats struct {
	name       string
	total      int
	bad        int
	fromFailed int
	toFailed   int
	renamed    map[string]int
	properties map[string]*propertyStats
}

type propertySummary struct {
	name   string
	blocks int
	propertyStats
}

func main() {
	byName := make(map[string]*blockStats)
	total, bad, fromFailed, toFailed, renamed, propertyChanged := 0, 0, 0, 0, 0, 0
	waterloggedStates := 0
	_, source, _, _ := runtime.Caller(0)
	statesPath := filepath.Join(filepath.Dir(source), "../block-converts/tmp/hub_java_states.list")
	for _, input := range blockconverts.ReadJavaStates(statesPath) {
		input = input.Tidy()
		if _, exists := input.States().Get("waterlogged"); exists {
			waterloggedStates++
		}
		name := input.NameWithNameSpace()
		b := byName[name]
		if b == nil {
			b = &blockStats{
				name:       name,
				renamed:    make(map[string]int),
				properties: make(map[string]*propertyStats),
			}
			byName[name] = b
		}
		total++
		b.total++
		bedrock, ok := converts.ConvertFromJava(input)
		if !ok {
			b.bad++
			bad++
			b.fromFailed++
			fromFailed++
			continue
		}
		output, ok := converts.ConvertToJava(bedrock)
		if !ok {
			b.bad++
			bad++
			b.toFailed++
			toFailed++
			continue
		}
		output = output.Tidy()
		if input == output {
			continue
		}
		b.bad++
		bad++
		if name != output.NameWithNameSpace() {
			b.renamed[output.NameWithNameSpace()]++
			renamed++
		}
		if input.States() != output.States() {
			propertyChanged++
		}
		recordProperties(b, input.States(), output.States())
	}

	blocks := make([]*blockStats, 0)
	properties := make(map[string]*propertySummary)
	for _, b := range byName {
		if b.bad == 0 {
			continue
		}
		blocks = append(blocks, b)
		for name, p := range b.properties {
			s := properties[name]
			if s == nil {
				s = &propertySummary{name: name}
				properties[name] = s
			}
			s.blocks++
			s.changed += p.changed
			s.missing += p.missing
			s.added += p.added
		}
	}
	// 优先展示受影响状态最多的方块，同数时按名称排序。
	slices.SortFunc(blocks, func(a, b *blockStats) int {
		return cmp.Or(cmp.Compare(b.bad, a.bad), cmp.Compare(a.name, b.name))
	})
	summaries := slices.Collect(maps.Values(properties))
	slices.SortFunc(summaries, func(a, b *propertySummary) int {
		return cmp.Or(
			cmp.Compare(b.blocks, a.blocks),
			cmp.Compare(b.total(), a.total()),
			cmp.Compare(a.name, b.name),
		)
	})

	fmt.Println("# Java → NEMC → Java 往返差异报告")
	fmt.Println()
	fmt.Println("本报告枚举转换中心的 hub_java_states.list，依次调用 ConvertFromJava 和 ConvertToJava，将原状态和返回状态分别 Tidy 后比较。它反映当前转换链的信息损失或不一致，不能单独证明 Bedrock 本身无法表达，也不能保证所有差异都能由 fixer 恢复。")
	fmt.Println()
	fmt.Printf("输入：block-converts/tmp/hub_java_states.list，Java %s；转换：presets/block-converts/v1。\n", converts.JavaTargetVersion)
	fmt.Println()
	fmt.Println("在仓库根目录重新生成：")
	fmt.Println()
	fmt.Println("```sh")
	fmt.Println("mkdir -p third_party/gen_resource/java-roundtrip/tmp && go run ./third_party/gen_resource/java-roundtrip > third_party/gen_resource/java-roundtrip/tmp/REPORT.md")
	fmt.Println("```")
	fmt.Println()
	fmt.Println("## 总览")
	fmt.Println()
	fmt.Printf("- 枚举：%d 个状态，%d 种方块。\n", total, len(byName))
	fmt.Printf("- 往返不一致或失败：%d 个状态（%.2f%%），涉及 %d 种方块；一致 %d 个状态。\n", bad, 100*float64(bad)/float64(total), len(blocks), total-bad)
	fmt.Printf("- Java → NEMC 失败：%d；NEMC → Java 失败：%d。\n", fromFailed, toFailed)
	fmt.Printf("- 两步成功但不一致：%d；其中方块名变化 %d，属性变化 %d（两者可能重叠）。\n", bad-fromFailed-toFailed, renamed, propertyChanged)
	fmt.Printf("- 输入包含 waterlogged 属性的状态：%d。结论只覆盖输入实际枚举的属性，未枚举的属性不能据此判断是否保留。\n", waterloggedStates)
	printFamilies(byName)
	fmt.Println()
	fmt.Println("## 属性影响范围")
	fmt.Println()
	fmt.Println("按受影响方块种数降序排列。改值、缺失、新增分别表示同名属性值不同、返回状态没有该属性、返回状态多出该属性。仅统计两步成功的状态；转换失败不推断具体丢失属性。各属性计数可能来自同一状态，不能相加当作状态总数。")
	fmt.Println()
	fmt.Println("| 属性 | 方块种数 | 改值状态数 | 缺失状态数 | 新增状态数 |")
	fmt.Println("| --- | ---: | ---: | ---: | ---: |")
	for _, p := range summaries {
		fmt.Printf("| `%s` | %d | %d | %d | %d |\n", p.name, p.blocks, p.changed, p.missing, p.added)
	}
	fmt.Println()
	fmt.Println("## 按方块列出全部差异")
	fmt.Println()
	fmt.Println("按不一致状态数降序排列。方块名省略 minecraft: 前缀；属性后的数字为受影响状态数，括号标明差异类型。失败列依次为 Java → NEMC / NEMC → Java。方块改名时，属性差异仍按原名归组，需结合返回方块名理解。")
	fmt.Println()
	fmt.Println("| 方块 | 不一致 / 总状态 | 变化或丢失的属性 | 返回新增属性 | 返回方块名（状态数） | 失败 |")
	fmt.Println("| --- | ---: | --- | --- | --- | ---: |")
	for _, b := range blocks {
		var losses, additions, names []string
		for _, name := range slices.Sorted(maps.Keys(b.properties)) {
			p := b.properties[name]
			if p.changed > 0 {
				losses = append(losses, fmt.Sprintf("`%s`:%d（改值）", name, p.changed))
			}
			if p.missing > 0 {
				losses = append(losses, fmt.Sprintf("`%s`:%d（缺失）", name, p.missing))
			}
			if p.added > 0 {
				additions = append(additions, fmt.Sprintf("`%s`:%d", name, p.added))
			}
		}
		for _, name := range slices.Sorted(maps.Keys(b.renamed)) {
			names = append(names, fmt.Sprintf("`%s`:%d", strings.TrimPrefix(name, "minecraft:"), b.renamed[name]))
		}
		fmt.Printf("| `%s` | %d / %d | %s | %s | %s | %d / %d |\n",
			strings.TrimPrefix(b.name, "minecraft:"), b.bad, b.total,
			cell(losses), cell(additions), cell(names), b.fromFailed, b.toFailed,
		)
	}
}

func printFamilies(blocks map[string]*blockStats) {
	fmt.Println()
	fmt.Println("## 重点类别核对")
	fmt.Println()
	fmt.Println("包含往返完全一致的类别；属性列是该类别所有差异属性的并集，具体计数与改名目标见逐方块表。")
	fmt.Println()
	fmt.Println("| 类别（名称匹配） | 受影响 / 总方块种数 | 不一致 / 总状态 | 差异属性 | 方块名变化状态数 |")
	fmt.Println("| --- | ---: | ---: | --- | ---: |")
	for _, family := range []struct {
		label  string
		suffix string
	}{
		{"栅栏", "_fence"},
		{"墙", "_wall"},
		{"门", "_door"},
		{"床", "_bed"},
		{"楼梯", "_stairs"},
		{"铁轨", "rail"},
		{"玻璃板", "glass_pane"},
		{"铁栏杆", "iron_bars"},
		{"栅栏门", "_fence_gate"},
		{"活板门", "_trapdoor"},
	} {
		totalBlocks, badBlocks, total, bad, renamed := 0, 0, 0, 0, 0
		properties := make(map[string]bool)
		for _, b := range blocks {
			if !strings.HasSuffix(b.name, family.suffix) {
				continue
			}
			totalBlocks++
			total += b.total
			bad += b.bad
			if b.bad > 0 {
				badBlocks++
			}
			for name := range b.properties {
				properties[name] = true
			}
			for _, count := range b.renamed {
				renamed += count
			}
		}
		fmt.Printf("| %s（`*%s`） | %d / %d | %d / %d | %s | %d |\n",
			family.label, family.suffix, badBlocks, totalBlocks, bad, total,
			cell(slices.Sorted(maps.Keys(properties))), renamed,
		)
	}
}

func recordProperties(b *blockStats, input, output ir.IRStates) {
	before := maps.Collect(input.KeyValueStrings)
	after := maps.Collect(output.KeyValueStrings)
	for key, value := range before {
		other, exists := after[key]
		if exists && value == other {
			continue
		}
		p := property(b, key)
		if exists {
			p.changed++
		} else {
			p.missing++
		}
	}
	for key := range after {
		if _, exists := before[key]; !exists {
			property(b, key).added++
		}
	}
}

func property(b *blockStats, key string) *propertyStats {
	p := b.properties[key]
	if p == nil {
		p = &propertyStats{}
		b.properties[key] = p
	}
	return p
}

func cell(items []string) string {
	if len(items) == 0 {
		return "—"
	}
	return strings.Join(items, "; ")
}
