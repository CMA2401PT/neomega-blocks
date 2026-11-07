# neomega-blocks

`neomega-blocks` 用来描述、查找和转换 Minecraft 方块。不同版本可能使用不同的方块名、属性名和属性值；旧格式还可能只给出数字 ID 与附加值。这个库把这些输入转换到一套确定的国际版 Bedrock 方块集合，并提供向 Java 方块的输出转换。

项目分为两个相互配合的部分：**运行时读取已经准备好的转换表；离线工具负责生成、修正和验证这些表。** `ir`、`storage` 和 `encode-decode` 提供通用逻辑；`presets` 提供选定版本的数据与规则；`third_party` 保存离线生成工具和来源声明。使用现成 preset 时只需调用 Go 接口。各包能力和 IRBlock 的 Tidy 约定见 [presets/README.md](presets/README.md)。

模型、几何和渲染资源由 [neomega-viewer-core](../neomega-viewer-core/README.md) 管理。

转换中心由 PyMCTranslate 的最新完整 Java 与国际版 Bedrock 组成，当前分别为 Java 26.3 和 Bedrock 26.60。`presets/java/v1` 提供同一 Java 目标的完整单层状态、属性域、默认值与历史兼容映射。`v1` 表示预置数据格式的修订号。

## 为什么要有一个转换中心

如果让每个版本都直接转换到其他所有版本，转换关系会越来越多。这里先把输入归入当前 Bedrock 状态全集，再根据需要转出到当前 Java。PyMCTranslate 的 Universal 表示负责连接各版本的语义，复杂转换提前生成为表。

```text
其他版本的 Bedrock ─┐
Java 方块 ─────────┼─→ Bedrock 26.60 全集 ⇄ Java 26.3
旧 Schematic ID ───┘
```

两侧状态都由 PyMCTranslate specification 的属性域做笛卡尔积枚举，也就是取遍每个属性的所有组合。Java 单层状态去掉 `waterlogged`，含水由独立层表达。`block-converts/v1.JavaTargetVersion` 引用 `java/v1.JavaTargetVersion`；`BedrockTargetVersion` 公布 Bedrock 目标版本。

历史 Java 完整状态直接经 PyMCTranslate 解释到 Java 目标，`java/v1.NormalizeSingleLayer` 提供这张兼容表的查询。多个版本共享同一输入时，采用最高版本的解释。转换不保证可逆，因为不同版本能表达的信息不同，部分转换还需要邻居或方块实体上下文。

历史 `block_palette.json` 只用于 legacy data value 查询。data value 是旧物品 aux 使用的方块变体编号；索引保留历史状态，供调用方再转换到当前中心。最新状态的网络哈希仍由名称和带类型的状态计算，因此相同写法的状态与历史 palette 保持相同哈希。

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

## 运行第一个例子

模块路径为 `neomega-blocks`，要求 Go 1.25.5。在克隆后的本模块目录运行：

```sh
go run ./examples/convert
```

[`examples/convert/main.go`](examples/convert/main.go) 把缺少部分属性的 Java 栅栏门转为目标 Bedrock 方块，通过网络哈希找回它，最后转换一个旧 Schematic 数字方块。运行时只需要 Go，转换表已经嵌入模块。命令输出：

```text
minecraft:jungle_fence_gate ["in_wall_bit"=false,"minecraft:cardinal_direction"="north","open_bit"=true]
哈希反查一致：true
minecraft:brown_concrete_powder []
```

外部输入先调用 `Tidy()`，使名称和属性顺序符合哈希查询约定。`Tidy` 不会补游戏默认属性，缺参由转换入口内的 Canon 处理。具体约定见 [presets/README.md](presets/README.md)。

在工作区以外的项目使用本地源码时，需要同时取得 `neomega-blocks` 和 `neomega-nbt`；调用方的 `go.mod` 应 require `neomega-blocks v0.0.0`，并分别 replace 这两个模块到本地目录。`neomega-blocks` 自身已声明对 NBT 的依赖。

需要自己构造属性时，可以使用 `ir.NewIRBlockFromGo(name, props)`，其中 `props` 为 `map[string]any`。需要修改已有属性时，`block.States().Set(...)`、`Delete(...)`、`Update(...)` 会返回新的 states；将它与原名称组合后得到新方块。

## 根据输入选择入口

不同输入携带的信息不同，所以不必先把所有数据都转成字符串。preset 提供这些入口：

| 已有数据或目标 | 入口 |
| --- | --- |
| Bedrock 方块名与状态 | `ConvertFromBedrock(IRBlock)` |
| Java 方块名与状态 | `ConvertFromJava(IRBlock)` |
| 旧 Schematic 数字 ID、meta → Bedrock | `ConvertFromSchematic(id, meta)` |
| 旧 Schematic 数字 ID、meta → Java | `SchematicToJava(id, meta)`，直接使用 Java 表，避免经过 Bedrock 损失状态 |
| 目标 Bedrock 方块转 Java | `ConvertToJava(IRBlock)` |
| 已知网络哈希，取目标方块 | `BedrockHashToBlockStrict(hash)` |
| 遍历目标集合中的网络哈希与方块 | `BedrockBlocks()`，返回只读迭代器，顺序不固定 |
| 名称与 legacy data value，取哈希或方块 | `BedrockDataValueToHash(name, value)` / `BedrockDataValueToBlock(name, value)` |
| 历史网络哈希反查名称与 data value | `BedrockHashToBlockData(hash)` |

读取返回的 `bool` 判断是否找到转换或数据。`ConvertFrom*` 失败时返回 `minecraft:unknown []` 和 `false`；不要只检查返回的方块文本。`ConvertToJava` 面向 preset 的目标 Bedrock 表示，处理其他版本时可以先调用 `ConvertFromBedrock`。

这些接口没有源版本参数。多个版本中相同文本的解释由生成表时的优先级决定；Java 默认属性按名称合并，高版本优先。因此，它是一套预先选定策略的兼容转换，不是可在运行时任选版本的翻译器。

## 其他入口与维护资料

需要 Java 目标状态全集、属性默认值或旧状态的显示兼容映射时，使用 [`presets/java/v1`](presets/java/v1)。这些接口与方块转换 preset 的用途不同，状态全集、含水属性和兼容映射的约定见 [presets/README.md](presets/README.md)。

需要自定义紧凑查询表时，可以单独使用 `storage/blocks-name` 或 `storage/num_lookup`。公共 preset 不提供运行时注册接口。

分层边界、转换顺序、资源来源与更新命令见 [AGENTS.md](AGENTS.md)。离线生成器的完整输入与产物关系见 [生成工具说明](third_party/gen_resource/README.md)，许可声明见 [NOTICE.md](third_party/pymctranslate-notices/NOTICE.md)。
