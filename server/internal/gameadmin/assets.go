package gameadmin

import (
	"context"

	"casino/internal/storage"
)

func adminAssetURL(value string) string {
	out, err := storage.PublicURL(context.Background(), value)
	if err != nil {
		return value
	}
	return out
}

func adminAssetRef(value string) string {
	return storage.NormalizeReference(value)
}

func signGameAsset(row *gameRecord) {
	row.ThumbnailURL = adminAssetURL(row.ThumbnailURL)
	row.LocalThumbnailURL = adminAssetURL(row.LocalThumbnailURL)
	row.BackgroundURL = adminAssetURL(row.BackgroundURL)
	row.LaunchImageURL = adminAssetURL(row.LaunchImageURL)
}

func signCategoryAsset(row *categoryRecord) {
	row.Icon = adminAssetURL(row.Icon)
	row.BannerURL = adminAssetURL(row.BannerURL)
}

func signTagAsset(row *tagRecord) {
	row.Icon = adminAssetURL(row.Icon)
}

func signCollectionAsset(row *collectionRecord) {
	row.BannerURL = adminAssetURL(row.BannerURL)
}
