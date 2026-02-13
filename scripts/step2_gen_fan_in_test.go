package scripts

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"neomega-blocks/ir"
	presets "neomega-blocks/presets/v1"
	"neomega-blocks/specific_converts"
	"neomega-blocks/storage/num_lookup"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type MappingPair struct {
	In, Out ir.IRBlock
}

func LoadMappingPairs(file string) []MappingPair {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(bytes.NewReader(data))
	pairs := []MappingPair{}
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineStr := strings.TrimSpace(string(line))
		idx := strings.Index(lineStr, "#")
		if idx == -1 {
			panic(-1)
		}
		pairs = append(pairs, MappingPair{
			In:  ir.IRBlock(lineStr[:idx]).Tidy(),
			Out: ir.IRBlock(lineStr[idx+1:]).Tidy(),
		})
	}
	fmt.Println(len(pairs))
	return pairs
}

func outputFix(in ir.IRBlock) ir.IRBlock {
	var err error
	var outBlock ir.IRBlock
	switch in.ShortName() {
	case "stone_slab", "stone_slab2", "stone_slab3", "stone_slab4", "wooden_slab":
		in = ir.IRBlock(fmt.Sprintf("%v %v", in.NameWithNameSpace(), in.States().Update(func(s string, a any) (string, any) {
			if s == "stone_slab_type" {
				s = "material"
			}
			if s == "material" && a == "stone" {
				return "material", "stone_brick"
			}
			return s, a
		})))
		material, ok := in.States().Get("material")
		if ok {
			outBlock, err = ir.NewIRBlockFromGo(
				fmt.Sprintf("minecraft:%v_slab", material),
				map[string]any{"minecraft:vertical_half": in.States().GetDefault("type", "bottom")},
			)
		}
	case "carpet":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_carpet", in.States().GetDefault("color", "")),
			nil,
		)
	case "double_wooden_slab", "double_stone_slab", "double_stone_slab2", "double_stone_slab3", "double_stone_slab4":
		in = ir.IRBlock(fmt.Sprintf("%v %v", in.NameWithNameSpace(), in.States().Update(func(s string, a any) (string, any) {
			if s == "material" {
				switch a {
				case "stone":
					return "material", "stone_brick"
				case "smooth_stone_block":
					return "material", "smooth_stone"
				default:
					return s, a
				}

			}
			return s, a
		})))
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_double_slab", in.States().GetDefault("material", "")),
			nil,
		)

	case "stained_glass":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_stained_glass", in.States().GetDefault("color", "")),
			nil,
		)
	case "stained_glass_pane":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_stained_glass_pane", in.States().GetDefault("color", "")),
			nil,
		)
	case "stained_hardened_clay":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_terracotta", in.States().GetDefault("color", "")),
			nil,
		)
	case "leaves", "leaves2":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_leaves", in.States().GetDefault("material", "")),
			nil,
		)
	case "planks":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_planks", in.States().GetDefault("material", "")),
			nil,
		)
	case "log", "log2":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_log", in.States().GetDefault("material", "")),
			nil,
		)
	case "fence":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_fence", in.States().GetDefault("material", "oak")),
			nil,
		)
	case "stonebrick":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_stone_bricks", in.States().GetDefault("variant", "")),
			nil,
		)
		if outBlock == "minecraft:normal_stone_bricks []" {
			outBlock = ir.IRBlock("minecraft:stone_bricks []")
		}
	case "yellow_flower":
		outBlock = ir.IRBlock("minecraft:dandelion []")
	case "shulker_box":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_shulker_box", in.States().GetDefault("color", "")),
			nil,
		)
	case "silver_shulker_box":
		outBlock = ir.IRBlock("minecraft:light_gray_shulker_box []")
	case "coral_fan_hang", "coral_fan_hang2", "coral_fan_hang3":
		if in.States().GetDefault("dead", "false") == "false" {
			outBlock, err = ir.NewIRBlockFromGo(
				fmt.Sprintf("minecraft:%v_coral_fan", in.States().GetDefault("coral_type", "")),
				nil,
			)
		} else {
			outBlock, err = ir.NewIRBlockFromGo(
				fmt.Sprintf("minecraft:dead_%v_coral_fan", in.States().GetDefault("coral_type", "")),
				nil,
			)
		}
	case "coral_fan":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_coral_fan", in.States().GetDefault("coral_type", "")),
			nil,
		)
	case "coral_fan_dead":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:dead_%v_coral_fan", in.States().GetDefault("coral_type", "")),
			nil,
		)
	case "coral":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_coral", in.States().GetDefault("coral_type", "")),
			nil,
		)
	case "coral_block":
		if in.States().GetDefault("dead", "false") == "false" {
			outBlock, err = ir.NewIRBlockFromGo(
				fmt.Sprintf("minecraft:%v_coral_block", in.States().GetDefault("coral_type", "")),
				nil,
			)
		} else {
			outBlock, err = ir.NewIRBlockFromGo(
				fmt.Sprintf("minecraft:dead_%v_coral_block", in.States().GetDefault("coral_type", "")),
				nil,
			)
		}
	case "wool":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_wool", in.States().GetDefault("color", "")),
			nil,
		)

	case "concretePowder":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_concrete_powder", in.States().GetDefault("color", "")),
			nil,
		)
	case "concrete_powder":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_concrete_powder", in.States().GetDefault("color", "")),
			nil,
		)
	case "concrete":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v_concrete", in.States().GetDefault("color", "")),
			nil,
		)
	case "tallgrass":
		if in.States().GetDefault("plant_type", "") == "grass" {
			outBlock, err = ir.NewIRBlockFromGo(
				"minecraft:tall_grass",
				map[string]any{},
			)
		} else {
			outBlock, err = ir.NewIRBlockFromGo(
				"minecraft:large_fern",
				map[string]any{},
			)
		}
	case "skull":
		outBlock, err = ir.NewIRBlockFromGo(
			"minecraft:skeleton_skull",
			nil,
		)
	case "double_plant", "red_flower":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:%v", in.States().GetDefault("plant_type", "")),
			nil,
		)
	case "monster_egg":
		outBlock, err = ir.NewIRBlockFromGo(
			fmt.Sprintf("minecraft:infested_%v", in.States().GetDefault("material", "")),
			nil,
		)
	case "bell":
		in = ir.IRBlock(fmt.Sprintf("%v %v", in.NameWithNameSpace(), in.States().Update(func(s string, a any) (string, any) {
			if s == "attachment" {
				switch a {
				case "double_wall":
					return s, "multiple"
				case "single_wall":
					return s, "side"
				case "ceiling":
					return s, "hanging"
				case "floor":
					return s, "standing"
				}

			}
			return s, a
		})))
	case "tripWire":
		outBlock = ir.IRBlock(fmt.Sprintf("minecraft:trip_wire %v", in.States()))
	case "tripwire":
		outBlock = ir.IRBlock(fmt.Sprintf("minecraft:trip_wire %v", in.States()))
	case "piston_extension":
		outBlock = ir.IRBlock(fmt.Sprintf("minecraft:air %v", in.States()))
	case "pistonArmCollision":
		outBlock = ir.IRBlock(fmt.Sprintf("minecraft:piston_arm_collision %v", in.States().Update(func(s string, a any) (string, any) {
			if s == "facing" {
				switch a {
				case "east":
					return "facing_direction", int32(5)
				case "west":
					return "facing_direction", int32(4)
				case "south":
					return "facing_direction", int32(3)
				case "north":
					return "facing_direction", int32(2)
				case "up":
					return "facing_direction", int32(1)
				case "down":
					return "facing_direction", int32(0)
				}
			}
			return s, a
		})))
	}
	if err != nil {
		panic(err)
	}
	if outBlock != "" {
		return outBlock
	}
	return in
}

