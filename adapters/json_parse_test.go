package adapters_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink/adapters"
	"github.com/smartcontractkit/chainlink/internal/cltest"
	"github.com/stretchr/testify/assert"
)

func TestJsonParse_Perform(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		value           string
		path            []string
		want            string
		wantError       bool
		wantResultError bool
	}{
		{"existing path", `{"high": "11850.00", "last": "11779.99"}`, []string{"last"}, "11779.99", false, false},
		{"nonexistent path", `{"high": "11850.00", "last": "11779.99"}`, []string{"doesnotexist"}, "", true, false},
		{"double nonexistent path", `{"high": "11850.00", "last": "11779.99"}`, []string{"no", "really"}, "", true, true},
		{"array index path", `{"data": [{"availability": "0.99991"}]}`, []string{"data", "0", "availability"}, "0.99991", false, false},
		{"float value", `{"availability": 0.99991}`, []string{"availability"}, "0.99991", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := cltest.RunResultWithValue(test.value)
			adapter := adapters.JsonParse{Path: test.path}
			result := adapter.Perform(input, nil)
			val, err := result.Get("value")
			assert.Nil(t, err)
			assert.Equal(t, test.want, val.String())

			if test.wantResultError {
				assert.NotNil(t, result.GetError())
			} else {
				assert.Nil(t, result.GetError())
			}
		})
	}
}
