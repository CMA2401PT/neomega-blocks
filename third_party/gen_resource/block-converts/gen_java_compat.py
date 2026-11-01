"""Generate best-effort, versionless Java -> target Java state compatibility."""

from pathlib import Path
import json

from amulet_nbt import from_snbt
from PyMCTranslate.py3.api import Block
from gen_fan_out_mapping import translations, iter_full_blocks, JAVA_DIR, TMP_DIR

TARGET_VERSION = (1, 21, 4)


def single_layer(block: Block) -> Block:
    return Block(
        block.namespace,
        block.base_name,
        {key: value for key, value in block.properties.items() if key != "waterlogged"},
    )


def state_text(block: Block) -> str:
    properties = ",".join(
        f'{json.dumps(key)}={value.to_snbt()}'
        for key, value in sorted(block.properties.items())
        if key != "waterlogged"
    )
    return f"{block.namespaced_name} [{properties}]"


def translate_state(*, source, target, block: Block, native: set[str]) -> tuple[str, bool]:
    universal, entity, needs_context = source.to_universal(block, force_blockstate=True)
    source_needs_context = needs_context
    result, _, needs_context = target.from_universal(
        universal, entity, force_blockstate=True,
    )
    if not isinstance(result, Block):
        raise ValueError("target translation produces an entity")
    if result.extra_blocks:
        raise ValueError("translation produces additional block layers")
    # Complete defaults in the target schema, then require a real target state.
    specification = target.get_specification(
        namespace=result.namespace,
        base_name=result.base_name,
        force_blockstate=True,
    )
    properties = {
        key: from_snbt(value)
        for key, value in specification.get("defaults", {}).items()
    }
    properties.update(result.properties)
    normalized = state_text(Block(result.namespace, result.base_name, properties))
    if normalized not in native:
        raise ValueError("translation is outside the target state universe")
    return normalized, source_needs_context or needs_context


def generate(*, output: Path, report_path: Path) -> dict:
    native = {
        state_text(block)
        for block in iter_full_blocks("java", TARGET_VERSION)
    }
    target = translations.get_version("java", TARGET_VERSION).block
    seen: set[str] = set()
    mappings: dict[str, str] = {}
    rejected: list[dict] = []
    context_dependent: list[dict] = []
    versions = sorted(translations.version_numbers("java"), reverse=True)
    for version in versions:
        source = translations.get_version("java", version).block
        added = 0
        for original in iter_full_blocks("java", version):
            block = single_layer(original)
            key = state_text(block)
            if key in seen:
                continue
            # Claim before translating: a failed newer interpretation must not
            # silently acquire a different meaning from an older version.
            seen.add(key)
            added += 1
            if version == TARGET_VERSION:
                continue
            try:
                value, needs_context = translate_state(source=source, target=target, block=block, native=native)
            except (KeyError, ValueError) as error:
                rejected.append({"state": key, "version": list(version), "reason": str(error)})
                continue
            if needs_context:
                context_dependent.append({"state": key, "version": list(version), "target": value})
            if key != value:
                mappings[key] = value
        print(f"java {version}: {added} new source states", flush=True)
    # Context-dependent candidates are accepted when they form a valid target
    # state, matching the best-effort preview policy; the report keeps them visible.
    # Target identities live in the existing native state list; store only additions.
    report = {
        "target": list(TARGET_VERSION),
        "sourceVersions": versions,
        "nativeStates": len(native),
        "uniqueSourceStates": len(seen),
        "compatibilityMappings": len(mappings),
        "rejected": rejected,
        "contextDependent": context_dependent,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text("".join(f"{key}#{value}\n" for key, value in sorted(mappings.items())), encoding="utf-8")
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(json.dumps(report, indent=2, ensure_ascii=False)+"\n", encoding="utf-8")
    print(f"native={len(native)}, compatible={len(mappings)}, context-dependent={len(context_dependent)}, rejected={len(rejected)}", flush=True)
    return report


if __name__ == "__main__":
    generate(
        output=JAVA_DIR / "java_compatibility.list",
        report_path=TMP_DIR / "java_compatibility_report.json",
    )
