# 预置资源的离线生成

这里的工具将 PyMCTranslate 的方块定义与跨版本语义生成为 Go 嵌入资源。转换中心连接最新完整的 Java 与国际版 Bedrock，当前目标为 Java 26.3 和 Bedrock 26.60。完整状态、默认属性、属性域和历史兼容映射来自同一数据源，运行时只需查表。

`block-converts` 负责生成，`java-roundtrip` 负责只读分析往返差异。中间文件写入被 Git 忽略的 `tmp/`，最终资源写入 `presets` 的消费包。运行时使用约定见 [`presets`](../../presets/README.md)。

## 数据如何进入转换中心

PyMCTranslate 的 specification 描述方块的属性域和默认值。生成器取每个属性的完整笛卡尔积，得到两侧状态全集；Java 单层全集去掉 `waterlogged` 并去重，含水由独立层表达。Bedrock 状态保留上游 byte、int、string 的 NBT 类型，Go 沿用名称与 states 的 NBT 网络哈希算法。

Universal 是 PyMCTranslate 用来连接各版本语义的中间表示。Java、Bedrock 输入按源版本从高到低枚举，同一个整理后的输入键只采用最高版本解释。缺参形式可以成为映射键，但实际传入 Universal 转换的对象始终先补齐默认属性。schema 查询和双向转换都显式使用 `force_blockstate=True`，使数字 ID 时代的状态也按同一 schema 解释。

Go 将候选输出整理到目标 Bedrock 全集，复用墙和悬挂告示牌的特殊规则，再裁减可由 Canon 补默认值复用的条目。旧属性若会被 Canon 删除且含有不同语义，就保留原输入映射。Java 历史兼容表则从历史完整状态直接转到目标 Java。需要上下文的候选在能形成合法目标状态时采用默认结果，诊断写入 `tmp/java_compatibility_report.json`。

历史 `block_palette.json` 保存名称、带类型的状态、网络哈希与 data value，只生成 legacy 索引和历史状态表。data value 是旧物品 aux 使用的变体编号。原始 NBT palette 作为历史参考保留。`schematic.json` 保存旧 Schematic 的数字 ID、meta 到 Java 状态的映射；直接 Java 输出保留 Java 属性，Bedrock 输出复用 Java 转入表。

## 准备与生成

模块根目录的 `pyproject.toml` 固定 Python 3.11、PyMCTranslate 1.2.53 和 amulet-nbt 2.1.8，`uv.lock` 固定依赖。许可依据和 Required Notice 见 [NOTICE](../pymctranslate-notices/NOTICE.md)。安装 uv 和 Go 后，在 `neomega-blocks` 根目录运行：

```sh
export GOCACHE=/tmp/gocache
uv sync --locked
sh third_party/gen_resource/block-converts/generate.sh
```

完整入口同步 `.venv/`，执行 Python 验证、状态与映射导出，再逐个执行 Go 生成步骤。每个 Go 步骤单独启动测试，使下一个步骤重新编译前一步更新的 embed。写资源入口使用 `resourcegen` 标签，普通测试不会执行它们。发生错误时脚本停止，可以修复后重跑对应步骤。

升级目标时，修改 `pyproject.toml` 的 PyMCTranslate 固定版本，执行 `uv lock` 和 `uv sync --locked`。然后修改 `block-converts/target_versions.py` 中的目标元组，以及 `java/v1.JavaTargetVersion` 和 `block-converts/v1.BedrockTargetVersion`，执行完整生成入口。

## 每张表由谁生成

| 生成步骤 | 输入与输出 |
| --- | --- |
| `export_defaults.py` | 导出各历史版本的默认属性，以及 `tmp/bedrock_defaults.json` 目标默认属性。 |
| `gen_fan_in_mapping.py` | 枚举历史完整与缺参输入，写出 `tmp/fan_in_{bedrock,java}.mapping`。 |
| `gen_fan_out_mapping.py` | 导出 Bedrock 全集、转出候选、Java 完整含水分析清单，以及 `java_states.list` 和 `java_block_schema.list`。 |
| `gen_java_compat.py` | 直接生成 `presets/java/v1/java_compatibility.list` 与诊断报告。 |
| `TestStep0ReadBedrockStates` | 从 `tmp/bedrock_states.list` 生成 `0_bedrock_blocks.bin`。 |
| `TestReadDefaultsReadBlocksConsistency` | 核对目标状态与默认属性的键，生成 `1_bedrock_default_props.bin`。 |
| `TestStep2_1_GenBedrockFanin` | 整理候选并生成 `2_bedrock_fan_in.bin`。 |
| `TestStep2_2_GenJavaDefaultPropFanin` | 合并历史默认属性，高版本优先，生成 `2_2_java_default_props.bin`。 |
| `TestStep2_3_GenJavaFanin` | 整理候选并生成 `2_3_java_fan_in.bin`。 |
| `TestStep3GenJavaFanout` | 生成 `3_to_java.bin`，检查全部目标 Bedrock 状态可转出。 |
| `TestStep4BuildBlockDataByName` | 从历史 palette 生成 `4_block_data.bin` 与 `5_legacy_blocks.bin`。 |
| `TestSchematic` | 生成 `6_schematic_fan_in.bin` 与 `7_schematic_to_java.json.gz`。 |

