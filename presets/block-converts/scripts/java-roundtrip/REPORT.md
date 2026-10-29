# Java → NEMC → Java 往返差异报告

本报告枚举 JavaFullStates()，依次调用 ConvertFromJava 和 ConvertToJava，将原状态和返回状态分别 Tidy 后比较。它反映当前转换链的信息损失或不一致，不能单独证明 Bedrock 本身无法表达，也不能保证所有差异都能由 fixer 恢复。

输入：presets/java/v1，Java 1.21.4；转换：presets/block-converts/v1。

在仓库根目录重新生成：

```sh
go run ./presets/block-converts/scripts/java-roundtrip > presets/block-converts/scripts/java-roundtrip/REPORT.md
```

## 总览

- 枚举：18254 个状态，1095 种方块。
- 往返不一致或失败：8908 个状态（48.80%），涉及 323 种方块；一致 9346 个状态。
- Java → NEMC 失败：0；NEMC → Java 失败：0。
- 两步成功但不一致：8908；其中方块名变化 587，属性变化 8325（两者可能重叠）。
- 输入包含 waterlogged 属性的状态：0。结论只覆盖输入实际枚举的属性，未枚举的属性不能据此判断是否保留。

## 重点类别核对

包含往返完全一致的类别；属性列是该类别所有差异属性的并集，具体计数与改名目标见逐方块表。

| 类别（名称匹配） | 受影响 / 总方块种数 | 不一致 / 总状态 | 差异属性 | 方块名变化状态数 |
| --- | ---: | ---: | --- | ---: |
| 栅栏（`*_fence`） | 13 / 13 | 195 / 208 | east; north; south; west | 0 |
| 墙（`*_wall`） | 0 / 26 | 0 / 4212 | — | 0 |
| 门（`*_door`） | 21 / 21 | 1134 / 1344 | facing; hinge; open; powered | 0 |
| 床（`*_bed`） | 15 / 16 | 240 / 256 | — | 240 |
| 楼梯（`*_stairs`） | 58 / 58 | 1856 / 2320 | shape | 0 |
| 铁轨（`*rail`） | 0 / 4 | 0 / 46 | — | 0 |
| 玻璃板（`*glass_pane`） | 17 / 17 | 255 / 272 | east; north; south; west | 0 |
| 铁栏杆（`*iron_bars`） | 1 / 1 | 15 / 16 | east; north; south; west | 0 |
| 栅栏门（`*_fence_gate`） | 12 / 12 | 192 / 384 | powered | 0 |
| 活板门（`*_trapdoor`） | 21 / 21 | 336 / 672 | powered | 0 |

## 属性影响范围

按受影响方块种数降序排列。改值、缺失、新增分别表示同名属性值不同、返回状态没有该属性、返回状态多出该属性。仅统计两步成功的状态；转换失败不推断具体丢失属性。各属性计数可能来自同一状态，不能相加当作状态总数。

| 属性 | 方块种数 | 改值状态数 | 缺失状态数 | 新增状态数 |
| --- | ---: | ---: | ---: | ---: |
| `powered` | 72 | 1943 | 0 | 0 |
| `shape` | 58 | 1856 | 0 | 0 |
| `facing` | 55 | 785 | 0 | 0 |
| `east` | 38 | 1536 | 0 | 0 |
| `north` | 38 | 1536 | 0 | 0 |
| `south` | 38 | 1536 | 0 | 0 |
| `west` | 38 | 1536 | 0 | 0 |
| `hinge` | 21 | 336 | 0 | 0 |
| `open` | 21 | 336 | 0 | 0 |
| `distance` | 11 | 132 | 0 | 0 |
| `rotation` | 7 | 210 | 0 | 0 |
| `up` | 6 | 397 | 0 | 0 |
| `age` | 5 | 2 | 0 | 4 |
| `down` | 4 | 91 | 0 | 0 |
| `power` | 3 | 240 | 0 | 0 |
| `type` | 3 | 22 | 0 | 0 |
| `snowy` | 3 | 3 | 0 | 0 |
| `signal_fire` | 2 | 16 | 0 | 0 |
| `extended` | 2 | 12 | 0 | 0 |
| `note` | 1 | 1104 | 0 | 0 |
| `instrument` | 1 | 1100 | 0 | 0 |
| `locked` | 1 | 32 | 0 | 0 |
| `short` | 1 | 12 | 0 | 0 |
| `stage` | 1 | 10 | 0 | 0 |
| `bottom` | 1 | 8 | 0 | 0 |
| `has_book` | 1 | 8 | 0 | 0 |
| `cracked` | 1 | 4 | 0 | 0 |
| `has_record` | 1 | 1 | 0 | 0 |

