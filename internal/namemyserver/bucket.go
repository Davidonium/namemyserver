package namemyserver

import "time"

type Bucket struct {
	ID          int32
	Name        string
	Description string
	Cursor      int32
	ArchivedAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (b *Bucket) MarkArchived() {
	b.ArchivedAt = time.Now()
}

func (b Bucket) Archived() bool {
	return !b.ArchivedAt.IsZero()
}
