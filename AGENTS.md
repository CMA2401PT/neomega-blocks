# 维护 neomega-blocks

本模块把不同来源的方块归到预置 NEMC/Bedrock 集合，并提供 Java 输出和方块身份查询。运行时消费嵌入资源，离线工具生成这些资源；面向使用者的 API 与约束见 [README.md](README.md)，工作区规则继承 [根 AGENTS.md](../AGENTS.md)。

这是方块语义基础库。模型、几何与渲染资源由 `neomega-viewer-core` 管理；本模块不得依赖 viewer、converter 或它们的 core。保留对 `neomega-nbt` 的显式 require/replace，使模块单独编译仍成立。

## 结构与变更边界

| 位置 | 维护职责 |
| --- | --- |
| `ir`、`encode-decode` | 方块文本、属性操作、NBT 事件与网络哈希。 |
| `storage` | 名称和数值映射的紧凑存储，不负责方块语义。 |
| `presets/block-converts/v1` | 目标 palette、默认属性、输入映射与 Java 输出。 |
| `presets/block-converts/specific-converts` | 运行时和生成端共用的特殊语义规则。 |
| `presets/java/v1` | Java 状态全集、兼容状态与属性查询。 |
| `third_party/gen_resource` | 数据来源、离线生成步骤和固定输入。 |
| `examples` | 可运行的使用教程，与 README 同步。 |

## 默认属性与旧属性的处理顺序

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

## 验收

日常代码和示例修改不需要安装 Python 或重新生成资源。在工作区根目录运行：

```sh
./check.sh neomega-blocks
```

在模块目录执行使用者示例：

```sh
go run ./examples/convert
```

检查入口执行依赖边界检查、`go vet`、`go test` 和脱离工作区编译。写资源的生成步骤使用 `resourcegen` 构建标签，普通测试不会执行它们。更新生成资源必须使用上文的生成入口，遵循 [生成工具说明](third_party/gen_resource/README.md) 的输入和验证要求，不要手工修补二进制产物。资源许可与 Required Notice 必须随资源保留。
