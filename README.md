# neomega-blocks

`neomega-blocks` 用来描述、查找和转换 Minecraft 方块。不同版本可能使用不同的方块名、属性名和属性值；旧格式还可能只给出数字 ID 与附加值。这个库把这些输入转换到一套确定的 NEMC/Bedrock 方块集合，并提供向 Java 方块的输出转换。

项目分为两个相互配合的部分：**运行时读取已经准备好的转换表；离线工具负责生成、修正和验证这些表。** 使用现成 preset 时，只需调用 Go 接口。维护目标版本或新增方块支持时，才需要运行 `presets/block-converts/scripts` 下的准备流程。

## 为什么要有一个转换中心

如果让每个版本都直接转换到其他所有版本，转换关系会越来越多。这里选择以 NEMC palette 中的方块集合作为中心：先把输入归入这个集合，再根据需要输出。

```text
其他版本的 Bedrock ─┐
Java 方块 ─────────┼─→ preset 的 NEMC/Bedrock 方块集合 ─→ Java 方块
旧 Schematic ID ───┘                ↕
                           网络哈希、名称与 data value
```

这里的“当前方块集合”由 preset 的数据决定，不是运行时检测出来的游戏版本。生成脚本以 Bedrock `1.21.50` 为转换参考，Java 输出目标为 `1.21.4`。运行时可通过 `v1.JavaTargetVersion` 和 `v1.BedrockTargetVersion` 查询预设目标；渲染资源应与 Java 目标一致。实际可用的 NEMC 方块由输入的 `block_palette.json` 决定。

转换通常不能保证可逆。不同源方块可能归到同一个目标方块，某些源属性也没有对应的目标属性。例如，NEMC 的 `micro_block` 在 Java 输出表中被映射为 `air`。

## 先统一怎样描述一个方块

要连接不同来源，首先需要一种方便比较、修改和查看的表示。`ir.IRBlock` 是“方块名 + 状态”的文本中间表示，例如：

```text
minecraft:nether_wart ["age"=3]
minecraft:jungle_fence_gate ["open"="true"]
```

`IRStates` 表示方括号里的属性。它支持字符串、`int32`，以及表示布尔值的 `int8(0/1)`。文本中的 `true` 与字符串 `"true"` 保留类型区别。IR 统一了表达形式，Java 和 Bedrock 的属性含义仍由转换规则处理。

为了让调用者从文本逐步走到可转换的状态，库提供三种不同操作：

| 操作 | 解决的问题 |
| --- | --- |
| `IsValid` | 检查文本语法，不能据此判断游戏是否存在这个方块 |
| `Tidy` | 补默认命名空间、整理空白和属性顺序 |
| `BedrockCanon` / `JavaCanon` | 按已知默认属性补缺项、删多余项，并尝试统一属性类型 |

`Canon` 成功表示找到该名称的默认属性，不代表每个属性值都属于游戏允许的取值。需要确认目标方块确实存在时，应使用转换结果或目标表查询。`Tidy` 是宽松整理操作，无法整理的 states 可能回退为空；需要拒绝语法错误的输入时，先调用 `IsValid`。

### 缺少属性为什么仍然可以转换

`block[a=opt1,b=opt2]` 省略一个属性后成为 `block[a=opt1]`，仍然是需要支持的合法输入。缺少的属性应使用对应默认值，而不是让转换器自行猜测。

PyMCTranslate 对缺参输入的处理并不可靠，可能 fallback 到错误方块。因此，缺参支持依靠默认属性归一化：

```text
合法的缺参方块 → Canon 补齐默认属性 → 完整状态的转换规则或映射
```

Python 脚本保留的 `allow_missing` 参数没有行为作用，不应重新启用“把缺参对象直接交给 PyMCTranslate”的分支。脚本枚举过程中即使保留缺参形式作为输入键，实际交给 `to_universal` 的也是补齐后的 `block_with_default`。运行时通过 Canon 尽量复用完整状态的映射。

## 完成第一次转换

上述处理已经组合在 `presets/block-converts/v1` 中。普通使用者无需自己读取映射文件，也无需安装 Python 或 PyMCTranslate。

项目使用 Go 1.25.5 或更高版本。模块路径是 `neomega-blocks`，并通过工作区使用相邻的 `neomega-nbt`。目录应类似：

```text
父目录/
  go.work
  neomega-nbt/
  neomega-blocks/
```

若尚无工作区，在父目录执行：

```sh
go work init ./neomega-nbt ./neomega-blocks
```

下面的程序把缺少部分属性的 Java 栅栏门转为目标 Bedrock 方块，再通过网络哈希找回它。保存到 `neomega-blocks/examples/basic/main.go` 后，在 `neomega-blocks` 根目录运行 `go run ./examples/basic`。

