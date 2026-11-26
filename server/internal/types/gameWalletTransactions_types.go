package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameWalletTransactionsRequest request params
type CreateGameWalletTransactionsRequest struct {
	TxnNo         string `json:"txnNo" binding:""`
	UserID        uint64 `json:"userID" binding:""`
	WalletType    string `json:"walletType" binding:""` // cash, bonus, withdrawable, frozen
	Direction     string `json:"direction" binding:""`  // credit, debit
	Category      string `json:"category" binding:""`   // deposit, withdrawal, bet, payout, cash_reward, commission
	SourceType    string `json:"sourceType" binding:""`
	SourceID      uint64 `json:"sourceID" binding:""`
	OrderNo       string `json:"orderNo" binding:""`
	Amount        string `json:"amount" binding:""`
	BalanceBefore string `json:"balanceBefore" binding:""`
	BalanceAfter  string `json:"balanceAfter" binding:""`
	Remark        string `json:"remark" binding:""`
}

// UpdateGameWalletTransactionsByIDRequest request params
type UpdateGameWalletTransactionsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	TxnNo         string `json:"txnNo" binding:""`
	UserID        uint64 `json:"userID" binding:""`
	WalletType    string `json:"walletType" binding:""` // cash, bonus, withdrawable, frozen
	Direction     string `json:"direction" binding:""`  // credit, debit
	Category      string `json:"category" binding:""`   // deposit, withdrawal, bet, payout, cash_reward, commission
	SourceType    string `json:"sourceType" binding:""`
	SourceID      uint64 `json:"sourceID" binding:""`
	OrderNo       string `json:"orderNo" binding:""`
	Amount        string `json:"amount" binding:""`
	BalanceBefore string `json:"balanceBefore" binding:""`
	BalanceAfter  string `json:"balanceAfter" binding:""`
	Remark        string `json:"remark" binding:""`
}

// GameWalletTransactionsObjDetail detail
type GameWalletTransactionsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	TxnNo         string     `json:"txnNo"`
	UserID        uint64     `json:"userID"`
	WalletType    string     `json:"walletType"` // cash, bonus, withdrawable, frozen
	Direction     string     `json:"direction"`  // credit, debit
	Category      string     `json:"category"`   // deposit, withdrawal, bet, payout, cash_reward, commission
	SourceType    string     `json:"sourceType"`
	SourceID      uint64     `json:"sourceID"`
	OrderNo       string     `json:"orderNo"`
	Amount        string     `json:"amount"`
	BalanceBefore string     `json:"balanceBefore"`
	BalanceAfter  string     `json:"balanceAfter"`
	Remark        string     `json:"remark"`
	CreatedAt     *time.Time `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

// CreateGameWalletTransactionsReply only for api docs
type CreateGameWalletTransactionsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameWalletTransactionsByIDReply only for api docs
type DeleteGameWalletTransactionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameWalletTransactionsByIDReply only for api docs
type UpdateGameWalletTransactionsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameWalletTransactionsByIDReply only for api docs
type GetGameWalletTransactionsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWalletTransactions GameWalletTransactionsObjDetail `json:"gameWalletTransactions"`
	} `json:"data"` // return data
}

// ListGameWalletTransactionssRequest request params
type ListGameWalletTransactionssRequest struct {
	query.Params
}

// ListGameWalletTransactionssReply only for api docs
type ListGameWalletTransactionssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWalletTransactionss []GameWalletTransactionsObjDetail `json:"gameWalletTransactionss"`
	} `json:"data"` // return data
}
