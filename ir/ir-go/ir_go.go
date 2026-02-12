package ir_go

// import (
// 	"neomega-blocks/ir/ir-go/prop_value"
// 	"sort"
// 	"strings"
// )

// type IRGoProp struct {
// 	Key   string
// 	Value prop_value.PropValue
// }

// type IRGoBlock struct {
// 	// should be name with namespace
// 	nameWithNameSpace string
// 	// should always be sorted, so cannot access directly
// 	props []IRGoProp
// }

// func (b *IRGoBlock) sort() {
// 	if b == nil || len(b.props) <= 1 {
// 		return
// 	}
// 	sort.Slice(b.props, func(i, j int) bool {
// 		return b.props[i].Key < b.props[j].Key
// 	})
// }

// func (b IRGoBlock) NameSpace() string {
// 	idx := strings.Index(b.nameWithNameSpace, ":")
// 	if idx == -1 {
// 		return ""
// 	}
// 	return b.nameWithNameSpace[:idx]
// }

// func (b IRGoBlock) FullName() string {
// 	return b.nameWithNameSpace
// }

// func (b IRGoBlock) ShortName() string {
// 	idx := strings.Index(b.nameWithNameSpace, ":")
// 	if idx == -1 {
// 		return b.nameWithNameSpace
// 	}
// 	return b.nameWithNameSpace[idx+1:]
// }

// func (b IRGoBlock) SortedProps() []IRGoProp {
// 	return b.props
// }

// func (b IRGoBlock) String() string {
// 	builder := &strings.Builder{}
// 	builder.WriteString(b.FullName())
// 	builder.WriteByte(' ')
// 	builder.WriteByte('[')
// 	for i, prop := range b.props {
// 		if i != 0 {
// 			builder.WriteByte(',')
// 		}
// 		builder.WriteByte('"')
// 		builder.WriteString(prop.Key)
// 		builder.WriteByte('"')
// 		builder.WriteByte('=')
// 		builder.WriteString(prop.Value.BedrockString())
// 	}
// 	builder.WriteByte(']')
// 	return builder.String()
// }
