"""Export installed PyMCTranslate defaults for the Go resource builders."""

import json
from pathlib import Path

import PyMCTranslate
from target_versions import BEDROCK_VERSION


OUTPUT_DIR = Path(__file__).resolve().parent / "tmp" / "defaults"


def main() -> None:
    translations = PyMCTranslate.new_translation_manager()
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    for platform in ("bedrock", "java"):
        for version in translations.version_numbers(platform):
            translator = translations.get_version(platform, version).block
            defaults = {
                f"minecraft:{name}": translator.get_specification(
                    namespace="minecraft",
                    base_name=name,
                    force_blockstate=True,
                ).get("defaults", {})
                for name in translator.base_names(
                    namespace="minecraft",
                    force_blockstate=True,
                )
            }
            version_name = "_".join(str(part) for part in version)
            output = OUTPUT_DIR / f"{platform}_{version_name}.json"
            output.write_text(json.dumps(defaults, sort_keys=True) + "\n")
            if (platform, version) == ("bedrock", BEDROCK_VERSION):
                (OUTPUT_DIR.parent / "bedrock_defaults.json").write_text(
                    json.dumps(defaults, sort_keys=True) + "\n"
                )


if __name__ == "__main__":
    main()