```go
package main

import (
    "fmt"

    "neomega-blocks/ir"
    blocks "neomega-blocks/presets/block-converts/v1"
)

func main() {
    input := ir.IRBlock(`minecraft:jungle_fence_gate ["open"="true"]`)
    bedrock, ok := blocks.ConvertFromJava(input)
    if !ok {
        panic("unsupported Java block")
    }
    fmt.Println(bedrock)

    restored, ok := blocks.BedrockHashToBlockStrict(bedrock.Hash())
    fmt.Println(restored == bedrock, ok)
}
```

输出：

```text
minecraft:jungle_fence_gate ["direction"=2,"in_wall_bit"=false,"open_bit"=true]
true true
```

需要自己构造属性时，可以使用 `ir.NewIRBlockFromGo(name, props)`，其中 `props` 为 `map[string]any`。需要修改已有属性时，`block.States().Set(...)`、`Delete(...)`、`Update(...)` 会返回新的 states；将它与原名称组合后得到新方块。

## 根据输入选择入口

不同输入携带的信息不同，所以不必先把所有数据都转成字符串。preset 提供这些入口：

| 已有数据或目标 | 入口 |
| --- | --- |
| Bedrock 方块名与状态 | `ConvertFromBedrock(IRBlock)` |
| Java 方块名与状态 | `ConvertFromJava(IRBlock)` |
| 旧 Schematic 数字 ID、meta | `ConvertFromSchematic(id, meta)` |
| 目标 Bedrock 方块转 Java | `ConvertToJava(IRBlock)` |
| 已知网络哈希，取目标方块 | `BedrockHashToBlockStrict(hash)` |
| 遍历目标集合中的网络哈希与方块 | `BedrockBlocks()`，返回只读迭代器，顺序不固定 |
| 名称与 legacy data value，取哈希或方块 | `BedrockDataValueToHash(name, value)` / `BedrockDataValueToBlock(name, value)` |
| 网络哈希反查名称与 data value | `BedrockHashToBlockData(hash)` |

读取返回的 `bool` 判断是否找到转换或数据。`ConvertFrom*` 失败时返回 `minecraft:unknown []` 和 `false`；不要只检查返回的方块文本。`ConvertToJava` 面向 preset 的目标 Bedrock 表示，处理其他版本时可以先调用 `ConvertFromBedrock`。

这些接口没有源版本参数。多个版本中相同文本的解释由生成表时的优先级决定；Java 默认属性按名称合并，高版本优先。因此，它是一套预先选定策略的兼容转换，不是可在运行时任选版本的翻译器。

## 为什么运行时只需要这些表

跨版本转换的复杂工作可以提前做完。运行时按需加载内嵌数据，使用 `sync.Once` 保证每张表只加载一次：

- Bedrock 输入先尝试目标表直接命中，再尝试 Canon，最后查询转入表。
- Java 输入优先处理墙、悬挂告示牌等特殊规则，再用 Canon 后的状态查表，必要时尝试原输入。
- 向 Java 输出时，使用特殊规则或目标 Bedrock 哈希到 Java IR 的映射。

为了减少大量状态文本的重复，`storage/blocks-name` 把方块名称和 states 分开保存。states 使用共享字符串片段的紧凑前缀树，哈希映射到重建索引。`storage/num_lookup` 用有序数组保存静态数字映射，并允许通过动态层覆盖或追加数据；`Frozen` 将动态层并入静态层，`Dump` 导出合并后的快照。

方块哈希也复用了 `neomega-nbt`：将 `name` 与 `states` 表达为 NBT 操作流，以小端二进制形式送入 FNV-1a 32 位写入器。它不是对 IR 文本直接求哈希。调用 `Hash` 前应使用整理后的名称与有序状态；`Hash` 本身不负责补齐默认属性。

表加载在地址满足对齐要求时直接引用输入字节；不满足时复制到对齐存储。加载器可能共享输入内存，因此调用者应避免修改输入字节。静态表与前缀树的 dump 含 Go 数字或节点的原生内存布局，生成和使用时应保持兼容的布局与字节序。

## 转换表从哪里来

要更新目标版本，必须知道每张表的依据。`scripts` 是数据准备工具：Python 借助 PyMCTranslate 生成候选映射，Go 再把候选结果修正为 NEMC 目标集合里的方块，并压缩为运行时表。

准备阶段使用三类输入：

- `block_palette.json`：目标 NEMC 的实际方块集合，包括名称、属性类型、网络哈希和 data value。`ReadBlocks` 会核对计算哈希与 palette 记录是否一致。
- 本地 `PyMCTranslate/`：转换代码及 `json/versions` 数据库，提供默认属性与经 universal 表示中转的跨版本映射。生成脚本从高版本向低版本遍历，重复输入优先保留高版本结果。
- `schematic.json`：旧 `id:meta` 到 Java 方块文本的映射，再经 Java 转入逻辑连接到目标集合。

