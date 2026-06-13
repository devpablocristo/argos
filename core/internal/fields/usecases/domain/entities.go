package domain

import "time"

type Field struct {
	ID         string     `json:"id"`
	OrgID      string     `json:"org_id,omitempty"`
	Name       string     `json:"name"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}
