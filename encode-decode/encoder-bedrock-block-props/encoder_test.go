package encoder_bedrock_block_props

import (
	decoder_shim "neomega-nbt/decoder-shim"
	decoder_snbt "neomega-nbt/decoder-snbt"
	"testing"
)

func TestBedrockPropsDecode(t *testing.T) {
	assertSnbt := func(propStr string, snbtStr string) {
		enc, getter := NewEncoderToBytes(nil)
		_, err := decoder_snbt.DecodeSnbtTo([]uint8(snbtStr), decoder_shim.NewDecoderShim(enc), decoder_snbt.DecodeOption{})
		if err != nil {
			t.Fatalf("on decoding %v, get error: %v", snbtStr, err)
		}
		outStr := string(getter())
		if outStr != propStr {
			t.Fatalf("on decoding: %v, expect: %v, get: %v", snbtStr, propStr, outStr)
		}
	}
	assertSnbt(`[]`, `{}`)
	assertSnbt(`["key1"=""]`, `{"key1": ""}`)
	assertSnbt(`[""="","1"=""]`, `{"": "", "1": ""}`)
	assertSnbt(`["trueKey"=true]`, `{"trueKey": 1b}`)
	assertSnbt(`["trueKey"=true,"falseKey"=false]`, `{"trueKey": 1b, "falseKey": 0b}`)
	assertSnbt(`["intKey"=0]`, `{"intKey": 0}`)
	assertSnbt(`["intKey"=1]`, `{"intKey": 1}`)
	assertSnbt(`["intKey"=321]`, `{"intKey": 321}`)
	assertSnbt(`["strKey"="someStr","trueKey"=true,"falseKey"=false,"intKey1"=1,"intKey321"=321]`, `{"strKey": "someStr", "trueKey": 1b, "falseKey": 0b, "intKey1": 1, "intKey321": 321}`)
}
