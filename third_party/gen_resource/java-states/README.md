# Java 状态资源

这一步为浏览、渲染和 fixer 生成当前 Java 目标版本的完整单层状态、属性域、默认值与历史兼容映射。目标由 `presets/java/v1.JavaTargetVersion` 指定，当前为 26.3。包名中的 `v1` 表示预置数据格式修订号。

转换中心的 Java 侧保持 1.21.4，因为项目将 PyMCTranslate（Amulet）的 Java 数据使用范围固定到这里，以尊重 Amulet 的许可；converter 也按此版本写存档。状态知识需要认识当前目标存档，所以使用 Mojang 官方报告，并通过迁移表连接中心状态。

## 状态如何连接

Mojang 服务端的数据生成器提供所有注册状态、属性域和默认状态。生成器按注册 ID 枚举，去掉 `waterlogged` 后 Tidy 并去重。Tidy 是整理方块名称与属性顺序的操作。含水由独立层表达，因此属性域也去掉 `waterlogged`；属性值保持报告次序，方块和属性键分别排序以稳定差异。

历史状态按以下顺序解释：

```text
历史完整状态 → Amulet 兼容表 → 中心 Java 1.21.4 状态 → 迁移表 → 当前目标状态
```

Amulet 兼容表来自 `block-converts/tmp/hub_java_compatibility.list`，完整中心状态来自同目录的 `hub_java_states.list`。迁移规则位于 `migrate.go`，先替换名称与旧属性，再用目标默认状态补缺项。`chain` 改名为 `iron_chain`；`creaking_heart` 的 `active=true` 对应 `awake`，`active=false` 对应 `uprooted`。迁移结果必须存在于目标全集，错误会指出需要补规则的源状态。

目标全集中的源状态保留自身身份。其余中心状态以及中心兼容表中的历史源状态写入目标兼容表。运行时通过 `java/v1.NormalizeSingleLayer` 查询解释结果，转换中心写出的 1.21.4 状态因此也能用于当前目标的浏览和修复。

## 准备官方报告

在 `neomega-blocks` 根目录运行以下命令，需要 Python 3.11 和目标游戏版本支持的 Java 运行时：

```sh
.venv/bin/python third_party/gen_resource/java-states/prepare-report.py --version 26.3
```

`prepare-report.py` 从 Mojang 官方 version manifest 找到指定版本，下载版本元数据与 server.jar，再运行服务端内的数据生成器。`--java /path/to/java` 可指定 Java 可执行文件。报告保存为 `tmp/blocks.json`，服务端嵌套 jar 中的游戏版本保存为 `tmp/version.json`。生成步骤要求其中的 `id` 与目标常量一致。

## 生成运行时资源

先准备 Python 环境和中心中间结果，在模块根目录执行：

```sh
uv sync --locked
.venv/bin/python third_party/gen_resource/block-converts/gen_fan_out_mapping.py
.venv/bin/python third_party/gen_resource/block-converts/gen_java_compat.py
```

然后运行写资源的 Go 步骤：

```sh
cd third_party/gen_resource/java-states
go test -tags resourcegen -v -count=1 -run '^TestGenerateJavaStates$' .
```

输入为 Mojang 26.3 报告和上述两个 Python 脚本生成的中心 1.21.4 中间结果时，该命令输出 1,286 种方块、22,398 个单层状态和 2,077 条兼容映射。产物为 `presets/java/v1/java_states.list`、`java_block_schema.list` 和 `java_compatibility.list`。

同一步还写出 `tmp/hub_to_target.mapping`，含 15 条真正变化的中心状态。`block-converts` 的 Java 转入生成步骤使用它，为旧状态的转入键追加等价的当前目标形式，并保留原键的含水属性。完整生成入口 `block-converts/generate.sh` 会在 Python 步骤之后、Go 转换表步骤之前执行本步骤。

## 更新目标版本

准备新版本的官方报告，修改 `java/v1.JavaTargetVersion`，并在 `migrate.go` 补充相对中心状态的改名或属性改写规则，再运行完整生成入口。生成错误给出不属于目标全集的迁移结果，用它定位缺失规则。转换中心的 Java 侧继续保持 1.21.4。

报告解析和迁移逻辑使用内存样例测试，日常验证在模块根目录执行 `go test ./third_party/gen_resource/java-states`。完整验收见 [生成工具说明](../README.md)。
