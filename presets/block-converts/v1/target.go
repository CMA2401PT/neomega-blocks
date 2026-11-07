package v1

// JavaTargetVersion is the Java state schema produced by this preset.
// Consumers should load a matching resource bundle.
// 转换中心将 Amulet 的 Java 数据使用范围固定到 1.21.4，converter 按此版本写存档。
// 浏览、渲染或 fixer 需要当前目标状态时，调用 java/v1.NormalizeSingleLayer。
const JavaTargetVersion = "1.21.4"

// BedrockTargetVersion is the translation reference. The NEMC palette defines
// the actual set of block hashes, including its platform-specific extensions.
const BedrockTargetVersion = "1.21.50"
