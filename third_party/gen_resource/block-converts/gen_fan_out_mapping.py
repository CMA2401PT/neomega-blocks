from pathlib import Path
import itertools
from typing import Dict, Iterator, List, Set, Tuple, cast
import PyMCTranslate
from amulet_nbt import ByteTag,IntTag,StringTag
from PyMCTranslate.py3.api import Block
from target_versions import JAVA_VERSION, BEDROCK_VERSION


SCRIPT_DIR = Path(__file__).resolve().parent
TMP_DIR = SCRIPT_DIR / "tmp"

JAVA_DIR = SCRIPT_DIR.parents[2] / "presets" / "java" / "v1"

translations = PyMCTranslate.new_translation_manager()


def prop_value_to_nbt(prop_value:str):
    if prop_value.startswith('"'):
        return StringTag(prop_value[1:-1])
    elif prop_value=="0b":
        return ByteTag(0)
    elif prop_value=="1b":
        return ByteTag(1)
    else:
        return IntTag(int(prop_value))

def iter_props(props:List[Tuple[str,List[str]]],defaults:Dict[str,str]):
    props_nbt: List[Tuple[str, List[StringTag | ByteTag | IntTag]]]=[(k,[
        prop_value_to_nbt(v) for v in vs]) 
        for k,vs in props
    ]
    defaults_nbt={k:prop_value_to_nbt(v) for k,v in defaults.items()}
    # 第一个返回有缺失的参数，第二个返回缺失值被默认值替代的参数
    def _iter(props_nbt:List[Tuple[str, List[StringTag | ByteTag | IntTag]]])\
        ->Iterator[Tuple[Dict[str,StringTag | ByteTag | IntTag],Dict[str,StringTag | ByteTag | IntTag]]]:
        if len(props_nbt)==0:
            yield {},{}
        else:
            this_prop_name,this_prop_values=props_nbt[0]
            other_props=props_nbt[1:]
            for ps,ps_with_default in _iter(other_props):
                ps_with_default[this_prop_name]=defaults_nbt[this_prop_name]
                yield ps,ps_with_default
            for prop_value in this_prop_values[::-1]:
                for ps,ps_with_default in _iter(other_props):
                    ps[this_prop_name]=prop_value
                    ps_with_default[this_prop_name]=prop_value
                    yield ps,ps_with_default
    return _iter(props_nbt)

def block_to_blockstr(block:Block)->str:
    name=block.namespaced_name
    props=block.properties
    def cast_prop(prop:ByteTag|IntTag|StringTag):
        if isinstance(prop,ByteTag):
            if prop.py_int==0:
                return "false"
            elif prop.py_int==1:
                return "true"
            else:
                raise ValueError(prop)
        elif isinstance(prop,IntTag):
            return prop.to_snbt()
        else:
            assert isinstance(prop,StringTag)
            return prop.to_snbt()
    prop_strings=[f'"{k}"={cast_prop(v)}' for k,v in sorted(props.items())]
    return f"{name} [{','.join(prop_strings)}]"

def all_blocks(platform_name:str,version:Tuple[int,...],allow_missing:bool=True):
    # allow_missing 保留但不参与逻辑；转换必须使用补齐默认属性的对象。
    exist:Set[str]=set()
    version_data=translations.get_version(platform_name,version)
    block_translator=version_data.block
    def iter_blocks()->Iterator[Tuple[Block,Block]]:
        block_names=block_translator.base_names("minecraft",force_blockstate=True)
        for block_name in block_names:
            block_specification = block_translator.get_specification(
                namespace="minecraft",
                base_name=block_name,
                force_blockstate=True,
            )
            properties: Dict[str, List[str]] = cast(Dict[str,List[str]],block_specification.get("properties", {}))
            defaults: Dict[str, str]=cast(Dict[str,str],block_specification.get("defaults",{}))
            for props,props_with_default in iter_props(list(properties.items()),defaults=defaults):
                block:Block=cast(Block,Block("minecraft",block_name,props))
                block_with_default:Block=cast(Block,Block("minecraft",block_name,props_with_default))
                block_str=block_to_blockstr(block)
                # block_str:str=cast(str,block.snbt_blockstate)
                if block_str in exist:
                    continue
                exist.add(block_str)
                yield block,block_with_default
    return block_translator,iter_blocks()


