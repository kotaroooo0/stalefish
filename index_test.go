package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		documents []Document
		analyzer  Analyzer
		index     Index
	}{
		{
			documents:
			analyzer:  Analyzer{StandardTokenizer{}, []CharFilter{}, []Filter{}},
			index:     []string{},
		},
	}

	for _, tt := range cases {
		if diff := cmp.Diff(tt.tokens, tt.analyzer.Analyze(tt.text)); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

func testDocument1() Document{

}
