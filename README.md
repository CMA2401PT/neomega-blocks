# neomega-blocks

本地研究扩展：新增独立的 [Java 26.3 预设](presets/java/v26_3/README.md)，使用官方 Java 状态报告与 Geyser 映射。调用 `v26_3.Open()` 获取明确版本的目录；`go run ./examples/java26` 可验证。原 `v1` 数据、目标版本和下文所述行为保持原样。新流程见 `third_party/gen_resource/geyser`，来源声明见 `third_party/geyser-notices/NOTICE.md`。

`neomega-blocks` 用来描述、查找和转换 Minecraft 方块。不同版本可能使用不同的方块名、属性名和属性值；旧格式还可能只给出数字 ID 与附加值。这个库把这些输入转换到一套确定的 NEMC/Bedrock 方块集合，并提供向 Java 方块的输出转换。

项目分为两个相互配合的部分：**运行时读取已经准备好的转换表；离线工具负责生成、修正和验证这些表。** `ir`、`storage` 和 `encode-decode` 提供通用逻辑；`presets` 提供选定版本的数据与规则；`third_party` 保存离线生成工具和来源声明。使用现成 preset 时只需调用 Go 接口。各包能力和 IRBlock 的 Tidy 约定见 [presets/README.md](presets/README.md)。

模型、几何和渲染资源由 [neomega-viewer-core](../neomega-viewer-core/README.md) 管理。

直接浏览 Java 存档时，`presets/java/v1` 另提供旧 Java 状态到 1.21.4 的兼容映射。它由 PyMCTranslate 的 Universal 表示连接源状态与目标状态，高版本解释优先，运行时不带源版本。几何生成器让兼容输入共享目标状态的形状；含水仍由 viewer 单独拆层。使用方式见 [Java preset 说明](presets/README.md)。

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

## 运行第一个例子

模块路径为 `neomega-blocks`，要求 Go 1.25.5。在克隆后的本模块目录运行：

```sh
go run ./examples/convert
```

[`examples/convert/main.go`](examples/convert/main.go) 把缺少部分属性的 Java 栅栏门转为目标 Bedrock 方块，通过网络哈希找回它，最后转换一个旧 Schematic 数字方块。运行时只需要 Go，转换表已经嵌入模块。命令输出：

```text
minecraft:jungle_fence_gate ["direction"=2,"in_wall_bit"=false,"open_bit"=true]
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
| 网络哈希反查名称与 data value | `BedrockHashToBlockData(hash)` |

读取返回的 `bool` 判断是否找到转换或数据。`ConvertFrom*` 失败时返回 `minecraft:unknown []` 和 `false`；不要只检查返回的方块文本。`ConvertToJava` 面向 preset 的目标 Bedrock 表示，处理其他版本时可以先调用 `ConvertFromBedrock`。

这些接口没有源版本参数。多个版本中相同文本的解释由生成表时的优先级决定；Java 默认属性按名称合并，高版本优先。因此，它是一套预先选定策略的兼容转换，不是可在运行时任选版本的翻译器。

## 其他入口与维护资料

需要 Java 目标状态全集、属性默认值或旧状态的显示兼容映射时，使用 [`presets/java/v1`](presets/java/v1)。这些接口与方块转换 preset 的用途不同，状态全集、含水属性和兼容映射的约定见 [presets/README.md](presets/README.md)。

需要自定义紧凑查询表时，可以单独使用 `storage/blocks-name` 或 `storage/num_lookup`。公共 preset 不提供运行时注册接口。

分层边界、转换顺序、资源来源与更新命令见 [AGENTS.md](AGENTS.md)。离线生成器的完整输入与产物关系见 [生成工具说明](third_party/gen_resource/README.md)，许可声明见 [NOTICE.md](third_party/pymctranslate-notices/NOTICE.md)。