// 从 fan_in_bedrock.mapping 中读取所有基岩版方块映射的状态，并保存到 2_bedrock_fan_in.json
func TestStep2_1_GenBedrockFanin(t *testing.T) {
	pairs := LoadMappingPairs("fan_in_bedrock.mapping")
	tidyPairs := []MappingPair{}
	ignorePairs := 0
	nonsensePairs := 0
	for i, p := range pairs {
		cannonOut, ok := presets.NEMCBedrockCanon(outputFix(p.Out))
		if !ok {
			t.Errorf("output %v not in NEMC block space", p.Out.String())
		}
		_, ok = presets.NEMCBedrockHashToBlockStrict(cannonOut.Hash())
		if !ok {
			t.Errorf("output %v not in NEMC block space: hash=%v", p.Out.String(), p.Out.Hash())
		}
		pairs[i].Out = cannonOut
	}
	for _, p := range pairs {
		_, ok := presets.NEMCBedrockHashToBlockStrict(p.Out.Hash())
		if !ok {
			t.Fatalf("output %v not in NEMC block space: hash=%v", p.Out.String(), p.Out.Hash())
		}
		cannonIn, _ := presets.NEMCBedrockCanon(p.In)
		_, ok = presets.NEMCBedrockHashToBlockStrict(cannonIn.Hash())
		if !ok {
			if p.In.NameWithNameSpace() == p.Out.NameWithNameSpace() {
				ret := p.In.States().FuzzyCompare(p.Out.States())
				if ret.Extra == 0 && ret.Missing == 0 {
					nonsensePairs += 1
					continue
				}
			}
			tidyPairs = append(tidyPairs, p)
		} else {
			ignorePairs += 1
		}

	}
	fmt.Printf("skip %v pairs because cannon, skip %v because it is nonsense, %v remaining\n", ignorePairs, nonsensePairs, len(tidyPairs))
	outStr := ""
	for _, p := range tidyPairs {
		outStr += fmt.Sprintf("%v#%v\n", p.In.String(), p.Out.String())
	}
	os.WriteFile("fan_in_bedrock_tidy.mapping", []byte(outStr), 0644)
	lookup := num_lookup.NewTwoLayerNumMap[uint32, uint32]()
	for _, p := range tidyPairs {
		lookup.Set(uint32(p.In.Hash()), uint32(p.Out.Hash()))
	}
	if err := os.WriteFile("../presets/v1/2_bedrock_fan_in.bin", lookup.Dump(), 0o644); err != nil {
		t.Fatalf("write defaults bin: %v", err)
	}
}

