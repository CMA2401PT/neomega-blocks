# 预置资源的离线生成

这里的工具把上游方块定义和 NEMC palette 转成运行时预置资源。转换中心以 NEMC 方块集合连接多版本 Bedrock、Java 与旧 Schematic，Java 侧使用 1.21.4。浏览、渲染与 fixer 需要认识当前目标存档，因此独立的 Java 状态知识以 26.3 为目标。运行时包的使用约定见 [`presets`](../../presets/README.md)。

两套 Java 状态通过迁移表连接。迁移表记录中心状态到当前目标状态的名称与属性变化，历史状态则先经 Amulet 兼容映射解释为中心状态。`block-converts` 生成转换中心及其中间结果，`java-states` 连接 Mojang 报告与这些中间结果，`java-roundtrip` 分析转换中心的往返差异。固定输入跟随脚本保存，中间文件写入各自被 Git 忽略的 `tmp/`，最终资源写入 `presets` 的消费包。

## 数据来源与版本分工

转换中心使用 PyMCTranslate（Amulet）的跨版本转换数据，Java 数据使用范围固定到 1.21.4，以尊重 Amulet 的许可。converter 按此版本写 Java 存档，游戏打开时自行升级。Bedrock 转换参考为 1.21.50，NEMC 的实际目标集合由 `block_palette.json` 决定。来源声明见 [NOTICE](../pymctranslate-notices/NOTICE.md)。

Java 状态知识来自 Mojang 官方服务端的数据生成器。官方报告列出注册状态、默认状态和属性域。`java/v1.JavaTargetVersion` 选择当前目标，包名里的 `v1` 表示预置数据格式修订号。`java-states` 去掉 `waterlogged` 后生成单层状态，因为含水由独立层表达。

## 准备环境与官方报告

模块根目录的 `pyproject.toml` 声明 Python 3.11、PyMCTranslate 1.2.30 和 amulet-nbt 2.1.8；`uv.lock` 固定依赖。初始化命令在 `neomega-blocks` 根目录执行：

```sh
uv sync --locked
```

环境创建在 `.venv/`。Go 生成步骤读取 Python 导出的默认属性，输入与上游内部存储形式无关。

Java 状态生成需要先准备与目标常量一致的报告。在安装了目标游戏版本支持的 Java 运行时后执行：

```sh
.venv/bin/python third_party/gen_resource/java-states/prepare-report.py --version 26.3
```

脚本从 Mojang version manifest 定位版本，下载版本元数据与服务端，运行官方数据生成器，把 `blocks.json` 和服务端内的 `version.json` 放到 `java-states/tmp/`。指定 Java 可执行文件与单步生成的方法见 [java-states 说明](java-states/README.md)。

## block-converts：建立转换中心

生成以 `block_palette.json` 定义的 NEMC 方块集合为中心。源版本从高到低枚举，相同输入优先保留较高版本的解释。

固定输入包括：

- `block_palette.json` 保存实际方块名、状态、网络哈希与 data value。`read_blocks.go` 将其转成 tidy IRBlock，并核对网络哈希。
- `nemc3.7_be1.21.50_block_palette.nbt` 保存原始 palette 参考；当前入口读取前面的 JSON。
- `schematic.json` 保存旧 Schematic 的 `id:meta` 到 Java 方块文本的映射。
- PyMCTranslate 提供默认属性、合法取值与 Universal 转换关系。Universal 是连接各版本方块的中间表示。

NEMC 的工作台、熔炉和点燃熔炉多一个 `custom_appearance` byte 属性。`pymc_specification.py` 在读取 Bedrock 1.21.50 specification 时补上其取值和默认值，供枚举与默认属性导出使用。

