import os
import itertools
from typing import Dict, Iterator, List, Set, Tuple, cast
import PyMCTranslate
from amulet_nbt import ByteTag,IntTag,StringTag
from PyMCTranslate.py3.api import Block


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
    prop_strings=[f'"{k}"={cast_prop(v)}' for k,v in props.items()]
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
                "minecraft", block_name, force_blockstate=True
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
    # viewer 据此建 javaHash→状态 表，才能渲染北连栅栏这类 Bedrock 表达不了的状态。
    version_data=translations.get_version(platform_name,version)
    block_translator=version_data.block
    block_names=block_translator.base_names("minecraft",force_blockstate=True)
    for block_name in block_names:
        block_specification=block_translator.get_specification(
            "minecraft",block_name,force_blockstate=True
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


def iter_block_schema(platform_name:str,version:Tuple[int,...]):
    # 每个方块一条：属性名、取值域、默认值。fixer 据此知道某方块能改哪些
    # 属性、默认是什么，不必从完整组合反推（反推不出默认值）。
    version_data=translations.get_version(platform_name,version)
    block_translator=version_data.block
    for block_name in block_translator.base_names("minecraft",force_blockstate=True):
        spec=block_translator.get_specification("minecraft",block_name,force_blockstate=True)
        properties:Dict[str,List[str]]=cast(Dict[str,List[str]],spec.get("properties",{}))
        defaults:Dict[str,str]=cast(Dict[str,str],spec.get("defaults",{}))
        yield block_name,properties,defaults


def schema_line(block_name:str,properties:Dict[str,List[str]],defaults:Dict[str,str])->str:
    # 格式：minecraft:name<TAB>key=default|v1,v2,v3<TAB>key2=...
    # 默认值写在等号后、取值域前，二者用竖线分隔；无属性方块只有方块名。
    fields=[]
    for key,values in properties.items():
        default=defaults.get(key,values[0] if values else "")
        fields.append(f"{key}={default}|{','.join(values)}")
    return "\t".join([f"minecraft:{block_name}",*fields])


def main() -> None:
    # 方块知识库：名称 → 属性域 + 默认值。presets/java/v1 据此支持 fixer 的查询。
    with open(f"java_block_schema.list","w") as f:
        for block_name,properties,defaults in iter_block_schema("java",(1,21,4)):
            f.write(f"{schema_line(block_name,properties,defaults)}\n")

    # Java 侧完整状态全集：每行一个字段齐全的状态，供 viewer 直读建表。
    with open(f"fan_out_java_full.list","w") as f:
        for java_block in iter_full_blocks("java",(1,21,4)):
            f.write(f"{block_to_blockstr(java_block)}\n")

    output_translator,output_iter=all_blocks("java",(1,21,4))
    with open(f"fan_out_java_allowed.list","w") as f:
        for java_block,_ in output_iter:
            f.write(f"{block_to_blockstr(java_block)}\n")


    input_translator,input_iter=all_blocks("bedrock",(1,21,50))
    with open(f"fan_out_java.mapping","w") as f:
        for bedrock_block,block_with_default in input_iter:
            input_str:str=block_to_blockstr(bedrock_block)
            uni_block,_,ok=input_translator.to_universal(block_with_default)
            out_block,_,ok=output_translator.from_universal(uni_block)
            output_str=block_to_blockstr(out_block)
            f.write(f"{input_str}#{output_str}\n")


if __name__ == "__main__":
    main()
