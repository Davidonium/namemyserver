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
