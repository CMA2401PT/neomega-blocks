"""NEMC specification additions to the installed upstream database."""

from typing import Any

from PyMCTranslate import BlockTranslator


def get_specification(
    *,
    translator: BlockTranslator,
    platform: str,
    version: tuple[int, ...],
    name: str,
) -> dict[str, Any]:
    specification = translator.get_specification(
        namespace="minecraft",
        base_name=name,
        force_blockstate=True,
    )
    # NEMC adds this byte to these blocks. Keep it in enumeration and defaults
    # without modifying PyMCTranslate's installation or its shared database.
    if (
        platform == "bedrock"
        and version == (1, 21, 50)
        and name in {"crafting_table", "furnace", "lit_furnace"}
    ):
        specification["properties"] = {
            "custom_appearance": ["0b", "1b"],
            **specification.get("properties", {}),
        }
        specification["defaults"] = {
            "custom_appearance": "0b",
            **specification.get("defaults", {}),
        }
    return specification
