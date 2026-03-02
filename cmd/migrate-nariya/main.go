package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/damoang/angple-backend/internal/config"
	v2domain "github.com/damoang/angple-backend/internal/domain/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	configPath := flag.String("config", "configs/config.dev.yaml", "config file path")
	nariyaPath := flag.String("nariya-path", "/home/damoang/www/data/nariya/board", "nariya board data directory")
	dryRun := flag.Bool("dry-run", false, "parse and show results without writing to DB")
	verbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()

	loaded := config.LoadDotEnv()
	if len(loaded) > 0 {
		log.Printf("Loaded env files: %v", loaded)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logLevel := gormlogger.Warn
	if *verbose {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v", err)
	}
	defer sqlDB.Close()

	// Ensure table exists
	if err := db.AutoMigrate(&v2domain.V2BoardExtendedSettings{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Scan nariya PHP files
	pattern := filepath.Join(*nariyaPath, "board-*-pc.php")
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("Failed to glob: %v", err)
	}

	log.Printf("Found %d nariya board files in %s", len(files), *nariyaPath)

	var successCount, skipCount, errorCount int

	for _, filePath := range files {
		// Extract board ID from filename: board-{id}-pc.php
		base := filepath.Base(filePath)
		boardID := extractBoardID(base)
		if boardID == "" {
			log.Printf("[SKIP] Cannot extract board ID from: %s", base)
			skipCount++
			continue
		}

		// Parse PHP file
		phpData, err := parsePHPFile(filePath)
		if err != nil {
			log.Printf("[ERROR] %s: %v", boardID, err)
			errorCount++
			continue
		}

		// Convert to extended settings JSON
		settings := convertToExtendedSettings(phpData)

		settingsJSON, err := json.Marshal(settings)
		if err != nil {
			log.Printf("[ERROR] %s: failed to marshal JSON: %v", boardID, err)
			errorCount++
			continue
		}

		if *dryRun {
			if *verbose {
				log.Printf("[DRY-RUN] %s: %s", boardID, string(settingsJSON))
			} else {
				log.Printf("[DRY-RUN] %s: OK (%d PHP keys → JSON)", boardID, len(phpData))
			}
			successCount++
			continue
		}

		// Upsert to database
		record := &v2domain.V2BoardExtendedSettings{
			BoardID:  boardID,
			Settings: string(settingsJSON),
		}

		result := db.Where("board_id = ?", boardID).First(&v2domain.V2BoardExtendedSettings{})
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(record).Error; err != nil {
				log.Printf("[ERROR] %s: insert failed: %v", boardID, err)
				errorCount++
				continue
			}
		} else if result.Error != nil {
			log.Printf("[ERROR] %s: query failed: %v", boardID, result.Error)
			errorCount++
			continue
		} else {
			// Merge with existing settings
			var existing v2domain.V2BoardExtendedSettings
			db.Where("board_id = ?", boardID).First(&existing)

			merged := mergeSettings(existing.Settings, string(settingsJSON))
			if err := db.Model(&v2domain.V2BoardExtendedSettings{}).
				Where("board_id = ?", boardID).
				Update("settings", merged).Error; err != nil {
				log.Printf("[ERROR] %s: update failed: %v", boardID, err)
				errorCount++
				continue
			}
		}

		if *verbose {
			log.Printf("[OK] %s", boardID)
		}
		successCount++
	}

	log.Printf("Done: %d success, %d skipped, %d errors (total %d files)",
		successCount, skipCount, errorCount, len(files))
}

