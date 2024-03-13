package tagcloud

import (
	"cmp"
	"slices"
)

// TagCloud aggregates statistics about used tags
type TagCloud struct {
	m map[string]int
}

// TagStat represents statistics regarding single tag
type TagStat struct {
	Tag             string
	OccurrenceCount int
}

// New should create a valid TagCloud instance
func New() *TagCloud {
	return &TagCloud{
		m: map[string]int{},
	}
}

// AddTag should add a tag to the cloud if it wasn't present and increase tag occurrence count
// thread-safety is not needed
func (t *TagCloud) AddTag(tag string) {
	t.m[tag]++
}

// TopN should return top N most frequent tags ordered in descending order by occurrence count
// if there are multiple tags with the same occurrence count then the order is defined by implementation
// if n is greater that TagCloud size then all elements should be returned
// thread-safety is not needed
func (t *TagCloud) TopN(n int) []TagStat {
	// TODO: Implement this
	ts := make([]TagStat, len(t.m))
	i := 0
	for k, v := range t.m {
		ts[i] = TagStat{k, v}
		i++
	}
	slices.SortFunc(ts, func(a, b TagStat) int { return cmp.Compare(b.OccurrenceCount, a.OccurrenceCount) })
	return ts[:min(len(t.m), n)]
}
