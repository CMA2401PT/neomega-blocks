package ir

import "testing"

func TestName(t *testing.T) {
	blk := IRBlock("minecraft:block []")
	if blk.NameWithNameSpace() != "minecraft:block" {
		t.FailNow()
	}
	blk = IRBlock("minecraft:block")
	if blk.NameWithNameSpace() != "minecraft:block" {
		t.FailNow()
	}
	blk = IRBlock("minecraft:block")
	if blk.ShortName() != "block" {
		t.FailNow()
	}
	blk = IRBlock("block")
	if blk.ShortName() != "block" {
		t.FailNow()
	}

	blk = IRBlock("minecraft:block []")
	if blk.ShortName() != "block" {
		t.FailNow()
	}
	blk = IRBlock("block []")
	if blk.ShortName() != "block" {
		t.FailNow()
	}

	blk = IRBlock("minecraft:block")
	if blk.NameSpace() != "minecraft" {
		t.FailNow()
	}
	blk = IRBlock("block")
	if blk.NameSpace() != "" {
		t.FailNow()
	}

	blk = IRBlock("minecraft:block []")
	if blk.NameSpace() != "minecraft" {
		t.FailNow()
	}
	blk = IRBlock("block []")
	if blk.NameSpace() != "" {
		t.FailNow()
	}

	blk = IRBlock("block")
	if blk.States() != "[]" {
		t.FailNow()
	}

	blk = IRBlock(`block ["key1"="???"]`)
	if blk.States() != `["key1"="???"]` {
		t.FailNow()
	}
}

func TestIRStatesIsTidy(t *testing.T) {
	cases := []struct {
		states IRStates
		want   bool
	}{
		{states: `[]`, want: true},
		{states: `["a"=1,"b"=2]`, want: true},
		{states: `["a"="x y","b"=false]`, want: true},
		{states: `["a"=1,"a"=2]`, want: true},
		{states: `["b"=2,"a"=1]`, want: false},
		{states: `[ "a"=1]`, want: false},
		{states: `["a" =1]`, want: false},
		{states: `["a"= 1]`, want: false},
		{states: `["a"=1 ]`, want: false},
		{states: `["a"=1, "b"=2]`, want: false},
		{states: `["a"=]`, want: false},
	}

	for _, c := range cases {
		got := c.states.IsTidy()
		if got != c.want {
			t.Fatalf("IsTidy(%q) = %v, want %v", c.states, got, c.want)
		}
	}
}

func TestIRStatesIsValid(t *testing.T) {
	cases := []struct {
		states IRStates
		want   bool
	}{
		{states: `[]`, want: true},
		{states: `[ ]`, want: true},
		{states: `[
		]`, want: true},
		{states: `["a"=1]`, want: true},
		{states: `["b"=2,"a"=1]`, want: true},
		{states: `["a"=1,"b"=true,"c"=false]`, want: true},
		{states: `["a"=1,"b"="x","c"=-12]`, want: true},
		{states: `["a"=1,]`, want: true},
		{states: `["a"=1 , ]`, want: true},
		{states: `["a"=1,"b"=2,]`, want: true},
		{states: ` [ "a" = 1 , "b" = false , ] `, want: true},
		{states: `["a"="x\\\"y"]`, want: true},
		{states: `["a"="x,y"]`, want: true},
		{states: `["a"=""]`, want: true},
		{states: `["a"=-2147483648]`, want: true},
		{states: `["a"=true,"b"=false]`, want: true},
		{states: `["a"=2147483647]`, want: true},
		{states: `["a"=2147483648]`, want: false},
		{states: `["a"=-2147483649]`, want: false},
		{states: `["a"=]`, want: false},
		{states: `["a"=x]`, want: false},
		{states: `["a"=True]`, want: false},
		{states: `["a"=1,,]`, want: false},
		{states: `["a"=1,, "b"=2]`, want: false},
		{states: `["a"=1 "b"=2]`, want: false},
		{states: `["a"=1, "b"=]`, want: false},
		{states: `["a"="x]`, want: false},
		{states: `["a"="x\q"]`, want: false},
		{states: `[a=1]`, want: false},
		{states: `[""=1]`, want: true},
		{states: `["a"=01]`, want: true},
		{states: `a=1`, want: false},
	}

	for _, c := range cases {
		got := c.states.IsValid()
		if got != c.want {
			t.Fatalf("IsValid(%q) = %v, want %v", c.states, got, c.want)
		}
	}
}

