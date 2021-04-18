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
				Postings:       NewPostings(1, []uint64{0}, 1, nil),
				DocsCount:      1,
				PositionsCount: 1,
			},
			storageInvertedIndex: PostingList{
				Postings:       nil,
				DocsCount:      0,
				PositionsCount: 0,
			},
			expected: PostingList{
				Postings:       NewPostings(1, []uint64{0}, 1, nil),
				DocsCount:      1,
				PositionsCount: 1,
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings:       nil,
				DocsCount:      0,
				PositionsCount: 0,
			},
			storageInvertedIndex: PostingList{
				Postings:       NewPostings(1, []uint64{0}, 1, nil),
				DocsCount:      1,
				PositionsCount: 1,
			},
			expected: PostingList{
				Postings:       NewPostings(1, []uint64{0}, 1, nil),
				DocsCount:      1,
				PositionsCount: 1,
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings:       NewPostings(1, []uint64{0}, 1, NewPostings(3, []uint64{0}, 1, NewPostings(4, []uint64{3}, 1, nil))),
				DocsCount:      3,
				PositionsCount: 3,
			},
			storageInvertedIndex: PostingList{
				Postings:       NewPostings(2, []uint64{1, 2}, 2, NewPostings(4, []uint64{3}, 1, NewPostings(5, []uint64{12}, 1, nil))),
				DocsCount:      3,
				PositionsCount: 4,
			},
			expected: PostingList{
				Postings:       NewPostings(1, []uint64{0}, 1, NewPostings(2, []uint64{1, 2}, 2, NewPostings(3, []uint64{0}, 1, NewPostings(4, []uint64{3}, 1, NewPostings(5, []uint64{12}, 1, nil))))),
				DocsCount:      5,
				PositionsCount: 6,
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings:       NewPostings(3, []uint64{0}, 1, NewPostings(4, []uint64{0}, 1, NewPostings(5, []uint64{3}, 1, nil))),
				DocsCount:      3,
				PositionsCount: 3,
			},
			storageInvertedIndex: PostingList{
				Postings:       NewPostings(1, []uint64{1, 2}, 2, NewPostings(2, []uint64{3}, 1, nil)),
				DocsCount:      2,
				PositionsCount: 3,
			},
			expected: PostingList{
				Postings:       NewPostings(1, []uint64{1, 2}, 2, NewPostings(2, []uint64{3}, 1, NewPostings(3, []uint64{0}, 1, NewPostings(4, []uint64{0}, 1, NewPostings(5, []uint64{3}, 1, nil))))),
				DocsCount:      5,
				PositionsCount: 6,
			},
		},
		{
			memoryInvertedIndex: PostingList{
				Postings:       NewPostings(1, []uint64{0, 4}, 2, nil),
				DocsCount:      1,
				PositionsCount: 2,
			},
			storageInvertedIndex: PostingList{
				Postings:       NewPostings(3, []uint64{0, 1}, 2, nil),
				DocsCount:      1,
				PositionsCount: 2,
			},
			expected: PostingList{
				Postings:       NewPostings(1, []uint64{0, 4}, 2, NewPostings(3, []uint64{0, 1}, 2, nil)),
				DocsCount:      2,
				PositionsCount: 4,
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
