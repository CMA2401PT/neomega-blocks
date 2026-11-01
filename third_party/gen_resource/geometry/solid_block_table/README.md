# 生成 Java 几何、元数据和实心表

这个生成器为同一套 Java 状态建立静态几何与实心判定。它读取本目录中的模型和图集，使用 `geometry/java_block_model` 选择模型和烘焙形状，再由 `geometry` 计算边界覆盖，将三份产物直接写到对应的 preset 包，供 `go:embed` 使用。

输入文件由维护者从一套匹配的 `neomega-assets/dist/final` 手动复制到本目录。`go generate` 只读取这些本地副本，不会从相邻仓库自动更新输入。复制来的文件被 Git 忽略；重新取得仓库后，需要先准备输入才能生成。

## 准备匹配的输入

Java 状态来自 `presets/java/v1`，资源版本由 `build-info.json` 提供。生成前会检查两个版本相同。图集、坐标和模型必须来自同一次资源构建。资源定义和读取代码属于 blocks，assets 只作为文件生产者；生成器没有对 assets Go 模块的依赖。`blocksAtlases.json` 的纹理键必须使用完整资源 ID，例如 `minecraft:block/stone`。

旧版输入通过 `JavaCompatibility` 连接到目标状态。例如 `minecraft:grass []` 直接共享 `minecraft:short_grass []` 的形状、染色和实心判定。兼容状态不需要另配旧版本资源。更新兼容输入时，先按[转换资源生成说明](../../README.md)重新生成映射，再运行这里的建表命令。

本目录需要以下六个文件：

| 输入 | 用途 |
| --- | --- |
| `blockstates.json` | 根据属性选择 variant 和 multipart 部件。 |
| `models.json` | 解析父模型、元素、面和纹理槽。 |
| `blocksAtlases.json` | 查找贴图区域及动画帧信息。 |
| `blocksAtlasLatest.png` | 读取像素，区分不透明、镂空和半透明纹理。 |
| `blockRenderers.json` | 识别普通模型、特殊模型、流体、不可见和动态方块。 |
| `build-info.json` | 检查资源的 Java 目标版本。 |

例如，在工作区根目录手动复制：

```sh
cp neomega-assets/dist/final/blockstates.json neomega-blocks/third_party/gen_resource/geometry/solid_block_table/
cp neomega-assets/dist/final/models.json neomega-blocks/third_party/gen_resource/geometry/solid_block_table/
cp neomega-assets/dist/final/blocksAtlases.json neomega-blocks/third_party/gen_resource/geometry/solid_block_table/
cp neomega-assets/dist/final/blocksAtlasLatest.png neomega-blocks/third_party/gen_resource/geometry/solid_block_table/
cp neomega-assets/dist/final/blockRenderers.json neomega-blocks/third_party/gen_resource/geometry/solid_block_table/
cp neomega-assets/dist/final/build-info.json neomega-blocks/third_party/gen_resource/geometry/solid_block_table/
```

输入素材的来源与声明随 `neomega-assets` 的交付记录维护，见 [neomega-assets/NOTICE.md](../../../../../neomega-assets/NOTICE.md)。

## 生成和验证

在 `neomega-blocks` 根目录运行：

```sh
go generate ./presets/java/v1/geometry
```

指令从生成器源码所在目录读取输入，直接更新：

| 输出路径，均相对于模块根目录 | 消费者 |
| --- | --- |
| `presets/java/v1/geometry/geometry.nvtables` | viewer 查询状态对应的共享形状、边界覆盖及流体满高度形状。 |
| `presets/java/v1/geometry/geometry.nvtables.java.json` | 版本、渲染类别与跳过原因；跳过状态直接记录 tidy IRBlock。 |
| `presets/java/v1/solidtable/geometry.nvsolid` | converter/fixer 判断邻居六个面是否被不透明矩形完整覆盖。 |

产物无需手动复制。实心表与几何表由同一个入口生成，实心表包不再单独声明第二个 `go:generate`。

需要检查真实输入与入库、嵌入和编解码结果是否一致时，运行：

```sh
NEOMEGA_ASSETS_DIR="$PWD/third_party/gen_resource/geometry/solid_block_table" \
  go test -v -count=1 ./third_party/gen_resource/geometry/solid_block_table/internal/tablegen
```

这项检查会重新建表，并逐项比较状态索引、全部四边形、边界覆盖、满高度映射、元数据和实心表。日常检查无需准备这六份输入，可在工作区根目录运行 `./check.sh neomega-blocks neomega-viewer neomega-converter neomega-assets`。

## 生成策略放在哪里

`main.go` 负责读取路径并写出三份资源。`internal/tablegen` 直接枚举 `java/v1.JavaFullStates` 的 tidy IR 方块，组合模型解析和烘焙步骤、共享相同形状并建立索引，然后为 `JavaCompatibility` 的输入增加指向目标形状的索引及实心位。`internal/resources` 保存具体 Minecraft 预览规则：固定染色、玻璃与流体透明度、干燥默认属性和静态流体模型。

通用 `Select` 保留候选及权重，建表策略固定选择每组首项，所有命中的 multipart 部件都会保留。计算实心位之前先合并同一状态的全部部件，保证多部件共同覆盖整面时能够识别。动态的 `moving_piston` 状态会记录跳过原因。

中间诊断文件如有需要应写在本目录的 `tmp/`，该目录被 Git 忽略。最终预置文件只由上述生成入口更新。
