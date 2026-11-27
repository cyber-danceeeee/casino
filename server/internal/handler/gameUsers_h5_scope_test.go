package handler

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"casino/internal/content"
)

func TestH5GameUserPhoneIsUniqueWithinSiteScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE game_users (id INTEGER PRIMARY KEY, site_id INTEGER NOT NULL, phone_country_code TEXT NOT NULL, phone TEXT NOT NULL, deleted_at DATETIME, UNIQUE (site_id, phone_country_code, phone))").Error)

	insert := func(id, siteID uint64, countryCode, phone string) error {
		return db.Exec("INSERT INTO game_users (id, site_id, phone_country_code, phone) VALUES (?, ?, ?, ?)", id, siteID, countryCode, phone).Error
	}

	require.NoError(t, insert(1, 0, "+92", "3088888888"))
	require.NoError(t, insert(2, 3, "+92", "3088888888"))
	require.NoError(t, insert(3, 5, "+92", "3088888888"))
	require.NoError(t, insert(4, 4, "+92", "3088888888"))
	require.Error(t, insert(5, 3, "+92", "3088888888"))

	h5A, err := h5FindGameUserByPhone(db, content.Scope{SiteID: 3}, "+92", "3088888888")
	require.NoError(t, err)
	require.Equal(t, uint64(2), h5A.ID)

	h5B, err := h5FindGameUserByPhone(db, content.Scope{SiteID: 5}, "+92", "3088888888")
	require.NoError(t, err)
	require.Equal(t, uint64(3), h5B.ID)

	_, err = h5FindGameUserByPhone(db, content.Scope{SiteID: 3}, "+1", "3088888888")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestH5GameUserPhoneDuplicateClassification(t *testing.T) {
	require.True(t, h5GameUserPhoneDuplicate(&mysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry '7-3-+92-3088888888' for key 'game_users.uk_game_users_scope_phone'",
	}))
	require.True(t, h5GameUserPhoneDuplicate(&mysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry '+92-3088888888' for key 'game_users.uk_game_users_phone'",
	}))
	require.False(t, h5GameUserPhoneDuplicate(&mysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry 'I123' for key 'game_users.uk_game_users_invite_code'",
	}))
	require.False(t, h5GameUserPhoneDuplicate(errors.New("connection failed")))
}

func TestH5AuthPhoneCountryCode(t *testing.T) {
	code, err := h5AuthPhoneCountryCode(" +92 ")
	require.NoError(t, err)
	require.Equal(t, "+92", code)

	code, err = h5AuthPhoneCountryCode("")
	require.NoError(t, err)
	require.Equal(t, "+92", code)

	for _, value := range []string{"92", "+", "+9a", "+12345"} {
		_, err = h5AuthPhoneCountryCode(value)
		require.Error(t, err)
	}
}

func TestH5AuthPhonePreservesIndonesianNumbersAndDistinctPrefixes(t *testing.T) {
	for _, value := range []string{"81234567890", "821234567890", "3008567530"} {
		phone, err := h5AuthPhone(value)
		require.NoError(t, err)
		require.Equal(t, value, phone)
	}
	first, err := h5AuthPhone("81234567890")
	require.NoError(t, err)
	second, err := h5AuthPhone("91234567890")
	require.NoError(t, err)
	require.NotEqual(t, first, second)
	for _, value := range []string{"12345", "1234567890123456"} {
		_, err = h5AuthPhone(value)
		require.Error(t, err)
	}
}
