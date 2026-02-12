package decoder_bedrock_block_props

import (
	encoder_snbt "neomega-nbt/encoder-snbt"
	"testing"
)

func TestBedrockPropsDecode(t *testing.T) {
	assertSnbt := func(propStr string, snbtStr string) {
		enc, getter := encoder_snbt.NewEncoderToBytes(nil)
		err := Decode(propStr, enc, true)
		if err != nil {
			t.Fatalf("on decoding %v, get error: %v", propStr, err)
		}
		outStr := string(getter())
		if outStr != snbtStr {
			t.Fatalf("on decoding: %v, expect: %v, get: %v", propStr, snbtStr, outStr)
		}
	}
	assertSnbt(`[]`, `{}`)
	assertSnbt(`[""=""]`, `{"": ""}`)
	assertSnbt(`[  ""  =  ""  ]`, `{"": ""}`)
	assertSnbt(`[""="",]`, `{"": ""}`)
	assertSnbt(`[""="",  "1" =""]`, `{"": "", "1": ""}`)
	assertSnbt(`["strKey"="someStr",]`, `{"strKey": "someStr"}`)
	assertSnbt(`["trueKey"=  true ,]`, `{"trueKey": 1b}`)
	assertSnbt(`["trueKey"=  true ,"falseKey"=false]`, `{"trueKey": 1b, "falseKey": 0b}`)
	assertSnbt(`["intKey"=  1 ,]`, `{"intKey": 1}`)
	assertSnbt(`["intKey"=  321 ,]`, `{"intKey": 321}`)
	assertSnbt(`["intKey"=  1,]`, `{"intKey": 1}`)
	assertSnbt(`["intKey"= 1]`, `{"intKey": 1}`)
	assertSnbt(`["strKey"="someStr","trueKey"=  true ,"falseKey"=false,"intKey1"=  1 ,"intKey321"=  321]`, `{"strKey": "someStr", "trueKey": 1b, "falseKey": 0b, "intKey1": 1, "intKey321": 321}`)
}
