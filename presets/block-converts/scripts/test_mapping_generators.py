"""Small conversion fixtures; never regenerate the repository's mapping files."""

from contextlib import chdir
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from amulet_nbt import IntTag, StringTag
import gen_fan_in_mapping as fan_in
import gen_fan_out_mapping as fan_out
from PyMCTranslate.py3.api import Block


class FakeTranslator:
    def __init__(self) -> None:
        self.converted = 0

    def base_names(self, namespace: str, *, force_blockstate: bool):
        return ["example"]

    def get_specification(self, namespace: str, name: str, *, force_blockstate: bool):
        return {
            "properties": {"a": ['"default"', '"other"'], "b": ["0", "1"]},
            "defaults": {"a": '"default"', "b": "0"},
        }

    def to_universal(self, block: Block):
        # An incomplete input here would permit the unwanted fallback path.
        if set(block.properties) != {"a", "b"}:
            raise AssertionError(f"incomplete translation input: {block}")
        self.converted += 1
        return block, None, False

    def from_universal(self, block: Block):
        return block, None, False


class FakeManager:
    def __init__(self, translator: FakeTranslator) -> None:
        self.block = translator

    def version_numbers(self, platform: str):
        return [(1, 0, 0)]

    def get_version(self, platform: str, version):
        return self


class MappingGeneratorsTest(unittest.TestCase):
    def test_partial_keys_always_have_complete_translation_values(self) -> None:
        for module in (fan_in, fan_out):
            with self.subTest(module=module.__name__):
                translator = FakeTranslator()
                with patch.object(module, "translations", FakeManager(translator)):
                    _, blocks = module.all_blocks("bedrock", (1, 0, 0))
                    seen_missing = False
                    count = 0
                    for partial, complete in blocks:
                        count += 1
                        self.assertEqual(set(complete.properties), {"a", "b"})
                        for key, value in partial.properties.items():
                            self.assertEqual(complete.properties[key], value)
                        if "a" not in partial.properties:
                            self.assertEqual(complete.properties["a"], StringTag("default"))
                            seen_missing = True
                        if "b" not in partial.properties:
                            self.assertEqual(complete.properties["b"], IntTag(0))
                            seen_missing = True
                    self.assertTrue(seen_missing)
                    self.assertEqual(count, 9)

    def test_fan_in_translates_complete_objects_and_keeps_partial_keys(self) -> None:
        translator = FakeTranslator()
        with tempfile.TemporaryDirectory() as folder:
            output = Path(folder) / "mapping.txt"
            with patch.object(fan_in, "translations", FakeManager(translator)):
                fan_in.gen_fan_in(
                    output_file=str(output),
                    platform="bedrock",
                    output_translator=translator,
                )
            rows = output.read_text().splitlines()
        self.assertEqual(translator.converted, 9)
        self.assertTrue(any(row.startswith("minecraft:example []#") for row in rows))

    def test_fan_out_translates_complete_objects_and_keeps_partial_keys(self) -> None:
        translator = FakeTranslator()
        with tempfile.TemporaryDirectory() as folder, chdir(folder):
            with patch.object(fan_out, "translations", FakeManager(translator)):
                fan_out.main()
            rows = Path("fan_out_java.mapping").read_text().splitlines()
            allowed = Path("fan_out_java_allowed.list").read_text().splitlines()
        self.assertEqual(translator.converted, 9)
        self.assertEqual(len(allowed), 9)
        self.assertTrue(any(row.startswith("minecraft:example []#") for row in rows))


if __name__ == "__main__":
    unittest.main()
