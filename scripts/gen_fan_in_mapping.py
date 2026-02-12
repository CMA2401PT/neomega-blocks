import os
from typing import Dict, Generator, List, Set, Tuple, cast
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

def iter_props(props:List[Tuple[str,List[str]]]):
    props_nbt: List[Tuple[str, List[StringTag | ByteTag | IntTag]]]=[(k,[
        prop_value_to_nbt(v) for v in vs]) 
        for k,vs in props
    ]
    def _iter(props_nbt:List[Tuple[str, List[StringTag | ByteTag | IntTag]]])\
        ->Generator[Dict[str,StringTag | ByteTag | IntTag]]:
        if len(props_nbt)==0:
            yield {}
        else:
            this_prop_name,this_prop_values=props_nbt[0]
            other_props=props_nbt[1:]
            for ps in _iter(other_props):
                yield ps
            for prop_value in this_prop_values[::-1]:
                for ps in _iter(other_props):
                    ps[this_prop_name]=prop_value
                    yield ps
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
def all_blocks(platform_name:str,version:Tuple[int,...]):
    exist:Set[str]=set()
    version_data=translations.get_version(platform_name,version)
    block_translator=version_data.block
    def iter_blocks()->Generator[Block]:
        block_names=block_translator.base_names("minecraft",force_blockstate=True)
        for block_name in block_names:
            block_specification = block_translator.get_specification(
                "minecraft", block_name, force_blockstate=True
            )
            properties: Dict[str, List[str]] = cast(Dict[str,List[str]],block_specification.get("properties", {}))
            # defaults: Dict[str, str]=cast(Dict[str,str],block_specification.get("defaults",{}))
            for props in iter_props(list(properties.items())):
                block:Block=cast(Block,Block("minecraft",block_name,props))
                block_str=block_to_blockstr(block)
                # block_str:str=cast(str,block.snbt_blockstate)
                if block_str in exist:
                    continue
                exist.add(block_str)
                yield block
    return block_translator,iter_blocks()

exist:Set[str]=set()
def gen_fan_in(output_file:str,platform:str,output_translator:PyMCTranslate.BlockTranslator):
    with open(output_file,"w") as f:
        for version_number in translations.version_numbers(platform)[::-1]:
            # for blocks like stone, hanging sign, newer version could have more correct mapping than older version
            print(version_number)
            translater,block_iter=all_blocks(platform,version_number)
            for block in block_iter:
                blk_str:str=block_to_blockstr(block)
                if blk_str in exist:
                    continue
                exist.add(blk_str)
                uni_block=translater.to_universal(block)[0]
                out_block=output_translator.from_universal(uni_block)[0]
                out_str=block_to_blockstr(out_block)
                if out_str==blk_str:
                    continue
                f.write(f"{blk_str}#{out_str}\n")
platform="bedrock"
output_file=f"fan_in_{platform}.mapping"
output_translator: PyMCTranslate.BlockTranslator=translations.get_version("bedrock",(1,21,50)).block

gen_fan_in(output_file,platform,output_translator)

platform="java"
output_file=f"fan_in_{platform}.mapping"
output_translator: PyMCTranslate.BlockTranslator=translations.get_version("bedrock",(1,21,50)).block

gen_fan_in(output_file,platform,output_translator)