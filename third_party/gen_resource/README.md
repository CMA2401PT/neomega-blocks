# 预置资源的离线生成

这里的工具把上游方块定义和本地 NEMC palette 转换为 Go 运行时使用的预置资源。Python 使用 PyMCTranslate 枚举状态并生成候选映射；Go 依据 NEMC 目标集合和专用规则修正映射，再压缩成查询表。运行时包及其使用约定见 [`presets`](../../presets/README.md)。

脚本按目的组织：`block-converts` 生成方块转换表及它依赖的 Java 状态定义；`geometry/solid_block_table` 生成静态几何、Java 元数据和六方向实心表；`java-roundtrip` 分析已有预置转换的往返差异。脚本所需的固定输入跟随脚本保存，中间文件只写入各自的 `tmp/`。`tmp/` 被 Git 忽略，可删除后重新生成。最终嵌入资源直接写入 `presets` 的对应 Go 包。

## 准备 Python 环境

模块根目录的 `pyproject.toml` 声明 Python 3.11、PyMCTranslate 1.2.30 和 amulet-nbt 2.1.8；`uv.lock` 固定完整依赖。PyMCTranslate 通过 uv 安装，源码目录不保存在仓库中。

从 `neomega-blocks` 根目录初始化：

```sh
uv sync --locked
```

环境创建在 `neomega-blocks/.venv/`，该目录被 Git 忽略。Go 生成步骤读取 Python 导出的默认属性，不依赖上游包内部是 JSON 目录还是压缩数据库。

## block-converts：生成运行时方块资源

生成以 `block_palette.json` 定义的 NEMC 方块集合为中心。Bedrock 转换参考为 1.21.50，Java 输出目标为 1.21.4。源版本按从高到低的顺序枚举，相同输入优先保留较高版本的解释。

固定输入包括：

- `block_palette.json`：NEMC 的实际方块名、状态、网络哈希和 data value。`read_blocks.go` 将其转换为 tidy IRBlock，并核对网络哈希。
- `nemc3.7_be1.21.50_block_palette.nbt`：保留的原始 palette 参考文件；当前生成入口读取前面的 JSON。
- `schematic.json`：旧 Schematic 的 `id:meta` 到 Java 方块文本的映射。
- uv 安装的 PyMCTranslate：提供各版本的默认属性、合法取值和经 universal 中间表示连接的转换关系。来源声明见 [NOTICE](../pymctranslate-notices/NOTICE.md)。

NEMC 的工作台、熔炉和点燃熔炉多一个 `custom_appearance` byte 属性。`pymc_specification.py` 在读取 Bedrock 1.21.50 specification 时补上其取值和默认值。这个规则用于枚举和默认属性导出，不改动安装环境中的上游库。

主要代码各自承担这些工作：

| 入口或代码 | 解决的问题 |
| --- | --- |
| `generate.sh` | 初始化 uv 环境，依次运行 Python 和 Go 生成步骤，最后验证预置包。 |
| `export_defaults.py` | 通过库接口导出各版本默认属性，写入 `tmp/defaults/{platform}_{version}.json`，供 Go 读取。 |
| `gen_fan_in_mapping.py` | 将各版本 Bedrock、Java 状态转入 Bedrock 参考版本，生成 `tmp/fan_in_bedrock.mapping` 和 `tmp/fan_in_java.mapping`。 |
| `gen_fan_out_mapping.py` | 生成 `tmp/fan_out_java.mapping`、允许的 Java 状态清单，以及直接写入 `presets/java/v1` 的完整状态全集和属性定义。 |
| `step0_read_blocks_test.go`、`step1_read_consistency_test.go` | 建立目标方块集合，并核对 palette 与默认属性的 key 集合。 |
| `step2_gen_fan_in_test.go` | 修正旧名称和属性差异，裁减可由 Canon 复用的映射，建立 Bedrock 与 Java 转入表。 |
| `step3_gen_fan_out_test.go` | 建立目标 Bedrock 到 Java 的映射，检查 palette 中的方块均能转出。 |
| `step4_block_data.go`、`step4_block_data_test.go` | 建立名称、legacy data value 到网络哈希的索引。 |
| `step5_schematic_test.go` | 将旧 Schematic ID/meta 经 Java 转入逻辑连接到目标集合。 |
| `test_mapping_generators.py` | 用内存样例检查缺参输入键对应的转换对象始终补齐默认属性。 |
| `java_assets_test.go` | 可选地核对 palette 的 Java 转出结果能选择到匹配版本的模型资源。 |

Python 可以保留缺参形式作为映射键，但传给 `to_universal` 的对象必须补齐默认属性。脚本中的 `allow_missing` 参数为兼容原调用保留，不改变此行为。运行时由 Canon 补齐默认值，再复用完整状态的转换结果。

### 最终嵌入资源

以下产物与消费它们的 Go 文件放在一起。默认属性的 `.bin` 文件实际保存 JSON，其余 `.bin` 使用各自的紧凑存储格式。

