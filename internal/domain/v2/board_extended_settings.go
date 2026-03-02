package v2

import "time"

// V2BoardExtendedSettings stores flexible JSON-based settings for board features
// that don't map to existing g5_board columns.
// Stored as a single JSON column per board for schema-free extensibility.
type V2BoardExtendedSettings struct {
	BoardID   string    `gorm:"column:board_id;type:varchar(20);primaryKey" json:"board_id"`
	Settings  string    `gorm:"column:settings;type:json" json:"settings"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name
func (V2BoardExtendedSettings) TableName() string { return "v2_board_extended_settings" }
