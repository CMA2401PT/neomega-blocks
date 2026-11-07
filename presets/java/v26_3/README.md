# Java 26.3 / Geyser Bedrock 1.26.60

调用 `Open()` 获取版本明确的只读 `catalog.Catalog`。包含 1,286 种方块、35,723 个真实完整状态及 Geyser 主方块映射；Java DataVersion 为 5023。

运行示例：`go run ./examples/java26`（在模块根目录）。

`Normalize` 只补官方默认值，拒绝非法状态；`JavaCandidates` 保留反向歧义。数据和转换不包含模型、贴图、方块实体或完整流体层处理。旧 `java/v1` 的显示兼容规则没有自动合并到本版本。

生成入口：

```powershell
go run ./third_party/gen_resource/geyser/cmd/generate
go run ./third_party/gen_resource/geyser/cmd/generate -check
```

默认读取模块外 `../../sources`。其他项目可传入 `-sources <目录>`、`-lock <锁文件>` 和 `-output <快照>`。仅使用嵌入数据时无需保留外部 sources。

输入锁：[26.3.lock.json](../../../third_party/gen_resource/geyser/versions/26.3.lock.json)。来源声明：[NOTICE](../../../third_party/geyser-notices/NOTICE.md)。