这些数据有独立的许可来源。本地 PyMCTranslate 记录 build 317，对应上游 1.2.30；其 Amulet Team License 1.0.0 包含非商业、教育和非竞争等使用条款。它不是 MIT 许可，生成运行时表也不应被视为消除了来源条款。原文、Required Notice 和版本依据见 [third_party/pymctranslate-notices/NOTICE.md](third_party/pymctranslate-notices/NOTICE.md)。更新上游版本前，应核对该版本的许可。

这些输入产生的运行时文件如下。扩展名为 `.bin` 的默认属性表实际保存 JSON；其余文件使用各自的紧凑存储格式。

| 产物，均位于 `presets/block-converts/v1` | 作用 | 生成入口 |
| --- | --- | --- |
| `0_bedrock_blocks.bin` | 目标网络哈希 → 方块文本，并附带 data value | `TestStep0ReadBlocksFromPalette` |
| `1_bedrock_default_props.bin` | Bedrock 名称 → 默认属性 | `TestReadDefaultsReadBlocksConsistency` |
| `2_bedrock_fan_in.bin` | 其他 Bedrock 状态哈希 → 目标哈希 | `TestStep2_1_GenBedrockFanin` |
| `2_2_java_default_props.bin` | Java 名称 → 默认属性 | `TestStep2_2_GenJavaDefaultPropFanin` |
| `2_3_java_fan_in.bin` | Java 状态哈希 → 目标哈希 | `TestStep2_3_GenJavaFanin` |
| `3_to_java.bin` | 目标哈希 → Java 方块文本 | `TestStep3GenJavaFanout` |
| `4_block_data.bin` | 名称、data value → 目标哈希 | `TestStep4BuildBlockDataByName` |
| `6_schematic_fan_in.bin` | Schematic ID、meta → 目标哈希 | `TestSchematic` |

Go 准备步骤并不是单纯打包候选数据。`outputFix` 修正旧名称、拆分后的方块和属性差异；墙、悬挂告示牌等交给专用规则；能够通过 Canon 复用的映射会被裁减。最终结果还会与目标集合核对。

## 如何更新方块表

更新表意味着重新确认“目标集合是什么”和“源输入怎样归入它”。下面按依赖顺序执行，所有命令均在标出的目录运行。已有 `.bin` 也是生成工具编译时的输入，应从带有这些产物的仓库开始。

### 1. 准备目标数据与版本

先替换 `scripts/block_palette.json`，并准备匹配的 PyMCTranslate 代码及数据库。palette 的 JSON 结构见 `scripts/read_blocks.go`：顶层 `blocks`，每个方块含 `name`、`network_id`、`data` 和带 `name/type/value` 的 `states`。属性类型为 `byte`、`int`、`string`。

更换参考版本时，需要一起核对：

- `gen_fan_in_mapping.py` 末尾的 Bedrock 输出版本。
- `gen_fan_out_mapping.py` 的 Bedrock 输入版本与 Java 输出版本。
- `presets/block-converts/v1/target.go` 中公开给消费者的版本常量。
- `step1_read_consistency_test.go`、`read_defaults_test.go` 中的 Bedrock specification 目录。
- 与版本相关的默认值测试、`outputFix` 和 `specific-converts` 规则。

`schematic.json` 只在旧格式输入范围发生变化时更新。PyMCTranslate 目录被 Git 忽略，palette 和部分中间文件也可能没有纳入版本控制；更换环境时需要另行准备这些输入。

### 2. 生成 Python 候选映射

此步骤使用 Python 3.11、`amulet-nbt` 及其依赖。下面的命令通过 uv 的独立环境运行，并固定 `amulet-nbt 2.1.8`；本地 `PyMCTranslate/` 从脚本目录导入。

先进入脚本目录，验证“缺参键与完整转换对象”的约定。这个测试使用小型内存样例，不生成真实映射：

```sh
# 从 neomega-blocks 根目录进入。
cd presets/block-converts/scripts
uv run --no-project --python 3.11 --with 'amulet-nbt==2.1.8' python -m unittest test_mapping_generators.py
```

然后生成候选映射：

```sh
uv run --no-project --python 3.11 --with 'amulet-nbt==2.1.8' python gen_fan_in_mapping.py
uv run --no-project --python 3.11 --with 'amulet-nbt==2.1.8' python gen_fan_out_mapping.py
```

第一条生成 `fan_in_bedrock.mapping`、`fan_in_java.mapping`。第二条生成 `fan_out_java.mapping` 和用于检查输出的 `fan_out_java_allowed.list`。缺参形式可以作为候选输入键，但转换调用必须继续使用补齐后的对象；不要让 `allow_missing` 改变这条原则。

### 3. 建立目标集合与默认属性

