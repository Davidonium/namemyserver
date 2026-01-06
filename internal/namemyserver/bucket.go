package namemyserver

import "time"

type Bucket struct {
	ID          int32
	Name        string
	Description string
	Cursor      int32
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	ArchivedAt  *time.Time

	FilterLengthEnabled bool
	FilterLengthMode    LengthMode
	FilterLengthValue   int
}

func (b *Bucket) MarkArchived() {
	n := time.Now()
	b.ArchivedAt = &n
}

func (b *Bucket) Recover() {
	b.ArchivedAt = nil
}

func (b Bucket) Archived() bool {
	return b.ArchivedAt != nil
}

// Filters returns the RandomPairFilters configured for this bucket.
// If length filtering is disabled, returns empty filters.
func (b Bucket) Filters() RandomPairFilters {
	if !b.FilterLengthEnabled {
		return RandomPairFilters{}
	}
	return RandomPairFilters{
		Length:     b.FilterLengthValue,
		LengthMode: b.FilterLengthMode,
	}
}