| 入口或代码 | 解决的问题 |
| --- | --- |
| `generate.sh` | 初始化环境，依次运行 Python 中间结果、Java 状态迁移和 Go 转换表生成，最后验证预置包。 |
| `export_defaults.py` | 导出各版本默认属性到 `tmp/defaults/{platform}_{version}.json`。 |
| `gen_fan_in_mapping.py` | 将多版本输入转入参考 Bedrock，写出 `tmp/fan_in_bedrock.mapping` 与 `tmp/fan_in_java.mapping`。 |
| `gen_fan_out_mapping.py` | 写出转出候选、允许状态和中心完整状态 `tmp/hub_java_states.list`。 |
| `gen_java_compat.py` | 写出历史 Java → 中心状态的 `tmp/hub_java_compatibility.list` 和诊断报告。 |
| `hub_java.go` | 读取并整理中心状态与源#目标映射，供原理图、迁移生成和往返分析共用。 |
| `step0_read_blocks_test.go`、`step1_read_consistency_test.go` | 建立目标集合，并核对 palette 与默认属性的 key 集合。 |
| `step2_gen_fan_in_test.go` | 修正旧名称与属性，裁减 Canon 可复用的映射；使用迁移表为 Java 转入键补充当前目标形式。 |
| `step3_gen_fan_out_test.go` | 建立 Bedrock → Java 中心状态的映射，检查 palette 均能转出。 |
| `step4_block_data.go`、`step4_block_data_test.go` | 建立名称、legacy data value 与网络哈希的索引。 |
| `step5_schematic_test.go` | 使用中心完整状态与兼容表对齐历史名称，生成直接 Java 1.21.4 表与 Bedrock 转入表。 |
| `test_mapping_generators.py` | 用内存样例验证缺参输入对应的转换对象会补齐默认属性。 |
| `java_assets_test.go` | 可选地核对中心 Java 输出与匹配版本的模型资源。 |

Python 可以保留缺参形式作为映射键，但传给 `to_universal` 的对象必须补齐默认属性。保留的 `allow_missing` 参数不改变此行为。运行时由 Canon 补默认值，再复用完整状态映射。

### 最终嵌入资源

资源与消费它们的 Go 文件放在一起。默认属性的 `.bin` 实际保存 JSON，其余 `.bin` 使用各自的紧凑格式。

| 最终目录 | 文件 | 生成步骤 |
| --- | --- | --- |
| `presets/block-converts/v1` | `0_bedrock_blocks.bin` | `TestStep0ReadBlocksFromPalette` |
| 同上 | `1_bedrock_default_props.bin` | `TestReadDefaultsReadBlocksConsistency` |
| 同上 | `2_bedrock_fan_in.bin` | `TestStep2_1_GenBedrockFanin` |
| 同上 | `2_2_java_default_props.bin` | `TestStep2_2_GenJavaDefaultPropFanin` |
| 同上 | `2_3_java_fan_in.bin` | `TestStep2_3_GenJavaFanin` |
| 同上 | `3_to_java.bin` | `TestStep3GenJavaFanout` |
| 同上 | `4_block_data.bin` | `TestStep4BuildBlockDataByName` |
| 同上 | `6_schematic_fan_in.bin`、`7_schematic_to_java.json.gz` | `TestSchematic` |
| `presets/java/v1` | `java_states.list`、`java_block_schema.list`、`java_compatibility.list` | `java-states/TestGenerateJavaStates` |

`*_tidy.mapping` 是 Go 筛选后的诊断映射，保存在 `block-converts/tmp/`。

## java-states：连接历史输入与当前目标

生成器按“历史状态 → Amulet 兼容表 → Java 1.21.4 → 迁移表 → Java 26.3”连接状态。目标全集已有的状态直接保留身份。中心状态的改名和属性替换由 `java-states/migrate.go` 定义，迁移后用目标默认状态补齐缺项；迁移结果不属于目标全集时，错误指出需要补规则的源状态。

`gen_java_compat.py` 按源版本从高到低选择同一完整状态的解释。需要邻居或方块实体的候选，在能形成合法中心状态时采用默认结果，并写入 `tmp/java_compatibility_report.json` 的 `contextDependent`。报告还记录无法形成合法中心状态的候选及原因。状态兼容用于显示与修复，旧 MCA/NBT 布局的读取仍由存档读取器处理。

