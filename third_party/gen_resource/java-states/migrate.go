package javastates

import (
	"fmt"

	"neomega-blocks/ir"
)

// 转换中心版本之后改名的方块。只记录名称变化，属性照搬。
var blockRenames = map[string]string{
	"minecraft:chain": "minecraft:iron_chain",
}

type propertyRewrite struct {
	from   string
	to     string
	values map[string]string
}

// 转换中心版本之后含义被换掉的属性，按新方块名索引。
// 取值对应关系与游戏自带的数据修复器一致：active=false 的心脏升级后是 uprooted。
var propertyRewrites = map[string][]propertyRewrite{
	"minecraft:creaking_heart": {
		{
			from: "active",
			to:   "creaking_heart_state",
			values: map[string]string{
				"true":  "awake",
				"false": "uprooted",
			},
		},
	},
}

func (r Registry) Migrate(source ir.IRBlock) (ir.IRBlock, error) {
	// 先保留旧属性的含义，再用目标默认值补缺项；未知结果提示维护者补迁移规则。
	name := source.NameWithNameSpace()
	if renamed, ok := blockRenames[name]; ok {
		name = renamed
	}
	properties := source.States().Delete("waterlogged")
	for _, rule := range propertyRewrites[name] {
		if value, ok := properties.Get(rule.from); ok {
			properties = properties.Delete(rule.from).Set(rule.to, rule.values[value.(string)])
		}
	}
	complete := r.Defaults[name]
	for key, value := range properties.KeyValues {
		complete = complete.Set(key, value)
	}
	target := ir.IRBlock(name + " " + string(complete)).Tidy()
	if !r.Known[target] {
		return "", fmt.Errorf(
			"状态 %s 迁移为 %s，但目标全集中不存在；请补充迁移规则",
			source,
			target,
		)
	}
	return target, nil
}
