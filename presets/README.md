# 预置方块能力

`presets` 把选定 Minecraft 版本的数据和转换规则封装成可直接使用的 Go 包。资源文件与使用它们的 Go 文件放在一起，通过 `go:embed` 编入程序。运行时直接读取这些资源。

需要更新这些资源时，使用 [离线生成工具](../third_party/gen_resource/README.md)。生成脚本及其输入保存在 `third_party/gen_resource`，中间文件保存在对应脚本目录的 `tmp/`。

## 按用途选择包

| 包 | 提供的能力 |
| --- | --- |
| [`block-converts/v1`](block-converts/v1) | 以 PyMCTranslate 的完整国际版 Bedrock 状态为目标集合，将多版本 Bedrock、Java 和旧 Schematic ID/meta 转入；将目标方块转出为 Java 26.3。另提供网络哈希、方块文本和 legacy data value 的查询。Bedrock 目标版本为 26.60。 |
| [`block-converts/specific-converts`](block-converts/specific-converts) | 墙与悬挂告示牌的 Java/Bedrock 专用转换规则。`v1` 与生成端共用这些规则。墙的函数用于 `IRStates.Update`；悬挂告示牌函数处理整个方块。旁边的 JSON 是说明状态差异的样例。 |
| [`java/v1`](java/v1) | Java 26.3 状态枚举、旧状态到目标版本的兼容映射、方块状态与哈希互查，以及目标版本的属性默认值和取值域查询。哈希反查还收录附带 `waterlogged="true"/"false"` 的状态，供转换器和状态修复使用。 |

两个包的 `v1` 都是预置数据格式的修订号。转换中心与 Java 状态知识共用 `java/v1.JavaTargetVersion`，当前为 26.3；Bedrock 目标常量为 26.60。两侧状态与转换都来自 PyMCTranslate 1.2.53。

`java/v1.JavaFullStates` 枚举目标 specification 的完整属性组合，去掉 `waterlogged` 并去重。含水交给独立层表达，因此状态全集和 `BlockProperties` 都不含 `waterlogged`；哈希反查中的含水变体由 Go 加入。

为了理解历史名称和属性，`JavaCompatibility` 提供历史完整状态到当前目标状态的映射，`JavaCompatibleStates` 枚举目标全集与这些额外输入。生成器从高到低枚举源版本，经 Universal 直接转到目标 Java 并补齐默认值。目标全集中的状态保留自身身份，运行时不需要源版本。

Bedrock 全集同样枚举 specification 的完整属性组合，并保留 byte、int、string 的 NBT 类型，网络哈希由这些带类型的状态计算。`BedrockHashToBlockStrict` 和 `BedrockBlocks` 查询最新全集。`BedrockDataValueToHash`、`BedrockDataValueToBlock` 与 `BedrockHashToBlockData` 查询独立保留的历史 palette 索引，方便旧物品 aux 先还原历史方块，再转入中心。

需要取得兼容状态的目标解释时，调用 `NormalizeSingleLayer`。它忽略 `waterlogged`，查询完整单层状态，返回目标状态与是否命中。几何表已经为兼容输入建立共享形状索引，因此 viewer 拆层后可直接用原始状态的 hash 查几何。`HashToBlock` 反查仍保留原始输入身份及其含水属性，便于诊断与拆层。

旧 schematic 转 Java 存档时，使用 `block-converts/v1.SchematicToJava(id, meta)`。它直接查询生成的 Java 26.3 映射，保留 Java 独有属性；`ConvertFromSchematic` 则面向 Bedrock。两者共用 `schematic.json` 来源，未知组合通过 bool 返回失败。旧绊线已经废弃的 suspended 位不改变现代 Java 状态，其余位仍按原值映射。

## IRBlock 的整理约定

`IRBlock` 是“方块名 + 属性”的文本表示。`Tidy` 会补默认命名空间并整理空白、属性顺序；`Canon` 进一步依据预置默认属性补缺项、去掉多余项。完成 Tidy 不等于属性齐全，也不等于方块属于目标集合。

使用 `block-converts/v1` 时，调用方应先对外部输入执行 `Tidy()`。转换入口会在部分路径执行 Canon，但直接命中和回退路径仍会使用输入哈希，不能依赖它们替调用方整理所有输入。成功返回的方块已经 tidy，可直接比较或取哈希。Java 输出直接属于目标单层状态全集，含水属性另结合源数据处理。

`BedrockCanon` 和 `JavaCanon` 自己会先执行 Tidy；即使返回 `false`，返回值也已经过 Tidy，只是没有找到对应默认属性。`BedrockHashToBlockStrict`、`BedrockDataValueToBlock` 和 `BedrockBlocks` 返回的预置方块已经 tidy。转换失败时应检查 `bool`，不要将失败返回值当作有效方块。

`specific-converts` 是底层规则。调用前先整理输入，调用方负责默认属性补齐与目标集合检查；通常直接使用 `v1` 更方便。悬挂告示牌转换成功时返回已经 tidy 的方块；失败时原样返回输入。墙的 replacer 交给 `IRStates.Update` 后得到有序属性，组合方块名时仍由调用方保证名称规范。

`java/v1.BlockToHash` 内部会 Tidy，调用方无需重复整理。`HashToBlock` 和 `JavaFullStates` 返回的方块已经 tidy。`BlockProperties` 返回的是共享的只读属性表，默认值保留 SNBT 引号，便于直接构造 `IRStates`；构造完仍应 Tidy 后再直接调用 `IRBlock.Hash()`。

几何表的 hash 来自 `java/v1.JavaCompatibleStates` 返回的 tidy IR 方块。身份反查直接使用 `java/v1.HashToBlock`。几何和实心查询直接接收这些数值 hash，返回静态形状或六方向位域。几何元数据中的 `SkippedState.Block` 也是 tidy IR 方块。原始模型和图集来自 viewer-core 的 viewer-assets，来源见 [资源声明](../../neomega-viewer-core/viewer-assets/NOTICE.md)。

## 更新嵌入资源

从 `neomega-blocks` 根目录执行：

```sh
sh third_party/gen_resource/block-converts/generate.sh
```

入口按 Python 状态与候选映射、Go 转换表的顺序更新两个 `v1` 包的资源。每处 `go:embed` 的注释标明了对应的生成步骤。具体输入、产物关系、单步命令和验证方式见[生成工具说明](../third_party/gen_resource/README.md)。

渲染几何与实心表由 [viewer-core 的建表器](../../neomega-viewer-core/viewer-assets/cmd/build-tables/README.md) 生成与分发。
