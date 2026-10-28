package specific_converts

import (
	"neomega-blocks/ir"
	"testing"
)

func TestWallPostConversion(t *testing.T) {
	for _, tc := range []struct {
		states ir.IRStates
		want   string
	}{
		{`["wall_post_bit"=true]`, "true"}, {`["wall_post_bit"=false]`, "false"},
	} {
		converted := tc.states.Update(BedrockToJavaWallRelplacer)
		if got := converted.GetDefault("up", nil); got != tc.want {
			t.Errorf("%s: got %v, want %s", tc.states, got, tc.want)
		}
	}
}