func TestIRStatesTidy(t *testing.T) {
	cases := []struct {
		in   IRStates
		want IRStates
	}{
		{in: `[]`, want: `[]`},
		{in: `[ ]`, want: `[]`},
		{in: ` [ ] `, want: ` [ ] `}, // invalid (leading/trailing outside brackets), keep original
		{in: `["b"=2,"a"=1]`, want: `["a"=1,"b"=2]`},
		{in: ` [ "b" = 2 , "a" = 1 ] `, want: ` [ "b" = 2 , "a" = 1 ] `}, // invalid, keep original
		{in: `["b" = 2 , "a" = 1]`, want: `["a"=1,"b"=2]`},
		{in: `[" k e y " = " v a l " ]`, want: `[" k e y "=" v a l "]`},
		{in: `["b"= true ,"a"= false]`, want: `["a"=false,"b"=true]`},
		{in: `["a"=1,]`, want: `["a"=1]`},
	}

	for _, c := range cases {
		got := c.in.Tidy()
		if got != c.want {
			t.Fatalf("Tidy(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIRStatesIterKeyValueString(t *testing.T) {
	states := IRStates(`["a"=1,"b"="x\\\"y","c"=true]`)
	got := ""
	for k, v := range states.KeyValueStrings {
		if got != "" {
			got += ";"
		}
		got += k + "=" + v
	}
	want := `a=1;b="x\\\"y";c=true`
	if got != want {
		t.Fatalf("IterKeyValueString() got %q, want %q", got, want)
	}
}

func TestIRStatesKeyValues(t *testing.T) {
	states := IRStates(`["n"=123,"b1"=true,"b0"=false,"s"="x\\\"y"]`)
	got := map[string]any{}
	states.KeyValues(func(k string, v any) bool {
		got[k] = v
		return true
	})

	if v, ok := got["n"].(int32); !ok || v != 123 {
		t.Fatalf(`key "n" type/value mismatch: %#v`, got["n"])
	}
	if v, ok := got["b1"].(int8); !ok || v != 1 {
		t.Fatalf(`key "b1" type/value mismatch: %#v`, got["b1"])
	}
	if v, ok := got["b0"].(int8); !ok || v != 0 {
		t.Fatalf(`key "b0" type/value mismatch: %#v`, got["b0"])
	}
	if v, ok := got["s"].(string); !ok || v != `x\"y` {
		t.Fatalf(`key "s" type/value mismatch: %#v`, got["s"])
	}
}

func TestIRStatesGet(t *testing.T) {
	states := IRStates(`["n"=123,"b1"=true,"b0"=false,"s"="x\\\"y"]`)

	if v, ok := states.Get("n"); !ok {
		t.Fatalf(`key "n" should be found`)
	} else if iv, ok := v.(int32); !ok || iv != 123 {
		t.Fatalf(`key "n" type/value mismatch: %#v`, v)
	}

	if v, ok := states.Get("b1"); !ok {
		t.Fatalf(`key "b1" should be found`)
	} else if bv, ok := v.(int8); !ok || bv != 1 {
		t.Fatalf(`key "b1" type/value mismatch: %#v`, v)
	}

	if v, ok := states.Get("s"); !ok {
		t.Fatalf(`key "s" should be found`)
	} else if sv, ok := v.(string); !ok || sv != `x\"y` {
		t.Fatalf(`key "s" type/value mismatch: %#v`, v)
	}

	if v, ok := states.Get("missing"); ok || v != nil {
		t.Fatalf(`missing key should return (nil,false), got (%#v,%v)`, v, ok)
	}
}

func TestIRStatesSet(t *testing.T) {
	cases := []struct {
		in    IRStates
		key   string
		value any
		want  IRStates
	}{
		{in: `[]`, key: "a", value: int32(1), want: `["a"=1]`},
		{in: `["b"=2]`, key: "a", value: int32(1), want: `["a"=1,"b"=2]`},
		{in: `["a"=1,"b"=2]`, key: "b", value: int32(3), want: `["a"=1,"b"=3]`},
		{in: `["a"=1,"b"=2,]`, key: "c", value: int32(4), want: `["a"=1,"b"=2,"c"=4]`},
		{in: `["a"=1]`, key: "s", value: `x"y`, want: `["a"=1,"s"="x\"y"]`},
		{in: `["a"=1]`, key: "b1", value: int8(1), want: `["a"=1,"b1"=true]`},
		{in: `["a"=1]`, key: "b0", value: int8(0), want: `["a"=1,"b0"=false]`},
		{in: `["a"=1]`, key: "bb", value: true, want: `["a"=1,"bb"=true]`},
	}
	for _, c := range cases {
		if got := c.in.Set(c.key, c.value); got != c.want {
			t.Fatalf("Set(%q, %q, %#v) = %q, want %q", c.in, c.key, c.value, got, c.want)
		}
	}
}

func TestIRStatesDelete(t *testing.T) {
	cases := []struct {
		in   IRStates
		key  string
		want IRStates
	}{
		{in: `["a"=1,"b"=2]`, key: "a", want: `["b"=2]`},
		{in: `["a"=1,"b"=2]`, key: "b", want: `["a"=1]`},
		{in: `["a"=1]`, key: "a", want: `[]`},
		{in: `["a"=1,"b"=2,]`, key: "a", want: `["b"=2]`},
		{in: `["a"=1]`, key: "missing", want: `["a"=1]`},
	}
	for _, c := range cases {
		if got := c.in.Delete(c.key); got != c.want {
			t.Fatalf("Delete(%q, %q) = %q, want %q", c.in, c.key, got, c.want)
		}
	}
}

func TestIRBlockIsTidy(t *testing.T) {
	cases := []struct {
		in   IRBlock
		want bool
	}{
		{in: IRBlock(`minecraft:block []`), want: true},
		{in: IRBlock(`minecraft:block ["a"=1]`), want: true},
		{in: IRBlock(`block []`), want: false},
		{in: IRBlock(`minecraft:block`), want: false},
		{in: IRBlock(`minecraft:block  []`), want: false},
		{in: IRBlock(`minecraft:block [ ]`), want: false},
		{in: IRBlock(`minecraft: []`), want: false},
	}
	for _, c := range cases {
		if got := c.in.IsTidy(); got != c.want {
			t.Fatalf("IRBlock.IsTidy(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIRBlockTidy(t *testing.T) {
	cases := []struct {
		in   IRBlock
		want IRBlock
	}{
		{in: IRBlock(`minecraft:block []`), want: IRBlock(`minecraft:block []`)},
		{in: IRBlock(`block []`), want: IRBlock(`minecraft:block []`)},
		{in: IRBlock(`block`), want: IRBlock(`minecraft:block []`)},
		{in: IRBlock(`minecraft:block`), want: IRBlock(`minecraft:block []`)},
		{in: IRBlock(`minecraft:block  [ "b" = 2 , "a" = 1 ]`), want: IRBlock(`minecraft:block ["a"=1,"b"=2]`)},
		{in: IRBlock(`minecraft:block ???`), want: IRBlock(`minecraft:block []`)},
	}
	for _, c := range cases {
		if got := c.in.Tidy(); got != c.want {
			t.Fatalf("IRBlock.Tidy(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIRBlockIsValid(t *testing.T) {
	cases := []struct {
		in   IRBlock
		want bool
	}{
		{in: IRBlock(`block`), want: true},
		{in: IRBlock(`block []`), want: true},
		{in: IRBlock(`minecraft:block []`), want: true},
		{in: IRBlock(`minecraft:block ["a"=1]`), want: true},
		{in: IRBlock(`minecraft:block [ ]`), want: true},
		{in: IRBlock(` []`), want: false},
		{in: IRBlock(`minecraft: []`), want: false},
		{in: IRBlock(`block ["a"=1,]`), want: true},
		{in: IRBlock(`block [x]`), want: false},
	}
	for _, c := range cases {
		if got := c.in.IsValid(); got != c.want {
			t.Fatalf("IRBlock.IsValid(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
