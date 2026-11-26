package depositconfig

import (
	"context"

	"casino/internal/model"
	"casino/internal/storage"
)

func depositAssetURL(value string) string {
	out, err := storage.PublicURL(context.Background(), value)
	if err != nil {
		return value
	}
	return out
}

func depositAssetRef(value string) string {
	return storage.NormalizeReference(value)
}

func signMethodAsset(row *methodRow) {
	row.IconURL = depositAssetURL(row.IconURL)
}

func signModelMethodAsset(row *model.GamePaymentMethods) {
	row.IconURL = depositAssetURL(row.IconURL)
}

func signVoucherAsset(row *model.GameDepositVouchers) {
	row.IconURL = depositAssetURL(row.IconURL)
}
