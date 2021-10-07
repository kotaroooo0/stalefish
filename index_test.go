package stalefish

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMerge(t *testing.T) {
	t.Parallel()
	cases := []struct {
		memoryInvertedIndex  PostingList
		storageInvertedIndex PostingList
		expected             PostingList
	}{
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
			storageInvertedIndex: PostingList{
				Postings: nil,
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: nil,
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0}, nil),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0}, NewPostings(3, []uint64{0}, NewPostings(4, []uint64{3}, nil))),
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(2, []uint64{1, 2}, NewPostings(4, []uint64{3}, NewPostings(5, []uint64{12}, nil))),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0}, NewPostings(2, []uint64{1, 2}, NewPostings(3, []uint64{0}, NewPostings(4, []uint64{3}, NewPostings(5, []uint64{12}, nil))))),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(3, []uint64{0}, NewPostings(4, []uint64{0}, NewPostings(5, []uint64{3}, nil))),
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{1, 2}, NewPostings(2, []uint64{3}, nil)),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{1, 2}, NewPostings(2, []uint64{3}, NewPostings(3, []uint64{0}, NewPostings(4, []uint64{0}, NewPostings(5, []uint64{3}, nil))))),
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings: NewPostings(1, []uint64{0, 4}, nil),
			},
			storageInvertedIndex: PostingList{
				Postings: NewPostings(3, []uint64{0, 1}, nil),
			},
			expected: PostingList{
				Postings: NewPostings(1, []uint64{0, 4}, NewPostings(3, []uint64{0, 1}, nil)),
			},
		},
	}
	for _, tt := range cases {
		merged := tt.memoryInvertedIndex.Merge(tt.storageInvertedIndex)
		if diff := cmp.Diff(merged, tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}

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
		tt.postings.PushBack(tt.inserted)
		if diff := cmp.Diff(tt.postings, tt.expected); diff != "" {
			t.Errorf("Diff: (-got +want)\n%s", diff)
		}
	}
}