## 按方块列出全部差异

按不一致状态数降序排列。方块名省略 minecraft: 前缀；属性后的数字为受影响状态数，括号标明差异类型。失败列依次为 Java → NEMC / NEMC → Java。方块改名时，属性差异仍按原名归组，需结合返回方块名理解。

| 方块 | 不一致 / 总状态 | 变化或丢失的属性 | 返回新增属性 | 返回方块名（状态数） | 失败 |
| --- | ---: | --- | --- | --- | ---: |
| `redstone_wire` | 1280 / 1296 | `east`:864（改值）; `north`:864（改值）; `south`:864（改值）; `west`:864（改值） | — | — | 0 / 0 |
| `note_block` | 1149 / 1150 | `instrument`:1100（改值）; `note`:1104（改值）; `powered`:575（改值） | — | — | 0 / 0 |
| `fire` | 496 / 512 | `east`:256（改值）; `north`:256（改值）; `south`:256（改值）; `up`:256（改值）; `west`:256（改值） | — | — | 0 / 0 |
| `calibrated_sculk_sensor` | 180 / 192 | `power`:180（改值） | — | — | 0 / 0 |
| `tripwire` | 120 / 128 | `east`:64（改值）; `north`:64（改值）; `south`:64（改值）; `west`:64（改值） | — | — | 0 / 0 |
| `chorus_plant` | 63 / 64 | `down`:32（改值）; `east`:32（改值）; `north`:32（改值）; `south`:32（改值）; `up`:32（改值）; `west`:32（改值） | — | — | 0 / 0 |
| `mushroom_stem` | 62 / 64 | `down`:31（改值）; `east`:32（改值）; `north`:32（改值）; `south`:32（改值）; `up`:31（改值）; `west`:32（改值） | — | — | 0 / 0 |
| `acacia_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `bamboo_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `birch_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `cherry_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `crimson_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `dark_oak_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `exposed_copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `iron_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `jungle_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `mangrove_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `oak_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `oxidized_copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `pale_oak_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `spruce_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `warped_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `waxed_copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `waxed_exposed_copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `waxed_oxidized_copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `waxed_weathered_copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `weathered_copper_door` | 54 / 64 | `facing`:24（改值）; `hinge`:16（改值）; `open`:16（改值）; `powered`:32（改值） | — | — | 0 / 0 |
| `brown_mushroom_block` | 53 / 64 | `down`:14（改值）; `east`:20（改值）; `north`:20（改值）; `south`:20（改值）; `up`:31（改值）; `west`:20（改值） | — | — | 0 / 0 |
| `red_mushroom_block` | 53 / 64 | `down`:14（改值）; `east`:20（改值）; `north`:20（改值）; `south`:20（改值）; `up`:31（改值）; `west`:20（改值） | — | — | 0 / 0 |
| `sculk_sensor` | 45 / 48 | `power`:45（改值） | — | — | 0 / 0 |
| `acacia_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `andesite_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `bamboo_mosaic_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `bamboo_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `birch_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `blackstone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `cherry_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `cobbled_deepslate_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `cobblestone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `crimson_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `dark_oak_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `dark_prismarine_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `deepslate_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `deepslate_tile_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `diorite_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `end_stone_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `exposed_cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `granite_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `jungle_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `mangrove_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `mossy_cobblestone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `mossy_stone_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `mud_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `nether_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `oak_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `oxidized_cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `pale_oak_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `polished_andesite_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `polished_blackstone_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `polished_blackstone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `polished_deepslate_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `polished_diorite_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `polished_granite_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `polished_tuff_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `prismarine_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `prismarine_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `purpur_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `quartz_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `red_nether_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `red_sandstone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `repeater` | 32 / 64 | `locked`:32（改值） | — | — | 0 / 0 |
| `resin_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `sandstone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `smooth_quartz_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `smooth_red_sandstone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `smooth_sandstone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `spruce_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `stone_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `stone_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `tuff_brick_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `tuff_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `warped_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `waxed_cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `waxed_exposed_cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `waxed_oxidized_cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `waxed_weathered_cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `weathered_cut_copper_stairs` | 32 / 40 | `shape`:32（改值） | — | — | 0 / 0 |
| `creeper_head` | 31 / 32 | `powered`:16（改值）; `rotation`:30（改值） | — | — | 0 / 0 |
| `dragon_head` | 31 / 32 | `powered`:16（改值）; `rotation`:30（改值） | — | — | 0 / 0 |
| `piglin_head` | 31 / 32 | `powered`:16（改值）; `rotation`:30（改值） | — | — | 0 / 0 |
| `player_head` | 31 / 32 | `powered`:16（改值）; `rotation`:30（改值） | — | — | 0 / 0 |
| `skeleton_skull` | 31 / 32 | `powered`:16（改值）; `rotation`:30（改值） | — | — | 0 / 0 |
| `wither_skeleton_skull` | 31 / 32 | `powered`:16（改值）; `rotation`:30（改值） | — | — | 0 / 0 |
| `zombie_head` | 31 / 32 | `powered`:16（改值）; `rotation`:30（改值） | — | — | 0 / 0 |
| `acacia_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `acacia_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `bamboo_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `bamboo_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `bell` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `birch_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `birch_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `black_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `black_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `blue_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `blue_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `brown_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `brown_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `cherry_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `cherry_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `crimson_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `crimson_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `cyan_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `cyan_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `dark_oak_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `dark_oak_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `exposed_copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `gray_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `gray_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `green_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `green_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `iron_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `jungle_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `jungle_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `light_blue_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `light_blue_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `light_gray_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `light_gray_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `lime_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `lime_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `magenta_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `magenta_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `mangrove_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `mangrove_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `oak_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `oak_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `orange_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `orange_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `oxidized_copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `pale_oak_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `pale_oak_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `pink_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `pink_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `purple_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `purple_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `red_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `spruce_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `spruce_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `vine` | 16 / 32 | `up`:16（改值） | — | — | 0 / 0 |
| `warped_fence_gate` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `warped_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `waxed_copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `waxed_exposed_copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `waxed_oxidized_copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `waxed_weathered_copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `weathered_copper_trapdoor` | 16 / 32 | `powered`:16（改值） | — | — | 0 / 0 |
| `white_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `yellow_banner` | 16 / 16 | — | — | `white_banner`:16 | 0 / 0 |
| `yellow_bed` | 16 / 16 | — | — | `red_bed`:16 | 0 / 0 |
| `acacia_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `bamboo_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `birch_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `black_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `blue_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `brown_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `cherry_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `crimson_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `cyan_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `dark_oak_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `gray_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `green_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `iron_bars` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `jungle_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `light_blue_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `light_gray_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `lime_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `magenta_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `mangrove_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `nether_brick_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `oak_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `orange_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `pale_oak_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `pink_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `purple_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `red_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `spruce_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `target` | 15 / 16 | `power`:15（改值） | — | — | 0 / 0 |
| `warped_fence` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `white_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `yellow_stained_glass_pane` | 15 / 16 | `east`:8（改值）; `north`:8（改值）; `south`:8（改值）; `west`:8（改值） | — | — | 0 / 0 |
| `acacia_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `acacia_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `azalea_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `bamboo_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `birch_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `birch_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `cherry_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `cherry_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `crimson_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `dark_oak_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `dark_oak_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `flowering_azalea_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `jungle_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `jungle_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `mangrove_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `mangrove_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `oak_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `oak_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `pale_oak_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `pale_oak_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `piston_head` | 12 / 24 | `short`:12（改值） | — | — | 0 / 0 |
| `polished_blackstone_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `spruce_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `spruce_leaves` | 12 / 14 | `distance`:12（改值） | — | — | 0 / 0 |
| `stone_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `warped_button` | 12 / 24 | `facing`:12（改值） | — | — | 0 / 0 |
| `moving_piston` | 11 / 12 | `facing`:10（改值）; `type`:6（改值） | — | — | 0 / 0 |
| `observer` | 11 / 12 | `facing`:10（改值）; `powered`:6（改值） | — | — | 0 / 0 |
| `mangrove_propagule` | 10 / 20 | `stage`:10（改值） | — | — | 0 / 0 |
| `campfire` | 8 / 16 | `signal_fire`:8（改值） | — | — | 0 / 0 |
| `chest` | 8 / 12 | `type`:8（改值） | — | — | 0 / 0 |
| `lectern` | 8 / 16 | `has_book`:8（改值） | — | — | 0 / 0 |
| `lever` | 8 / 24 | `facing`:8（改值） | — | — | 0 / 0 |
| `scaffolding` | 8 / 16 | `bottom`:8（改值） | — | — | 0 / 0 |
| `soul_campfire` | 8 / 16 | `signal_fire`:8（改值） | — | — | 0 / 0 |
| `trapped_chest` | 8 / 12 | `type`:8（改值） | — | — | 0 / 0 |
| `lightning_rod` | 6 / 12 | `powered`:6（改值） | — | — | 0 / 0 |
| `piston` | 6 / 12 | `extended`:6（改值） | — | — | 0 / 0 |
| `sticky_piston` | 6 / 12 | `extended`:6（改值） | — | — | 0 / 0 |
| `black_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `blue_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `brown_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `cyan_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `gray_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `green_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `light_blue_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `light_gray_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `lime_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `magenta_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `orange_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `pink_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `purple_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `red_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `white_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `yellow_shulker_box` | 5 / 6 | `facing`:5（改值） | — | — | 0 / 0 |
| `black_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `blue_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `brown_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `creeper_wall_head` | 4 / 8 | `powered`:4（改值） | — | — | 0 / 0 |
| `cyan_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `decorated_pot` | 4 / 8 | `cracked`:4（改值） | — | — | 0 / 0 |
| `dragon_wall_head` | 4 / 8 | `powered`:4（改值） | — | — | 0 / 0 |
| `gray_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `green_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `light_blue_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `light_gray_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `lime_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `magenta_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `orange_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `piglin_wall_head` | 4 / 8 | `powered`:4（改值） | — | — | 0 / 0 |
| `pink_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `player_wall_head` | 4 / 8 | `powered`:4（改值） | — | — | 0 / 0 |
| `purple_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `red_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `skeleton_wall_skull` | 4 / 8 | `powered`:4（改值） | — | — | 0 / 0 |
| `wither_skeleton_wall_skull` | 4 / 8 | `powered`:4（改值） | — | — | 0 / 0 |
| `yellow_wall_banner` | 4 / 4 | — | — | `white_wall_banner`:4 | 0 / 0 |
| `zombie_wall_head` | 4 / 8 | `powered`:4（改值） | — | — | 0 / 0 |
| `resin_brick_slab` | 3 / 3 | — | — | `andesite_slab`:3 | 0 / 0 |
| `pitcher_crop` | 2 / 10 | `age`:2（改值） | — | — | 0 / 0 |
| `blue_orchid` | 1 / 1 | — | — | `short_grass`:1 | 0 / 0 |
| `cave_air` | 1 / 1 | — | — | `air`:1 | 0 / 0 |
| `cave_vines_plant` | 1 / 2 | — | `age`:1 | `cave_vines`:1 | 0 / 0 |
| `grass_block` | 1 / 2 | `snowy`:1（改值） | — | — | 0 / 0 |
| `jukebox` | 1 / 2 | `has_record`:1（改值） | — | — | 0 / 0 |
| `kelp_plant` | 1 / 1 | — | `age`:1 | `kelp`:1 | 0 / 0 |
| `mycelium` | 1 / 2 | `snowy`:1（改值） | — | — | 0 / 0 |
| `podzol` | 1 / 2 | `snowy`:1（改值） | — | — | 0 / 0 |
| `potted_acacia_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_allium` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_azalea_bush` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_azure_bluet` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_bamboo` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_birch_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_blue_orchid` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_brown_mushroom` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_cactus` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_cherry_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_closed_eyeblossom` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_cornflower` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_crimson_fungus` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_crimson_roots` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_dandelion` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_dark_oak_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_dead_bush` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_fern` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_flowering_azalea_bush` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_jungle_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_lily_of_the_valley` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_mangrove_propagule` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_oak_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_open_eyeblossom` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_orange_tulip` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_oxeye_daisy` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_pale_oak_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_pink_tulip` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_poppy` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_red_mushroom` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_red_tulip` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_spruce_sapling` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_torchflower` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_warped_fungus` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_warped_roots` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_white_tulip` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `potted_wither_rose` | 1 / 1 | — | — | `flower_pot`:1 | 0 / 0 |
| `twisting_vines_plant` | 1 / 1 | — | `age`:1 | `twisting_vines`:1 | 0 / 0 |
| `void_air` | 1 / 1 | — | — | `air`:1 | 0 / 0 |
| `weeping_vines_plant` | 1 / 1 | — | `age`:1 | `weeping_vines`:1 | 0 / 0 |