// extractBoardID extracts board ID from filename like "board-free-pc.php"
func extractBoardID(filename string) string {
	// board-{id}-pc.php
	re := regexp.MustCompile(`^board-(.+)-pc\.php$`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

// parsePHPFile parses a nariya PHP config file and returns key-value pairs
func parsePHPFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	data := make(map[string]string)

	// Match PHP array entries: 'key' => 'value',
	re := regexp.MustCompile(`'([^']+)'\s*=>\s*'([^']*)'`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	for _, m := range matches {
		if len(m) == 3 {
			data[m[1]] = m[2]
		}
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no key-value pairs found")
	}

	return data, nil
}

// Extended settings JSON structure matching the TypeScript interfaces
type extendedSettings struct {
	Comment      *commentSettings      `json:"comment,omitempty"`
	Lucky        *luckySettings        `json:"lucky,omitempty"`
	XP           *xpSettings           `json:"xp,omitempty"`
	Features     *featureSettings      `json:"features,omitempty"`
	Notification *notificationSettings `json:"notification,omitempty"`
	Writing      *writingSettings      `json:"writing,omitempty"`
	Skin         *skinSettings         `json:"skin,omitempty"`
	Promotion    *promotionSettings    `json:"promotion,omitempty"`
}

type commentSettings struct {
	UseRecommend    bool   `json:"useRecommend"`
	UseDislike      bool   `json:"useDislike"`
	AuthorOnly      bool   `json:"authorOnly"`
	Paging          string `json:"paging"`
	PageSize        int    `json:"pageSize"`
	ImageSizeLimitMB int   `json:"imageSizeLimitMB"`
	AutoEmbed       bool   `json:"autoEmbed"`
}

type luckySettings struct {
	Points int `json:"points"`
	Odds   int `json:"odds"`
}

type xpSettings struct {
	Write   int `json:"write"`
	Comment int `json:"comment"`
}

type featureSettings struct {
	CodeHighlighter     bool   `json:"codeHighlighter"`
	ExternalImageSave   bool   `json:"externalImageSave"`
	TagLevel            int    `json:"tagLevel"`
	Rating              bool   `json:"rating"`
	MobileEditor        string `json:"mobileEditor"`
	CategoryMovePermit  string `json:"categoryMovePermit,omitempty"`
	CategoryMoveMessage string `json:"categoryMoveMessage,omitempty"`
	HideNickname        bool   `json:"hideNickname"`
}

type notificationSettings struct {
	NewPostReceivers string `json:"newPostReceivers"`
	Enabled          bool   `json:"enabled"`
}

type writingSettings struct {
	MaxPosts          int    `json:"maxPosts"`
	AllowedLevels     string `json:"allowedLevels"`
	RestrictedUsers   bool   `json:"restrictedUsers"`
	MemberOnly        bool   `json:"memberOnly"`
	MemberOnlyPermit  string `json:"memberOnlyPermit,omitempty"`
	AllowedMembersOne   string `json:"allowedMembersOne"`
	AllowedMembersTwo   string `json:"allowedMembersTwo"`
	AllowedMembersThree string `json:"allowedMembersThree"`
}

type skinSettings struct {
	Category string `json:"category"`
	List     string `json:"list"`
	View     string `json:"view"`
	Comment  string `json:"comment"`
}

type promotionSettings struct {
	InsertIndex *int `json:"insertIndex"`
	InsertCount *int `json:"insertCount"`
	MinPostCount *int `json:"minPostCount"`
}

func convertToExtendedSettings(php map[string]string) *extendedSettings {
	s := &extendedSettings{}

	// Comment settings
	commentPaging := "oldest"
	if php["comment_sort"] == "new" {
		commentPaging = "newest"
	}
	s.Comment = &commentSettings{
		UseRecommend:    php["comment_good"] == "1",
		UseDislike:      false,
		AuthorOnly:      php["author_only_comment"] == "1",
		Paging:          commentPaging,
		PageSize:        atoi(php["comment_rows"], 5000),
		ImageSizeLimitMB: atoi(php["comment_image_size"], 0),
		AutoEmbed:       php["comment_convert"] == "1",
	}

	// Lucky settings
	s.Lucky = &luckySettings{
		Points: atoi(php["lucky_point"], 0),
		Odds:   atoi(php["lucky_dice"], 0),
	}

	// XP settings
	s.XP = &xpSettings{
		Write:   atoi(php["xp_write"], 0),
		Comment: atoi(php["xp_comment"], 0),
	}

	// Feature settings
	s.Features = &featureSettings{
		CodeHighlighter:     php["code"] == "1",
		ExternalImageSave:   php["save_image"] == "1",
		TagLevel:            atoi(php["tag"], 0),
		Rating:              php["check_star_rating"] == "1",
		MobileEditor:        php["editor_mo"],
		CategoryMovePermit:  php["category_move_permit"],
		CategoryMoveMessage: php["category_move_message"],
		HideNickname:        php["check_list_hide_profile"] == "1",
	}

	// Notification settings
	notiMb := php["noti_mb"]
	s.Notification = &notificationSettings{
		NewPostReceivers: notiMb,
		Enabled:          php["noti_no"] != "1", // noti_no=1 means disabled
	}

	// Writing settings
	s.Writing = &writingSettings{
		MaxPosts:          atoi(php["limit_max_write"], 0),
		AllowedLevels:     php["writeable_level"],
		RestrictedUsers:   php["check_write_permit"] == "1",
		MemberOnly:        php["check_member_only"] == "1",
		MemberOnlyPermit:  php["member_only_permit"],
		AllowedMembersOne:   php["bo_write_allow_one"],
		AllowedMembersTwo:   php["bo_write_allow_two"],
		AllowedMembersThree: php["bo_write_allow_three"],
	}

	// Skin settings
	s.Skin = &skinSettings{
		Category: php["cate_skin"],
		List:     php["list_skin"],
		View:     php["view_skin"],
		Comment:  php["comment_skin"],
	}

	return s
}

func atoi(s string, defaultVal int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// mergeSettings merges nariya settings into existing DB settings
// Existing keys that nariya doesn't set are preserved
func mergeSettings(existingJSON, nariyaJSON string) string {
	existing := make(map[string]interface{})
	nariya := make(map[string]interface{})

	if err := json.Unmarshal([]byte(existingJSON), &existing); err != nil {
		return nariyaJSON
	}
	if err := json.Unmarshal([]byte(nariyaJSON), &nariya); err != nil {
		return existingJSON
	}

	// Deep merge: nariya values overwrite existing, but preserve keys not in nariya
	for key, nariyaVal := range nariya {
		existingVal, exists := existing[key]
		if !exists {
			existing[key] = nariyaVal
			continue
		}

		// If both are maps, merge recursively
		existingMap, existingIsMap := existingVal.(map[string]interface{})
		nariyaMap, nariyaIsMap := nariyaVal.(map[string]interface{})
		if existingIsMap && nariyaIsMap {
			for k, v := range nariyaMap {
				existingMap[k] = v
			}
			existing[key] = existingMap
		} else {
			existing[key] = nariyaVal
		}
	}

	result, err := json.Marshal(existing)
	if err != nil {
		return nariyaJSON
	}
	return string(result)
}