Java 文本资源位于 `presets/java/v1`，转换二进制位于 `presets/block-converts/v1`。默认属性的 `.bin` 内容是 JSON，其余二进制使用各自的紧凑格式。`*_tidy.mapping` 保留 Go 裁减后的诊断条目。

PyMCTranslate 1.2.53（build 396）、Java `(26, 3, 0)` 与 Bedrock `(26, 60, 0)` 的完整入口产物规模如下。Java 兼容报告记录 24,475 个唯一源状态、676 个需要上下文的候选和 0 个拒绝项。

| 数据 | 条目数 |
| --- | ---: |
| Java 单层全集 | 22,398 |
| Java 历史兼容映射 | 2,077 |
| Bedrock 全集 | 24,106 |
| Java 转入候选 / 裁减后映射 | 88,688 / 17,878 |
| Bedrock 转入候选 / 裁减后映射 | 104,522 / 20,090 |
| Bedrock 转出 Java 表 | 13,930 |

上表由完整入口生成后，在模块根目录用以下命令对最终文本与中间映射计数。兼容报告由 `gen_java_compat.py` 读取同一固定版本的 PyMCTranslate 生成。

```sh
wc -l presets/java/v1/java_states.list presets/java/v1/java_compatibility.list \
  third_party/gen_resource/block-converts/tmp/bedrock_states.list \
  third_party/gen_resource/block-converts/tmp/fan_in_java.mapping \
  third_party/gen_resource/block-converts/tmp/fan_in_java_tidy.mapping \
  third_party/gen_resource/block-converts/tmp/fan_in_bedrock.mapping \
  third_party/gen_resource/block-converts/tmp/fan_in_bedrock_tidy.mapping \
  third_party/gen_resource/block-converts/tmp/fan_out_java_tidy.mapping
```

单步运行 Python 时，在模块根目录执行 `.venv/bin/python third_party/gen_resource/block-converts/<脚本名>`。Go 步骤则进入 `third_party/gen_resource/block-converts`，按上表顺序执行，例如：

```sh
go test -tags resourcegen -v -count=1 -run '^TestStep0ReadBedrockStates$' .
```

## 验证资源与分析差异

在工作区根目录执行验收，在模块根目录执行示例：

```sh
./check.sh neomega-blocks
(cd neomega-blocks && go run ./examples/convert)
```

`TestBedrockPaletteSharedHashes` 比较新全集与历史 palette 中写法相同的状态及其网络哈希，`TestLegacyPaletteDataValues` 逐项验证历史 aux 反查，`TestBedrockJavaOutputsBelongToTarget` 要求全部 Bedrock 转出直接属于 Java 全集。Python 样例验证缺参补齐及旧 carpet、observer 的 blockstate 解释。方块状态的兼容生成不传递方块实体，避免告示牌文本等实体数据影响状态映射。

以 `block_palette.json` 和上述嵌入全集为输入，下面的命令得到 13,694 个同写法状态，网络哈希全部一致。样例同时覆盖 Bedrock 的字符串、整数和 byte 属性。

```sh
go test -v -count=1 -run '^TestBedrockPaletteSharedHashes$' ./third_party/gen_resource/block-converts
```

需要核对 Java 模型选择时，在模块根目录提供匹配目标版本的资源：

```sh
NEOMEGA_ASSETS_DIR=/path/to/resources \
  go test -v -count=1 -run '^TestJavaTargetAssets$' ./third_party/gen_resource/block-converts
```

检查读取 `build-info.json` 和 `blockstates.json`，验证目标版本、名称与 variant 选择。单层显示以 `waterlogged=false` 补全条件。

`java-roundtrip` 读取 `block-converts/tmp/hub_java_states.list`，执行 Java → Bedrock → Java 并统计名称、属性变化。报告反映转换链的信息损失，差异本身不能证明 Bedrock 无法表达。在模块根目录执行：

```sh
mkdir -p third_party/gen_resource/java-roundtrip/tmp
go run ./third_party/gen_resource/java-roundtrip > third_party/gen_resource/java-roundtrip/tmp/REPORT.md
```
