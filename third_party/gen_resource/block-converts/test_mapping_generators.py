"""Small conversion fixtures; never regenerate the repository's mapping files."""

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

    def get_specification(self, namespace: str, base_name: str, *, force_blockstate: bool):
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
        with tempfile.TemporaryDirectory() as folder:
            with patch.object(fan_out, "translations", FakeManager(translator)):
                fan_out.main(
                    tmp_dir=Path(folder) / "tmp",
                    java_dir=Path(folder) / "java",
                )
            rows = (Path(folder) / "tmp" / "fan_out_java.mapping").read_text().splitlines()
            allowed = (Path(folder) / "tmp" / "fan_out_java_allowed.list").read_text().splitlines()
        self.assertEqual(translator.converted, 9)
        self.assertEqual(len(allowed), 9)
        self.assertTrue(any(row.startswith("minecraft:example []#") for row in rows))


class JavaCompatibilityTest(unittest.TestCase):
    def test_newest_interpretation_wins_after_removing_waterlogged(self) -> None:
        import gen_java_compat as compat

        class Translator:
            def __init__(self, result: str):
                self.result = result
                self.calls = 0

            def to_universal(self, block, **kwargs):
                self.calls += 1
                self.assert_dry = "waterlogged" not in block.properties
                return Block("minecraft", self.result), None, False

            def from_universal(self, block, *args, **kwargs):
                return block, None, False

            def get_specification(self, **kwargs):
                return {"defaults": {}}

        high, low = Translator("stone"), Translator("dirt")
        target = Translator("stone")

        class Manager:
            def version_numbers(self, platform):
                return [(1, 12, 2), compat.TARGET_VERSION, (1, 20, 1)]

            def get_version(self, platform, version):
                class Version: pass
                result = Version()
                result.block = target if version == compat.TARGET_VERSION else high if version == (1, 20, 1) else low
                return result

        def states(platform, version):
            if version == compat.TARGET_VERSION:
                yield Block("minecraft", "stone")
                yield Block("minecraft", "dirt")
            else:
                yield Block("minecraft", "legacy", {"waterlogged": StringTag("true")})
                yield Block("minecraft", "legacy", {"waterlogged": StringTag("false")})

        with tempfile.TemporaryDirectory() as folder:
            output = Path(folder) / "mapping.list"
            with patch.object(compat, "translations", Manager()), patch.object(compat, "iter_full_blocks", states):
                compat.generate(output=output, report_path=Path(folder) / "report.json")
            self.assertEqual(output.read_text(), "minecraft:legacy []#minecraft:stone []\n")
        self.assertEqual(high.calls, 1)
        self.assertTrue(high.assert_dry)
        self.assertEqual(low.calls, 0)

    def test_real_grass_and_wall_translate_to_target_java(self) -> None:
        import gen_java_compat as compat
        target = compat.translations.get_version("java", compat.TARGET_VERSION).block
        samples = [
            ((1, 20, 1), Block("minecraft", "grass"), "minecraft:short_grass []"),
            ((1, 15, 2), Block("minecraft", "cobblestone_wall", {
                "north": StringTag("true"), "east": StringTag("false"),
                "south": StringTag("false"), "west": StringTag("false"), "up": StringTag("true"),
            }), 'minecraft:cobblestone_wall ["east"="none","north"="low","south"="none","up"="true","west"="none"]'),
        ]
        for version, block, expected in samples:
            source = compat.translations.get_version("java", version).block
            actual, _ = compat.translate_state(source=source, target=target, block=block, native={expected})
            self.assertEqual(actual, expected)


if __name__ == "__main__":
    unittest.main()
