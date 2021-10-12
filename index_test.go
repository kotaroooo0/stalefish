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

func TestAppearanceCountInDocument(t *testing.T) {
	ps := NewPostings(1, []uint64{0}, NewPostings(2, []uint64{1, 2}, NewPostings(3, []uint64{3, 4, 5}, nil)))
	tests := []struct {
		postings *Postings
		docID    DocumentID
		want     int
	}{
		{
			postings: ps,
			docID:    1,
			want:     1,
		},
		{
			postings: ps,
			docID:    2,
			want:     2,
		},
		{
			postings: ps,
			docID:    3,
			want:     3,
		},
		{
			postings: ps,
			docID:    4,
			want:     0,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("postings = %v, docID = %v, want = %v", tt.postings, tt.docID, tt.want), func(t *testing.T) {
			p := PostingList{
				Postings: tt.postings,
			}
			if got := p.AppearanceCountInDocument(tt.docID); got != tt.want {
				t.Errorf("PostingList.AppearanceCountInDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}
