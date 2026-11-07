package javastates

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"neomega-blocks/ir"
)

type ReportBlock struct {
	Properties map[string][]string `json:"properties"`
	States     []ReportState       `json:"states"`
}

type ReportState struct {
	ID         int               `json:"id"`
	Default    bool              `json:"default"`
	Properties map[string]string `json:"properties"`
}

type Registry struct {
	// 报告中的注册顺序用于状态枚举，默认状态用于跨版本补齐属性。
	States   []ir.IRBlock
	Defaults map[string]ir.IRStates
	Known    map[ir.IRBlock]bool
	Schema   []string
}

func ParseReport(data []byte) (Registry, error) {
	// 含水由独立层表达，因此状态全集、默认值和属性域统一去掉 waterlogged。
	var report map[string]ReportBlock
	if err := json.Unmarshal(data, &report); err != nil {
		return Registry{}, err
	}
	type registeredState struct {
		id    int
		block ir.IRBlock
	}
	var registered []registeredState
	result := Registry{
		Defaults: make(map[string]ir.IRStates),
		Known:    make(map[ir.IRBlock]bool),
	}
	for _, name := range slices.Sorted(maps.Keys(report)) {
		definition := report[name]
		for _, state := range definition.States {
			properties := ir.IRStates("[]")
			for key, value := range state.Properties {
				if key != "waterlogged" {
					properties = properties.Set(key, value)
				}
			}
			block := ir.IRBlock(name + " " + string(properties)).Tidy()
			registered = append(registered, registeredState{id: state.ID, block: block})
			if state.Default {
				result.Defaults[name] = block.States()
			}
		}
		fields := []string{name}
		for _, key := range slices.Sorted(maps.Keys(definition.Properties)) {
			if key == "waterlogged" {
				continue
			}
			var values []string
			for _, value := range definition.Properties[key] {
				values = append(values, strconv.Quote(value))
			}
			value, _ := result.Defaults[name].Get(key)
			field := fmt.Sprintf(
				"%s=%s|%s",
				key,
				strconv.Quote(value.(string)),
				strings.Join(values, ","),
			)
			fields = append(fields, field)
		}
		result.Schema = append(result.Schema, strings.Join(fields, "\t"))
	}
	slices.SortFunc(registered, func(a, b registeredState) int { return a.id - b.id })
	for _, state := range registered {
		if !result.Known[state.block] {
			result.Known[state.block] = true
			result.States = append(result.States, state.block)
		}
	}
	return result, nil
}