后续步骤要判断候选转换是否落在真实目标集合内，所以先建立目标表。以下命令仍在 `scripts` 目录运行：

```sh
go test -v -count=1 -run '^TestStep0ReadBlocksFromPalette$' .
go test -v -count=1 -run '^TestReadDefaultsReadBlocksConsistency$' .
```

第一步核对 palette 哈希并生成方块表；第二步比较 palette 与参考默认属性的 key 集合，再生成默认属性表。出现不匹配时，应检查数据版本或修正规则后重新执行，不能把失败产生的文件作为成功产物继续发布。

### 4. 依次生成转入、转出和旧格式索引

这些步骤使用前面的目标表来验证转换结果：

```sh
go test -v -count=1 -run '^TestStep2_1_GenBedrockFanin$' .
go test -v -count=1 -run '^TestStep2_2_GenJavaDefaultPropFanin$' .
go test -v -count=1 -run '^TestStep2_3_GenJavaFanin$' .
go test -v -count=1 -run '^TestStep3GenJavaFanout$' .
go test -v -count=1 -run '^TestStep4BuildBlockDataByName$' .
go test -v -count=1 -run '^TestSchematic$' .
```

每一条通过后再执行下一条。`go:embed` 在编译时读取文件，因此这些步骤必须分别启动 `go test`，让后一步重新编译并读取前一步的新产物；不要合并成一次运行整个脚本包。`-count=1` 则保证生成步骤实际执行。

生成过程中还会写出 `*_tidy.mapping`，便于检查裁减后的映射。Java 输出检查会覆盖 palette 中的方块；特殊转换也会与 Java 允许的状态集合核对。

### 5. 验证产物并检查差异

先在 `scripts` 中运行只读的数据检查：

```sh
go test -v -count=1 -run '^(TestReadDefaultsAllJSON|TestStep4StorageDumpLoad)$' .
```

再回到 `neomega-blocks` 根目录验证运行时：

```sh
cd ../../..
go test ./ir/... ./storage/... ./encode-decode/... ./presets/block-converts/v1 ./presets/block-converts/specific-converts
git diff --stat
git status --short
```

需要接入 `neomega-assets` 的渲染资源时，还应验证真实 palette 的转换结果能选择到资源里的模型。在 `neomega-blocks` 根目录执行：

```sh
NEOMEGA_ASSETS_DIR="$(cd ../neomega-assets/dist/final && pwd)" \
  go test -v -count=1 -run '^TestJavaTargetAssets$' ./presets/block-converts/scripts
```

该检查核对目标版本、方块名称和 variant 选择。PyMCTranslate 将水表示为附加方块，单层方块预览以 `waterlogged=false` 补全这项显示条件；区块应用需要结合独立水层决定含水状态。

检查目标版本、生成日志和实际方块样例，再将相应代码与 `.bin` 一起交付。失败步骤可能已经写入部分文件，所以更新应在可检查、可恢复的工作目录中进行。

## 怎样扩展

新增需求首先要判断：变化来自输入形式、单个转换规则，还是整个目标集合。把修改放在对应层，才能避免为一个方块改动全部数据流程。

**接入新的输入形式**时，先将名称与属性构造成 IR，再接入现成 preset。需要直接处理 NBT 操作流时，`IRBlock.DecodeToHL/DecodeToLL` 可以向 `neomega-nbt` 接收端输出；`NewIRBlockEncoder` 可以接收包含 `name` 和 `states` 的低层 NBT 操作，重建 IR。

**修正某类方块的语义转换**时，参考 `specific-converts` 中的墙和悬挂告示牌规则。接入运行时入口的同时，也要让生成端识别同一规则，避免把同类转换再次写入表中。测试应覆盖显式属性、缺参补默认值，以及确实可逆的状态往返。

**增加表驱动映射或更新目标版本**时，沿上述准备流程修改数据来源、默认属性和 `outputFix`，重新生成受影响的产物。扩展旧 Schematic 输入则从 `schematic.json` 开始，并使用已有 Java 转入逻辑生成最终索引。

若只需要自定义查询表，可以单独使用 `storage/blocks-name` 或 `storage/num_lookup`；preset 的全局表没有提供运行时注册接口。`ir/ir-go` 中的整块 Go 对象实现目前被注释，实际运行时中间表示是 `ir.IRBlock`。

## 日常测试

日常代码修改不需要重新生成转换表。在仓库根目录运行：

```sh
go test ./ir/... ./storage/... ./encode-decode/... ./presets/block-converts/v1 ./presets/block-converts/specific-converts
go vet ./...
```

`go vet` 不执行生成入口。`go test ./...` 会包括 `scripts` 中写文件的准备步骤，不适合作为只读的日常验证命令。基础用法也保存在 `presets/block-converts/v1/example_test.go` 中，作为可执行的示例测试。