def iter_full_blocks(platform_name:str,version:Tuple[int,...])->Iterator[Block]:
    # 每个字段都取遍自己的取值域，做完整笛卡尔积；没有缺字段的残缺态。
    # 中心状态必须包含 Java 独有属性，才能生成兼容表并分析转换中的信息损失。
    version_data=translations.get_version(platform_name,version)
    block_translator=version_data.block
    block_names=block_translator.base_names("minecraft",force_blockstate=True)
    for block_name in block_names:
        block_specification=block_translator.get_specification(
            namespace="minecraft",
            base_name=block_name,
            force_blockstate=True,
        )
        properties:Dict[str,List[str]]=cast(Dict[str,List[str]],block_specification.get("properties",{}))
        keys=list(properties.keys())
        if len(keys)==0:
            yield cast(Block,Block("minecraft",block_name,{}))
            continue
        value_lists=[[prop_value_to_nbt(v) for v in properties[k]] for k in keys]
        for combo in itertools.product(*value_lists):
            props={k:v for k,v in zip(keys,combo)}
            yield cast(Block,Block("minecraft",block_name,props))


def iter_block_schema(
    platform_name: str,
    version: tuple[int, ...],
) -> Iterator[tuple[str, dict[str, list[str]], dict[str, str]]]:
    # 每个方块一条：属性名、取值域、默认值。fixer 据此知道某方块能改哪些
    # 属性、默认是什么，不必从完整组合反推（反推不出默认值）。
    version_data=translations.get_version(platform_name,version)
    block_translator=version_data.block
    for block_name in block_translator.base_names("minecraft",force_blockstate=True):
        spec=block_translator.get_specification(
            namespace="minecraft",
            base_name=block_name,
            force_blockstate=True,
        )
        properties:Dict[str,List[str]]=cast(Dict[str,List[str]],spec.get("properties",{}))
        defaults:Dict[str,str]=cast(Dict[str,str],spec.get("defaults",{}))
        yield block_name,properties,defaults


def schema_line(block_name:str,properties:Dict[str,List[str]],defaults:Dict[str,str])->str:
    # 格式：minecraft:name<TAB>key=default|v1,v2,v3<TAB>key2=...
    # 默认值写在等号后、取值域前，二者用竖线分隔；无属性方块只有方块名。
    fields: list[str]=[]
    for key,values in sorted(properties.items()):
        if key == "waterlogged":
            continue
        default=defaults[key]
        fields.append(f"{key}={default}|{','.join(values)}")
    return "\t".join([f"minecraft:{block_name}",*fields])


def main(*, tmp_dir: Path = TMP_DIR, java_dir: Path = JAVA_DIR) -> None:
    tmp_dir.mkdir(parents=True, exist_ok=True)
    java_dir.mkdir(parents=True, exist_ok=True)
    with (java_dir / "java_block_schema.list").open("w") as f:
        for name, properties, defaults in iter_block_schema("java", JAVA_VERSION):
            f.write(f"{schema_line(name, properties, defaults)}\n")
    # 单层全集将水交给独立层表达；完整含水状态另供往返分析。
    dry_states: set[str] = {
        block_to_blockstr(Block(
            block.namespace,
            block.base_name,
            {key: value for key, value in block.properties.items() if key != "waterlogged"},
        ))
        for block in iter_full_blocks("java", JAVA_VERSION)
    }
    (java_dir / "java_states.list").write_text("".join(f"{state}\n" for state in sorted(dry_states)))
    # Bedrock 的原始 NBT 类型随 specification 保留，Go 用它计算网络哈希。
    with (tmp_dir / "bedrock_states.list").open("w") as f:
        for block in iter_full_blocks("bedrock", BEDROCK_VERSION):
            f.write(f"{block_to_blockstr(block)}\n")

    with (tmp_dir / "hub_java_states.list").open("w") as f:
        for java_block in iter_full_blocks("java",JAVA_VERSION):
            f.write(f"{block_to_blockstr(java_block)}\n")

    output_translator,output_iter=all_blocks("java",JAVA_VERSION)
    with (tmp_dir / "fan_out_java_allowed.list").open("w") as f:
        for java_block,_ in output_iter:
            f.write(f"{block_to_blockstr(java_block)}\n")


    input_translator,input_iter=all_blocks("bedrock",BEDROCK_VERSION)
    with (tmp_dir / "fan_out_java.mapping").open("w") as f:
        for bedrock_block,block_with_default in input_iter:
            input_str:str=block_to_blockstr(bedrock_block)
            uni_block,_,ok=input_translator.to_universal(
                block_with_default,
                force_blockstate=True,
            )
            out_block,_,ok=output_translator.from_universal(
                uni_block,
                force_blockstate=True,
            )
            output_str=block_to_blockstr(out_block)
            f.write(f"{input_str}#{output_str}\n")


if __name__ == "__main__":
    main()
