package agentadmin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMoneyUsesExactDecimalAndTwoMinorUnits(t *testing.T) {
	amount, err := money("100.005")
	require.NoError(t, err)
	require.Equal(t, "100.01", amount.StringFixed(2))

	_, err = money("-0.01")
	require.Error(t, err)
}

func TestPageBounds(t *testing.T) {
	pageNumber, pageSize, offset := page(ListQuery{Page: 0, PageSize: 500})
	require.Equal(t, 1, pageNumber)
	require.Equal(t, 100, pageSize)
	require.Equal(t, 0, offset)

	pageNumber, pageSize, offset = page(ListQuery{Page: 3, PageSize: 20})
	require.Equal(t, 3, pageNumber)
	require.Equal(t, 20, pageSize)
	require.Equal(t, 40, offset)
}

func TestSameDateIgnoresClockTime(t *testing.T) {
	left := time.Date(2026, 7, 16, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	right := time.Date(2026, 7, 16, 23, 59, 59, 0, time.FixedZone("UTC+8", 8*60*60))
	require.True(t, sameDate(left, right))
	require.False(t, sameDate(left, right.Add(24*time.Hour)))
}

func TestDatabaseBoolSupportsMySQLRepresentations(t *testing.T) {
	require.True(t, databaseBool(true))
	require.True(t, databaseBool(int64(1)))
	require.True(t, databaseBool([]byte("1")))
	require.False(t, databaseBool(int64(0)))
	require.False(t, databaseBool([]byte("0")))
}