| 最终目录 | 文件 | 生成步骤 |
| --- | --- | --- |
| `presets/block-converts/v1` | `0_bedrock_blocks.bin` | `TestStep0ReadBlocksFromPalette` |
| 同上 | `1_bedrock_default_props.bin` | `TestReadDefaultsReadBlocksConsistency` |
| 同上 | `2_bedrock_fan_in.bin` | `TestStep2_1_GenBedrockFanin` |
| 同上 | `2_2_java_default_props.bin` | `TestStep2_2_GenJavaDefaultPropFanin` |
| 同上 | `2_3_java_fan_in.bin` | `TestStep2_3_GenJavaFanin` |
| 同上 | `3_to_java.bin` | `TestStep3GenJavaFanout` |
| 同上 | `4_block_data.bin` | `TestStep4BuildBlockDataByName` |
| 同上 | `6_schematic_fan_in.bin` | `TestSchematic` |
| `presets/java/v1` | `fan_out_java_full.list`、`java_block_schema.list` | `gen_fan_out_mapping.py` |

`*_tidy.mapping` 是 Go 筛选后的诊断映射，仍属于中间文件，保存在 `block-converts/tmp/`。

### 执行完整生成

需要已安装 uv 和 Go，并从带有现成嵌入资源的仓库开始。Go 生成端会调用预置包；这些包在生成端编译时就需要读取现有嵌入资源。

在模块根目录执行：

```sh
sh third_party/gen_resource/block-converts/generate.sh
```

入口脚本可以从其他目录调用，它会先切换到自身目录。生成中遇到失败会停止；检查失败原因后重新运行对应步骤。生成失败可能留下部分更新的资源，提交前应检查产物差异。

需要单步检查时，在模块根目录准备 Python 中间文件：

```sh
.venv/bin/python -m unittest discover -s third_party/gen_resource/block-converts -p test_mapping_generators.py
.venv/bin/python third_party/gen_resource/block-converts/export_defaults.py
.venv/bin/python third_party/gen_resource/block-converts/gen_fan_in_mapping.py
.venv/bin/python third_party/gen_resource/block-converts/gen_fan_out_mapping.py
```

然后进入 `third_party/gen_resource/block-converts`，按最终产物表中从上到下的 Go 步骤逐个执行，例如：

```sh
go test -tags resourcegen -v -count=1 -run '^TestStep0ReadBlocksFromPalette$' .
go test -tags resourcegen -v -count=1 -run '^TestReadDefaultsReadBlocksConsistency$' .
```

每个步骤都必须单独启动 `go test`，使后一步重新编译并读取前一步更新的 embed。`-count=1` 保证实际执行生成。写资源的入口使用 `resourcegen` 构建标签，普通 `go test ./...` 和工作区 `check.sh` 不会执行它们。

更换目标版本时，需要一起核对 Python 脚本中的参考版本、`pymc_specification.py` 的 NEMC 规则、Go 生成步骤中的默认属性版本，以及两个预置包公开的版本常量。palette 与专用转换规则也必须匹配新的目标集合。

## geometry/solid_block_table：生成静态几何与实心表

viewer 和 fixer 需要共享同一套模型解释。这个目录使用 `geometry/java_block_model` 的模型定义、读取与烘焙，以及 `geometry` 的覆盖计算，统一生成 Java 几何表、Java 元数据和实心表。三份产物直接写入 `presets/java/v1/geometry` 与 `presets/java/v1/solidtable`。

先从一套匹配的 assets 手动复制六份输入到生成目录，再在模块根目录运行：

```sh
go generate ./presets/java/v1/geometry
```

输入列表、内部代码职责和真实资源验证命令见[几何生成说明](geometry/solid_block_table/README.md)。这条流程使用 Go；Python 环境用于前面的方块转换资源。

## java-roundtrip：生成转换差异报告

`main.go` 枚举 `presets/java/v1.JavaFullStates()`，执行 Java → NEMC → Java，再将输入输出 Tidy 后比较。报告按方块和属性统计差异，帮助定位转换丢失的信息。它不单独证明某个状态无法在 Bedrock 表达，也不保证所有差异都能从邻居恢复。

这个工具读取已经嵌入的预置数据，无需 Python。报告是分析结果，保存到它自己的 `tmp/`。在模块根目录运行：

```sh
mkdir -p third_party/gen_resource/java-roundtrip/tmp
go run ./third_party/gen_resource/java-roundtrip > third_party/gen_resource/java-roundtrip/tmp/REPORT.md
```

## 验证资源

完整生成入口已包括默认属性解析、data value 索引的导出载入检查和所有预置包测试。日常修改在工作区根目录运行：

```sh
./check.sh neomega-blocks
```

如需核对 Java 输出与渲染资源，在模块根目录执行：

```sh
NEOMEGA_ASSETS_DIR="$(cd ../neomega-assets/dist/final && pwd)" \
  go test -v -count=1 -run '^TestJavaTargetAssets$' ./third_party/gen_resource/block-converts
```

该检查读取 `build-info.json` 和 `blockstates.json`，验证目标版本、方块名称和 variant 选择。单层预览以 `waterlogged=false` 补全显示条件；处理存档时仍需结合独立水层确定含水状态。
