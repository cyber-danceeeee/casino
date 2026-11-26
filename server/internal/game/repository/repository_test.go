package repository

import (
	"encoding/json"
	"testing"
)

func TestSyncedGameJSONDefaultsAreValid(t *testing.T) {
	for field, value := range map[string]string{
		"name_i18n":          syncedGameNameI18nDefault,
		"category_ids_draft": syncedGameCategoryIDsDraftDefault,
	} {
		if !json.Valid([]byte(value)) {
			t.Fatalf("%s default is invalid JSON: %q", field, value)
		}
	}
}

func TestSyncedJSONOrDefaultRejectsEmptyDocument(t *testing.T) {
	for _, value := range []string{"", "   ", "{", "not-json"} {
		if got := syncedJSONOrDefault(value, "{}"); got != "{}" {
			t.Fatalf("invalid JSON %q normalized to %q", value, got)
		}
	}
	if got := syncedJSONOrDefault(`{"en":"Game"}`, "{}"); got != `{"en":"Game"}` {
		t.Fatalf("valid JSON was changed: %q", got)
	}
}
