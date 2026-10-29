package v1

import java "neomega-blocks/presets/java/v1"

// JavaTargetVersion is the Java state schema produced by this preset.
// Consumers should load a matching resource bundle.
const JavaTargetVersion = java.JavaTargetVersion

// BedrockTargetVersion is the translation reference. The NEMC palette defines
// the actual set of block hashes, including its platform-specific extensions.
const BedrockTargetVersion = "1.21.50"