用锁定依赖运行下节两个 Python 脚本，输入为 PyMCTranslate 的 Java 历史版本与中心 1.21.4，得到 18,254 个中心状态和 2,062 条中心兼容映射。兼容报告中有 666 个需要上下文的候选，没有被拒绝的候选。`TestGenerateJavaStates` 用这些中间结果和 Mojang 26.3 报告生成 22,398 个单层状态、2,077 条兼容映射与 15 条中心迁移；命令及输入位置见 [java-states 说明](java-states/README.md)。

## 按依赖顺序生成

完整生成需要 uv、Go、预先准备的 Mojang 报告，以及仓库现有的嵌入资源。Go 生成端调用预置包，编译时就需要读取现有资源。在模块根目录运行：

```sh
sh third_party/gen_resource/block-converts/generate.sh
```

入口可以从其他目录调用。发生错误时脚本停止，生成失败可能留下部分更新的资源，检查原因后可重跑对应步骤。

需要单步生成时，在模块根目录执行 Python 步骤：

```sh
.venv/bin/python -m unittest discover -s third_party/gen_resource/block-converts -p test_mapping_generators.py
.venv/bin/python third_party/gen_resource/block-converts/export_defaults.py
.venv/bin/python third_party/gen_resource/block-converts/gen_fan_in_mapping.py
.venv/bin/python third_party/gen_resource/block-converts/gen_fan_out_mapping.py
.venv/bin/python third_party/gen_resource/block-converts/gen_java_compat.py
```

然后先生成 Java 状态与迁移表：

```sh
(cd third_party/gen_resource/java-states && go test -tags resourcegen -v -count=1 -run '^TestGenerateJavaStates$' .)
```

再进入 `third_party/gen_resource/block-converts`，按产物表顺序逐个运行 Go 步骤，例如：

```sh
go test -tags resourcegen -v -count=1 -run '^TestStep0ReadBlocksFromPalette$' .
go test -tags resourcegen -v -count=1 -run '^TestReadDefaultsReadBlocksConsistency$' .
```

每个步骤单独启动 `go test`，使下一步重新编译前一步更新的 embed。`-count=1` 保证执行生成。写资源入口使用 `resourcegen` 标签，普通测试和工作区检查不会执行它们；报告解析与迁移逻辑由普通单元测试覆盖。

更新 Java 状态目标时，先准备新报告，再修改 `java/v1.JavaTargetVersion`，补迁移规则并重新生成。转换中心的 Java 侧保持 1.21.4。更换中心集合时，需核对 Python 参考版本、NEMC 扩展、默认属性、palette 与专用规则。

## java-roundtrip：分析中心状态的往返差异

工具读取 `block-converts/tmp/hub_java_states.list`，执行 Java → NEMC → Java，并将输入输出 Tidy 后比较。报告按方块和属性统计差异，帮助定位转换丢失的信息。差异不能单独证明 Bedrock 无法表达源状态，也不能保证从邻居恢复。

先用 `gen_fan_out_mapping.py` 准备中心状态清单，再在模块根目录运行：

```sh
mkdir -p third_party/gen_resource/java-roundtrip/tmp
go run ./third_party/gen_resource/java-roundtrip > third_party/gen_resource/java-roundtrip/tmp/REPORT.md
```

## 验证资源

完整入口包括默认属性解析、data value 索引导出载入检查与预置包测试。日常修改在工作区根目录执行：

```sh
./check.sh neomega-blocks
```

核对中心输出的模型选择时，提供与中心 Java 版本一致的资源目录，在模块根目录执行：

```sh
NEOMEGA_ASSETS_DIR=/path/to/resources \
  go test -v -count=1 -run '^TestJavaTargetAssets$' ./third_party/gen_resource/block-converts
```

检查读取 `build-info.json` 和 `blockstates.json`，验证中心版本、方块名称与 variant 选择。单层选择以 `waterlogged=false` 补全显示条件；存档含水状态仍需结合独立水层确定。面向当前目标的渲染资源应使用 `java/v1` 的版本与状态全集。
