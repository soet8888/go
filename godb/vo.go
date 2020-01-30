package db

import "time"

type Model struct {
	ID        uint `json:"id" gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
type Filter struct {
	Field   string      `json:"field"`
	Compare string      `json:"compare"`
	Value   interface{} `json:"value"`
}
