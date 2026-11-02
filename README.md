# neomega-blocks

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

### 缺少属性为什么仍然可以转换

`block[a=opt1,b=opt2]` 省略一个属性后成为 `block[a=opt1]`，仍然是需要支持的合法输入。缺少的属性应使用对应默认值，而不是让转换器自行猜测。

PyMCTranslate 对缺参输入的处理并不可靠，可能 fallback 到错误方块。因此，缺参支持依靠默认属性归一化：

```text
合法的缺参方块 → Canon 补齐默认属性 → 完整状态的转换规则或映射
```

Python 脚本保留的 `allow_missing` 参数没有行为作用，不应重新启用“把缺参对象直接交给 PyMCTranslate”的分支。脚本枚举过程中即使保留缺参形式作为输入键，实际交给 `to_universal` 的也是补齐后的 `block_with_default`。运行时通过 Canon 尽量复用完整状态的映射。

Canon 只适合处理缺参。它也会删掉目标版本没有的键，而旧版本的键往往带着含义。
例如 `stone ["stone_type"="granite"]` 删掉 `stone_type` 后成了 `stone`，旧的 `top_slot_bit=true` 删掉后成了下半砖。
因此生成转入表时，Canon 删掉了输入的键、且结果与映射不同的那一对会写进转入表；只补了默认值的输入不写，继续交给 Canon。
运行时也要先查转入表，再用 Canon，否则这些条目永远轮不到。

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

- Bedrock 输入先尝试目标表直接命中，再用原输入查转入表，最后尝试 Canon 后命中目标表。
- Java 输入优先处理墙、悬挂告示牌等特殊规则，再用原输入查转入表，最后用 Canon 后的状态查表。
- 向 Java 输出时，使用特殊规则或目标 Bedrock 哈希到 Java IR 的映射。

为了减少大量状态文本的重复，`storage/blocks-name` 把方块名称和 states 分开保存。states 使用共享字符串片段的紧凑前缀树，哈希映射到重建索引。`storage/num_lookup` 用有序数组保存静态数字映射，并允许通过动态层覆盖或追加数据；`Frozen` 将动态层并入静态层，`Dump` 导出合并后的快照。

方块哈希也复用了 `neomega-nbt`：将 `name` 与 `states` 表达为 NBT 操作流，以小端二进制形式送入 FNV-1a 32 位写入器。它不是对 IR 文本直接求哈希。调用 `Hash` 前应使用整理后的名称与有序状态；`Hash` 本身不负责补齐默认属性。

表加载在地址满足对齐要求时直接引用输入字节；不满足时复制到对齐存储。加载器可能共享输入内存，因此调用者应避免修改输入字节。静态表与前缀树的 dump 含 Go 数字或节点的原生内存布局，生成和使用时应保持兼容的布局与字节序。

## 转换表从哪里来

离线生成工具保存在 [`third_party/gen_resource`](third_party/gen_resource/README.md)。Python 借助 uv 安装的 PyMCTranslate 生成候选映射，Go 再将它们修正到 NEMC palette 的目标集合，并压缩为运行时表。固定输入跟随脚本保存，中间文件写入各脚本目录下被 Git 忽略的 `tmp/`，最终资源直接写入 `presets` 中消费它们的 Go 包。

更新时，在模块根目录执行：

```sh
sh third_party/gen_resource/block-converts/generate.sh
```

该入口根据根目录的 `pyproject.toml` 和 `uv.lock` 在 `.venv/` 初始化 Python 3.11 环境，并依次生成默认属性、候选映射、运行时转换表和 Java 状态定义。源码树不再保存 PyMCTranslate；NEMC 的 `custom_appearance` 扩展由生成端的 `pymc_specification.py` 显式补充。

PyMCTranslate 的来源声明与 Required Notice 见 [third_party/pymctranslate-notices/NOTICE.md](third_party/pymctranslate-notices/NOTICE.md)。输入来源、每张表的生成步骤、更新版本和单步验证方式见[生成工具说明](third_party/gen_resource/README.md)。

## 怎样扩展

新增需求首先要判断：变化来自输入形式、单个转换规则，还是整个目标集合。把修改放在对应层，才能避免为一个方块改动全部数据流程。

**接入新的输入形式**时，先将名称与属性构造成 IR，再接入现成 preset。需要直接处理 NBT 操作流时，`IRBlock.DecodeToHL/DecodeToLL` 可以向 `neomega-nbt` 接收端输出；`NewIRBlockEncoder` 可以接收包含 `name` 和 `states` 的低层 NBT 操作，重建 IR。

**修正某类方块的语义转换**时，参考 `specific-converts` 中的墙和悬挂告示牌规则。接入运行时入口的同时，也要让生成端识别同一规则，避免把同类转换再次写入表中。测试应覆盖显式属性、缺参补默认值，以及确实可逆的状态往返。

**增加表驱动映射或更新目标版本**时，沿生成工具说明中的准备流程修改数据来源、默认属性和 `outputFix`，重新生成受影响的产物。扩展旧 Schematic 输入则从 `third_party/gen_resource/block-converts/schematic.json` 开始，并使用已有 Java 转入逻辑生成最终索引。

若只需要自定义查询表，可以单独使用 `storage/blocks-name` 或 `storage/num_lookup`；preset 的全局表没有提供运行时注册接口。

## 日常测试

日常代码修改无需安装 Python 或重新生成资源。在工作区根目录运行：

```sh
./check.sh neomega-blocks
```

检查会执行 `go vet`、`go test`，并验证模块脱离工作区仍可编译。写资源的准备步骤使用 `resourcegen` 构建标签，普通 `go test ./...` 不会执行它们。基础用法保存在 `presets/block-converts/v1/example_test.go`，作为可执行的示例测试。
