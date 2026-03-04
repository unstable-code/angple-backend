package v2

import (
	v2 "github.com/damoang/angple-backend/internal/domain/v2"
	"gorm.io/gorm"
)

// ContentRepository provides access to g5_content table
type ContentRepository interface {
	FindAll() ([]*v2.ContentListItem, error)
	FindByID(coID string) (*v2.Content, error)
	Update(content *v2.Content) error
}

type contentRepository struct {
	db *gorm.DB
}

// NewContentRepository creates a new content repository
func NewContentRepository(db *gorm.DB) ContentRepository {
	return &contentRepository{db: db}
}

// FindAll returns all content pages (list view)
func (r *contentRepository) FindAll() ([]*v2.ContentListItem, error) {
	var items []*v2.ContentListItem
	err := r.db.Model(&v2.Content{}).
		Select("co_id, co_subject, co_seo_title, co_level, co_hit, co_html").
		Order("co_id ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID returns a single content page by ID
func (r *contentRepository) FindByID(coID string) (*v2.Content, error) {
	var content v2.Content
	err := r.db.Where("co_id = ?", coID).First(&content).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// Update updates an existing content page
func (r *contentRepository) Update(content *v2.Content) error {
	return r.db.Save(content).Error
}
