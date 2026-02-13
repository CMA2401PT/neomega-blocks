package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"neomega-blocks/ir"
)

func ReadDefaults(path string) map[string]ir.IRStates {
	// 遍历 path (默认是 jsonData) 下的所有 json 文件，从每个json文件中抽取出默认的 key/value 表示
	// 其中， value 是 snbt 形式的，且只会为 byte,int,string 格式
	// 使用抽取出的默认值构造 ir.IRStates
	// 需要注意，有的文件没有任何属性，这代表这个方块没有属性/默认属性，此时返回 []

	entries, err := os.ReadDir(path)
	if err != nil {
		panic(fmt.Errorf("read defaults dir %q: %w", path, err))
	}

	out := make(map[string]ir.IRStates, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		payload, err := os.ReadFile(filePath)
		if err != nil {
			panic(fmt.Errorf("read defaults file %q: %w", filePath, err))
		}

		var doc struct {
			Defaults map[string]string `json:"defaults"`
		}
		if err := json.Unmarshal(payload, &doc); err != nil {
			panic(fmt.Errorf("decode defaults file %q: %w", filePath, err))
		}

		states := ir.IRStates("[]")
		for k, v := range doc.Defaults {
			parsed, err := parseSNBTDefaultValue(v)
			if err != nil {
				panic(fmt.Errorf("parse default value in %q key %q: %w", filePath, k, err))
			}
			states = states.Set(k, parsed)
		}

		blockName := "minecraft:" + strings.TrimSuffix(entry.Name(), ".json")
		out[blockName] = states
	}
	return out
}

func parseSNBTDefaultValue(v string) (any, error) {
	if v == "" {
		return nil, fmt.Errorf("empty default value")
	}
	if strings.HasSuffix(v, "b") {
		i, err := strconv.ParseInt(strings.TrimSuffix(v, "b"), 10, 8)
		if err != nil {
			return nil, err
		}
		return int8(i), nil
	}
	if strings.HasPrefix(v, "\"") {
		s, err := strconv.Unquote(v)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return nil, err
	}
	return int32(i), nil
}
