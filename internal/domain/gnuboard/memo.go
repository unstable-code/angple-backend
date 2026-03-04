package gnuboard

import (
	"time"
)

// G5Memo represents the g5_memo table (Gnuboard messages/쪽지)
type G5Memo struct {
	MeID           int       `gorm:"column:me_id;primaryKey;autoIncrement" json:"me_id"`
	MeSendMbID     string    `gorm:"column:me_send_mb_id" json:"me_send_mb_id"`
	MeRecvMbID     string    `gorm:"column:me_recv_mb_id" json:"me_recv_mb_id"`
	MeMemo         string    `gorm:"column:me_memo" json:"me_memo"`
	MeReadDatetime string    `gorm:"column:me_read_datetime" json:"me_read_datetime"`
	MeSendDatetime time.Time `gorm:"column:me_send_datetime" json:"me_send_datetime"`
	MeType         string    `gorm:"column:me_type" json:"me_type"`       // 'recv' or 'send'
	MeSendID       int       `gorm:"column:me_send_id" json:"me_send_id"` // paired message ID
	MeSendIP       string    `gorm:"column:me_send_ip" json:"me_send_ip"`
}

// TableName returns the table name for GORM
func (G5Memo) TableName() string {
	return "g5_memo"
}

// IsRead returns whether the memo has been read
func (m *G5Memo) IsRead() bool {
	return m.MeReadDatetime != "" && m.MeReadDatetime != "0000-00-00 00:00:00"
}