// 从 PyMCTranslate 的数据库中读取 java 方块的默认状态
// 这可以帮助减少 fanin java 方块时需要保存的数据
func TestStep2_2_GenJavaDefaultPropFanin(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	baseDir := filepath.Dir(cur)
	versionsPath := filepath.Join(
		baseDir,
		"PyMCTranslate",
		"json",
		"versions",
	)
	entries, err := os.ReadDir(versionsPath)
	if err != nil {
		t.Fatalf("read versions dir %q: %v", versionsPath, err)
	}

	type javaVersionDir struct {
		name string
		ver  [3]int
	}
	javaDirs := make([]javaVersionDir, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "java_") {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(e.Name(), "java_"), "_")
		if len(parts) != 3 {
			continue
		}
		maj, err1 := strconv.Atoi(parts[0])
		min, err2 := strconv.Atoi(parts[1])
		patch, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		javaDirs = append(javaDirs, javaVersionDir{
			name: e.Name(),
			ver:  [3]int{maj, min, patch},
		})
	}
	if len(javaDirs) == 0 {
		t.Fatal("no java_* version directories found")
	}

	// 从高版本到低版本排序，如：java_1_21_4 -> java_1_21_2 -> ...
	sort.Slice(javaDirs, func(i, j int) bool {
		a := javaDirs[i].ver
		b := javaDirs[j].ver
		if a[0] != b[0] {
			return a[0] > b[0]
		}
		if a[1] != b[1] {
			return a[1] > b[1]
		}
		return a[2] > b[2]
	})

	defaults := map[string]ir.IRStates{}
	for _, d := range javaDirs {
		defaultsPath := filepath.Join(
			versionsPath,
			d.name,
			"block",
			"blockstate",
			"specification",
			"minecraft",
			"vanilla",
		)
		if _, err := os.Stat(defaultsPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("stat defaults path %q: %v", defaultsPath, err)
		}
		curDefaults := ReadDefaults(defaultsPath)
		for name, states := range curDefaults {
			// 高版本优先：若已存在则忽略低版本同名项。
			if _, exists := defaults[name]; exists {
				continue
			}
			defaults[name] = states
		}
	}
	t.Logf("merged java defaults from %d versions, total blocks=%d", len(javaDirs), len(defaults))

	// 保存 defaults map[string]ir.IRStates，和保存 map[string]string 一样直接 JSON 序列化。
	data, err := json.Marshal(defaults)
	if err != nil {
		t.Fatalf("marshal defaults: %v", err)
	}
	if err := os.WriteFile("../presets/v1/2_2_java_default_props.bin", data, 0o644); err != nil {
		t.Fatalf("write defaults bin: %v", err)
	}
}

