package v2

import (
	"net/http"
	"regexp"

	"github.com/damoang/angple-backend/internal/common"
	v2repo "github.com/damoang/angple-backend/internal/repository/v2"
	"github.com/gin-gonic/gin"
)

// ContentHandler handles content page API requests
type ContentHandler struct {
	contentRepo v2repo.ContentRepository
}

// NewContentHandler creates a new content handler
func NewContentHandler(contentRepo v2repo.ContentRepository) *ContentHandler {
	return &ContentHandler{contentRepo: contentRepo}
}

// validCoIDPattern allows alphanumeric, underscore, and hyphen only
var validCoIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ListContents returns all content pages
// GET /api/v2/admin/contents
func (h *ContentHandler) ListContents(c *gin.Context) {
	items, err := h.contentRepo.FindAll()
	if err != nil {
		common.V2ErrorResponse(c, http.StatusInternalServerError, "콘텐츠 목록 조회 실패", err)
		return
	}
	common.V2Success(c, items)
}

// GetContent returns a single content page
// GET /api/v2/admin/contents/:id
func (h *ContentHandler) GetContent(c *gin.Context) {
	coID := c.Param("id")
	if !validCoIDPattern.MatchString(coID) {
		common.V2ErrorResponse(c, http.StatusBadRequest, "유효하지 않은 콘텐츠 ID", nil)
		return
	}

	content, err := h.contentRepo.FindByID(coID)
	if err != nil {
		common.V2ErrorResponse(c, http.StatusNotFound, "콘텐츠를 찾을 수 없습니다", err)
		return
	}
	common.V2Success(c, content)
}

// UpdateContent updates a content page
// PUT /api/v2/admin/contents/:id
func (h *ContentHandler) UpdateContent(c *gin.Context) {
	coID := c.Param("id")
	if !validCoIDPattern.MatchString(coID) {
		common.V2ErrorResponse(c, http.StatusBadRequest, "유효하지 않은 콘텐츠 ID", nil)
		return
	}

	content, err := h.contentRepo.FindByID(coID)
	if err != nil {
		common.V2ErrorResponse(c, http.StatusNotFound, "콘텐츠를 찾을 수 없습니다", err)
		return
	}

	var req struct {
		CoSubject       *string `json:"co_subject"`
		CoContent       *string `json:"co_content"`
		CoMobileContent *string `json:"co_mobile_content"`
		CoHTML          *int    `json:"co_html"`
		CoSeoTitle      *string `json:"co_seo_title"`
		CoLevel         *int    `json:"co_level"`
		CoHrefContent   *string `json:"co_href_content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.V2ErrorResponse(c, http.StatusBadRequest, "요청 형식이 올바르지 않습니다", err)
		return
	}

	if req.CoSubject != nil {
		content.CoSubject = *req.CoSubject
	}
	if req.CoContent != nil {
		content.CoContent = *req.CoContent
	}
	if req.CoMobileContent != nil {
		content.CoMobileContent = *req.CoMobileContent
	}
	if req.CoHTML != nil {
		content.CoHTML = *req.CoHTML
	}
	if req.CoSeoTitle != nil {
		content.CoSeoTitle = *req.CoSeoTitle
	}
	if req.CoLevel != nil {
		content.CoLevel = *req.CoLevel
	}
	if req.CoHrefContent != nil {
		content.CoHrefContent = *req.CoHrefContent
	}

	if err := h.contentRepo.Update(content); err != nil {
		common.V2ErrorResponse(c, http.StatusInternalServerError, "콘텐츠 수정 실패", err)
		return
	}

	common.V2Success(c, content)
}

// GetPublicContent returns a single content page for public viewing
// GET /api/v2/contents/:id
func (h *ContentHandler) GetPublicContent(c *gin.Context) {
	coID := c.Param("id")
	if !validCoIDPattern.MatchString(coID) {
		common.V2ErrorResponse(c, http.StatusBadRequest, "유효하지 않은 콘텐츠 ID", nil)
		return
	}

	content, err := h.contentRepo.FindByID(coID)
	if err != nil {
		common.V2ErrorResponse(c, http.StatusNotFound, "콘텐츠를 찾을 수 없습니다", err)
		return
	}

	// Return only public-safe fields
	common.V2Success(c, gin.H{
		"co_id":        content.CoID,
		"co_subject":   content.CoSubject,
		"co_content":   content.CoContent,
		"co_html":      content.CoHTML,
		"co_seo_title": content.CoSeoTitle,
		"co_level":     content.CoLevel,
	})
}
