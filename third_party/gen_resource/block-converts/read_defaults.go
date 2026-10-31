package blockconverts

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"neomega-blocks/ir"
)

func ReadDefaults(path string) map[string]ir.IRStates {
	// export_defaults.py 通过库接口导出 SNBT 默认值，避免依赖上游包的内部数据布局。
	payload, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("read defaults file %q: %w", path, err))
	}
	var defaults map[string]map[string]string
	if err := json.Unmarshal(payload, &defaults); err != nil {
		panic(fmt.Errorf("decode defaults file %q: %w", path, err))
	}
	out := make(map[string]ir.IRStates, len(defaults))
	for name, properties := range defaults {
		states := ir.IRStates("[]")
		for k, v := range properties {
			parsed, err := parseSNBTDefaultValue(v)
			if err != nil {
				panic(fmt.Errorf("parse default value in %q block %q key %q: %w", path, name, k, err))
			}
			states = states.Set(k, parsed)
		}

		out[name] = states
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
