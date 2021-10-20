package stalefish

import (
	"fmt"
	"testing"
)

func TestMappingCharFilter_Filter(t *testing.T) {
	tests := []struct {
		mapper map[string]string
		s      string
		want   string
	}{
		{
			mapper: map[string]string{"か": "ka", "き": "ki"},
			s:      "かきくけこ",
			want:   "kakiくけこ",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("mapper = %v, s = %v, want = %v", tt.mapper, tt.s, tt.want), func(t *testing.T) {
			c := MappingCharFilter{
				mapper: tt.mapper,
			}
			if got := c.Filter(tt.s); got != tt.want {
				t.Errorf("MappingCharFilter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
