package v2

import "time"

// Content represents a static page in g5_content table (Gnuboard legacy)
type Content struct {
	CoID            string    `gorm:"column:co_id;primaryKey;type:varchar(20)" json:"co_id"`
	CoHTML          int       `gorm:"column:co_html;type:tinyint(4);default:1" json:"co_html"`
	CoSubject       string    `gorm:"column:co_subject;type:varchar(255)" json:"co_subject"`
	CoContent       string    `gorm:"column:co_content;type:longtext" json:"co_content"`
	CoSeoTitle      string    `gorm:"column:co_seo_title;type:varchar(255)" json:"co_seo_title"`
	CoMobileContent string    `gorm:"column:co_mobile_content;type:longtext" json:"co_mobile_content"`
	CoSkin          string    `gorm:"column:co_skin;type:varchar(255)" json:"co_skin"`
	CoMobileSkin    string    `gorm:"column:co_mobile_skin;type:varchar(255)" json:"co_mobile_skin"`
	CoTagFilterUse  int       `gorm:"column:co_tag_filter_use;type:tinyint(4);default:0" json:"co_tag_filter_use"`
	CoHit           int       `gorm:"column:co_hit;type:int(11);default:0" json:"co_hit"`
	CoIncludeHead   string    `gorm:"column:co_include_head;type:text" json:"co_include_head"`
	CoIncludeTail   string    `gorm:"column:co_include_tail;type:text" json:"co_include_tail"`
	CoHrefContent   string    `gorm:"column:co_hrefContent;type:varchar(255)" json:"co_href_content"`
	CoLevel         int       `gorm:"column:co_level;type:tinyint(4);default:0" json:"co_level"`
	CreatedAt       time.Time `gorm:"column:co_created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:co_updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for g5_content
func (Content) TableName() string {
	return "g5_content"
}

// ContentListItem is a minimal representation for list views
type ContentListItem struct {
	CoID       string `json:"co_id"`
	CoSubject  string `json:"co_subject"`
	CoSeoTitle string `json:"co_seo_title"`
	CoLevel    int    `json:"co_level"`
	CoHit      int    `json:"co_hit"`
	CoHTML     int    `json:"co_html"`
}

// UpdateContentRequest is the request body for updating a content page
type UpdateContentRequest struct {
	CoSubject       *string `json:"co_subject"`
	CoContent       *string `json:"co_content"`
	CoMobileContent *string `json:"co_mobile_content"`
	CoHTML          *int    `json:"co_html"`
	CoSeoTitle      *string `json:"co_seo_title"`
	CoLevel         *int    `json:"co_level"`
	CoHrefContent   *string `json:"co_href_content"`
}
