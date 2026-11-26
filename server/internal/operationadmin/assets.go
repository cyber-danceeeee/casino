package operationadmin

import (
	"context"

	"casino/internal/model"
	"casino/internal/storage"
)

func operationAssetURL(value string) string {
	out, err := storage.PublicURL(context.Background(), value)
	if err != nil {
		return value
	}
	return out
}

func operationAssetRef(value string) string {
	return storage.NormalizeReference(value)
}

func signBannerAssets(row *bannerRecord) {
	row.ImageURL = operationAssetURL(row.ImageURL)
	row.MobileImageURL = operationAssetURL(row.MobileImageURL)
}

func signVIPAssets(row *vipRecord) {
	row.IconURL = operationAssetURL(row.IconURL)
}

func signSocialAssets(row *socialLinkRecord) {
	row.IconURL = operationAssetURL(row.IconURL)
}

func signMusicAssets(row *model.GameOperationMusic) {
	row.AudioURL = operationAssetURL(row.AudioURL)
}
