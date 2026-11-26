package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"casino/internal/game/domain"
)

func TestDecimalMinorConversion(t *testing.T) {
	value := decimal.NewFromFloat(100.50)
	require.Equal(t, int64(10050), decimalToMinor(&value))
	require.True(t, decimal.NewFromFloat(100.50).Equal(minorToDecimal(10050)))
}

func TestWalletCommandHashStableForSameTransaction(t *testing.T) {
	cmd := domain.WalletCommand{
		ProviderCode: "softaggregator", Operation: domain.WalletOperationDebit,
		ProviderPlayerID: "sa_1", ProviderTransactionID: "call-1", AmountMinor: 100, Currency: "MXN",
	}
	cmd2 := cmd
	require.Equal(t, commandHash(cmd), commandHash(cmd2))
}

func TestWalletCommandHashDetectsConflictingAmount(t *testing.T) {
	cmd := domain.WalletCommand{
		ProviderCode: "softaggregator", Operation: domain.WalletOperationDebit,
		ProviderPlayerID: "sa_1", ProviderTransactionID: "call-1", AmountMinor: 100, Currency: "MXN",
	}
	cmd2 := cmd
	cmd2.AmountMinor = 101
	require.NotEqual(t, commandHash(cmd), commandHash(cmd2))
}

func TestValidateWalletCommandRejectsNegativeAmount(t *testing.T) {
	err := validateWalletCommand(domain.WalletCommand{
		ProviderTransactionID: "call-1", Currency: "MXN", AmountMinor: -1,
	})
	require.Error(t, err)
}

func TestRollbackDomainRequiresRelatedTransaction(t *testing.T) {
	err := validateWalletCommand(domain.WalletCommand{
		Operation: domain.WalletOperationRollback, ProviderTransactionID: "rollback-1", Currency: "MXN",
	})
	require.Error(t, err)
}
