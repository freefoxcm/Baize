package serve

import (
	"reflect"
	"testing"
)

func TestUsageCalendarLevels(t *testing.T) {
	for _, tc := range []struct {
		name       string
		tokens     []int64
		wantMax    int64
		wantLevels []int
	}{
		{name: "empty", tokens: nil, wantLevels: []int{}},
		{name: "inactive", tokens: []int64{0, 0}, wantLevels: []int{0, 0}},
		{name: "single active", tokens: []int64{0, 50}, wantMax: 50, wantLevels: []int{0, 5}},
		{name: "linear levels", tokens: []int64{1, 20, 21, 40, 41, 60, 61, 80, 81, 100}, wantMax: 100, wantLevels: []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5}},
		{name: "outlier capped", tokens: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 1000}, wantMax: 19, wantLevels: []int{1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 5}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scale, levels := usageCalendarLevels(tc.tokens)
			if scale.Method != usageCalendarScaleMethod || scale.Levels != usageCalendarScaleLevels || scale.MaxTokens != tc.wantMax {
				t.Fatalf("scale = %+v, want method %q levels %d max %d", scale, usageCalendarScaleMethod, usageCalendarScaleLevels, tc.wantMax)
			}
			if tc.wantLevels == nil {
				tc.wantLevels = []int{}
			}
			if levels == nil {
				levels = []int{}
			}
			if !reflect.DeepEqual(levels, tc.wantLevels) {
				t.Fatalf("levels = %v, want %v", levels, tc.wantLevels)
			}
		})
	}
}
