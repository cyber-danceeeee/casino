package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"mime"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"casino/configs"
	"casino/internal/config"
	"casino/internal/database"
	"casino/internal/model"
	"casino/internal/storage"
)

type gameRow struct {
	ID               uint64 `gorm:"column:id"`
	ProviderGameCode string `gorm:"column:provider_game_code"`
	Name             string `gorm:"column:name"`
	Category         string `gorm:"column:category"`
}

type imageFile struct {
	Path      string
	Category  string
	Stem      string
	Extension string
}

type match struct {
	Game  gameRow
	Image imageFile
}

func main() {
	var configFile string
	var sourceRoot string
	var apply bool
	var concurrency int
	flag.StringVar(&configFile, "c", configs.Path("casino.yml"), "configuration file")
	flag.StringVar(&sourceRoot, "source", "/Users/wimmer/Downloads/AG-游戏图片", "AG image directory")
	flag.BoolVar(&apply, "apply", false, "upload matched images and update the database")
	flag.IntVar(&concurrency, "concurrency", 8, "maximum concurrent OSS uploads")
	flag.Parse()

	if err := run(configFile, sourceRoot, apply, concurrency); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "AG thumbnail import failed: %v\n", err)
		os.Exit(1)
	}
}

func run(configFile, sourceRoot string, apply bool, concurrency int) error {
	if err := config.Init(configFile); err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(config.StorageConfig().Driver), "oss") {
		return errors.New("storage.driver must be oss")
	}
	if concurrency < 1 {
		concurrency = 1
	}
	database.InitDB()
	defer func() { _ = database.CloseDB() }()

	games, err := loadGames(database.GetDB())
	if err != nil {
		return err
	}
	images, err := collectImages(sourceRoot)
	if err != nil {
		return err
	}
	matches, unmatched := matchImages(games, images)
	fmt.Printf("AG games=%d images=%d matched=%d unmatched=%d apply=%v\n", len(games), len(images), len(matches), len(unmatched), apply)
	for _, game := range unmatched {
		fmt.Printf("[unmatched] id=%d category=%s providerGameCode=%s name=%q\n", game.ID, game.Category, game.ProviderGameCode, game.Name)
	}
	if !apply {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(concurrency)
	var uploaded atomic.Int64
	var failed atomic.Int64
	var failuresMu sync.Mutex
	failures := make([]string, 0)
	for _, item := range matches {
		item := item
		group.Go(func() error {
			if err := uploadAndUpdate(groupCtx, database.GetDB(), item); err != nil {
				failed.Add(1)
				failuresMu.Lock()
				failures = append(failures, fmt.Sprintf("%s: %v", item.Game.ProviderGameCode, err))
				failuresMu.Unlock()
				return nil
			}
			count := uploaded.Add(1)
			if count%50 == 0 || count == int64(len(matches)) {
				fmt.Printf("uploaded and updated %d/%d\n", count, len(matches))
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}
	sort.Strings(failures)
	for _, failure := range failures {
		fmt.Printf("[failed] %s\n", failure)
	}
	fmt.Printf("done: uploaded=%d failed=%d unmatched=%d\n", uploaded.Load(), failed.Load(), len(unmatched))
	if failed.Load() > 0 {
		return fmt.Errorf("%d thumbnail uploads failed", failed.Load())
	}
	return nil
}

func loadGames(db *gorm.DB) ([]gameRow, error) {
	var games []gameRow
	err := db.Table("game_games g").
		Select("g.id,g.provider_game_code,g.name,g.category").
		Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").
		Where("p.provider_code=? AND g.deleted_at IS NULL", "ag").
		Order("g.id ASC").
		Scan(&games).Error
	return games, err
}

func collectImages(root string) ([]imageFile, error) {
	images := make([]imageFile, 0)
	err := filepath.WalkDir(root, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if !supportedImageExtension(extension) {
			return nil
		}
		relative, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		if len(parts) != 2 || !strings.HasPrefix(strings.ToUpper(parts[0]), "AG_HOT_") {
			return nil
		}
		images = append(images, imageFile{
			Path:      filePath,
			Category:  parts[0],
			Stem:      strings.TrimSuffix(parts[1], filepath.Ext(parts[1])),
			Extension: extension,
		})
		return nil
	})
	return images, err
}

func matchImages(games []gameRow, images []imageFile) ([]match, []gameRow) {
	byKey := make(map[string]imageFile, len(images))
	for _, image := range images {
		key := imageKey(image.Category, image.Stem)
		current, exists := byKey[key]
		if !exists || preferredImage(image, current) {
			byKey[key] = image
		}
	}
	matches := make([]match, 0, len(games))
	unmatched := make([]gameRow, 0)
	for _, game := range games {
		candidates := []string{
			game.ProviderGameCode,
			providerPrefix(game.Category) + "_" + removeSpaces(game.Name),
			game.Name,
			removeSpaces(game.Name),
		}
		var image imageFile
		found := false
		for _, candidate := range candidates {
			if candidate = strings.TrimSpace(candidate); candidate == "" {
				continue
			}
			if value, ok := byKey[imageKey(game.Category, candidate)]; ok {
				image = value
				found = true
				break
			}
		}
		if !found {
			unmatched = append(unmatched, game)
			continue
		}
		matches = append(matches, match{Game: game, Image: image})
	}
	return matches, unmatched
}

func uploadAndUpdate(ctx context.Context, db *gorm.DB, item match) error {
	file, err := os.Open(item.Image.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	contentType := mime.TypeByExtension(item.Image.Extension)
	if item.Image.Extension == ".avif" {
		contentType = "image/avif"
	}
	objectName := item.Game.ProviderGameCode + item.Image.Extension
	objectKey := path.Join("game-thumbnails", "ag", item.Game.Category, objectName)
	asset, err := storage.Put(ctx, objectKey, file, info.Size(), contentType)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&model.GameGames{}).
		Where("id=? AND deleted_at IS NULL", item.Game.ID).
		Updates(map[string]any{
			"cover_url":           asset.Key,
			"local_thumbnail_url": asset.Key,
			"updated_at":          time.Now(),
		}).Error
}

func imageKey(category, stem string) string {
	return strings.ToLower(strings.TrimSpace(category) + "/" + strings.TrimSpace(stem))
}

func providerPrefix(category string) string {
	return strings.Replace(strings.TrimSpace(category), "_HOT_", "_", 1)
}

func removeSpaces(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, value)
}

func preferredImage(candidate, current imageFile) bool {
	return imagePriority(candidate.Extension) > imagePriority(current.Extension)
}

func imagePriority(extension string) int {
	switch strings.ToLower(extension) {
	case ".avif":
		return 5
	case ".webp":
		return 4
	case ".png":
		return 3
	case ".jpg", ".jpeg":
		return 2
	default:
		return 0
	}
}

func supportedImageExtension(extension string) bool {
	return imagePriority(extension) > 0
}