func javaWallDirectConvert(inJava ir.IRBlock) (outBedrock ir.IRBlock, ok bool) {
	if !strings.HasSuffix(inJava.NameWithNameSpace(), "_wall") {
		return "", false
	}
	inJava, ok = presets.JavaCanon(inJava)
	if !ok {
		return "", false
	}
	outStates := inJava.States().Update(specific_converts.JavaTobedrockWallRelplacer)
	bedrockBlock := ir.IRBlock(inJava.NameWithNameSpace() + " " + string(outStates))

	if presets.HasNEMCBlockHash(bedrockBlock.Hash()) {
		return bedrockBlock, ok
	}
	fmt.Println(inJava)
	// fmt.Println(expect)
	return "", false
}

func javaHangingSignDirectConvert(inJava ir.IRBlock) (outBedrock ir.IRBlock, ok bool) {
	out, ok := specific_converts.JavaToBedrockHangingSign(inJava)
	if !ok {
		return "", false
	}
	// bedrockBlock, ok := presets.NEMCBedrockCanon(out)
	// if !ok {
	// 	return "", false
	// }
	if presets.HasNEMCBlockHash(out.Hash()) {
		return out, ok
	}
	fmt.Println(inJava)
	return "", false
}

// 从 fan_in_java.mapping 中读取所有基岩版方块映射的状态，并保存到 2_3_java_fan_in.json
func TestStep2_3_GenJavaFanin(t *testing.T) {
	pairs := LoadMappingPairs("fan_in_java.mapping")
	fullPairs := map[ir.IRBlock]ir.IRBlock{}
	tidyPairs := []MappingPair{}

	for i, p := range pairs {
		cannonOut, ok := presets.NEMCBedrockCanon(outputFix(p.Out))
		if !ok {
			t.Errorf("output %v not in NEMC block space", p.Out.String())
		}
		_, ok = presets.NEMCBedrockHashToBlockStrict(cannonOut.Hash())
		if !ok {
			t.Errorf("output %v not in NEMC block space: hash=%v", p.Out.String(), p.Out.Hash())
		}
		pairs[i].Out = cannonOut
	}
	for _, p := range pairs {
		_, ok := presets.NEMCBedrockHashToBlockStrict(p.Out.Hash())
		if !ok {
			t.Fatalf("output %v not in NEMC block space", p.Out.String())
		}
		if _, ok := javaWallDirectConvert(p.In); ok {
			continue
		}
		if _, ok := javaHangingSignDirectConvert(p.In); ok {
			continue
		}
		cannonIn, _ := presets.JavaCanon(p.In)

		if p.In == cannonIn {
			fullPairs[p.In] = p.Out
			tidyPairs = append(tidyPairs, p)
		}
	}
	directConvert := 0
	ignorePairs := 0
	nonCannonPairs := 0
	for _, p := range pairs {
		if _, ok := javaWallDirectConvert(p.In); ok {
			directConvert += 1
			continue
		}
		if _, ok := javaHangingSignDirectConvert(p.In); ok {
			continue
		}
		cannonIn, ok := presets.JavaCanon(p.In)
		if !ok {
			t.FailNow()
		}
		_, found := fullPairs[cannonIn]
		if found {
			ignorePairs += 1
		} else {
			nonCannonPairs += 1
			tidyPairs = append(tidyPairs, p)
		}
	}

	fmt.Println(len(fullPairs))
	fmt.Printf("skip %v pairs because cannon, skip %v pairs because direct, %v (%v non cannon) remaining\n", ignorePairs, directConvert, len(tidyPairs), nonCannonPairs)
	outStr := ""
	for _, p := range tidyPairs {
		outStr += fmt.Sprintf("%v#%v\n", p.In.String(), p.Out.String())
	}
	os.WriteFile("fan_in_java_tidy.mapping", []byte(outStr), 0644)
	lookup := num_lookup.NewTwoLayerNumMap[uint32, uint32]()
	for _, p := range tidyPairs {
		lookup.Set(uint32(p.In.Hash()), uint32(p.Out.Hash()))
	}
	if err := os.WriteFile("../presets/v1/2_3_java_fan_in.bin", lookup.Dump(), 0o644); err != nil {
		t.Fatalf("write defaults bin: %v", err)
	}
}
