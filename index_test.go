package stalefish

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPushBack(t *testing.T) {
	t.Parallel()
	cases := []struct {
		postings *Postings
		inserted *Postings
		expected *Postings
	}{
		{
			postings: NewPostings(1, []uint64{}, NewPostings(2, []uint64{}, NewPostings(3, []uint64{}, NewPostings(4, []uint64{}, nil)))),
			inserted: NewPostings(0, []uint64{}, nil),
			expected: NewPostings(1, []uint64{}, NewPostings(0, []uint64{}, NewPostings(2, []uint64{}, NewPostings(3, []uint64{}, NewPostings(4, []uint64{}, nil))))),
		},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("postings = %v, inserted = %v, expected = %v,", tt.postings, tt.inserted, tt.expected), func(t *testing.T) {
			// When
			tt.postings.PushBack(tt.inserted)

			// Then
			if diff := cmp.Diff(tt.postings, tt.expected); diff != "" {
				t.Errorf("Diff: (-got +want)\n%s", diff)
			}
		})
	}
}
