package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.


// CreateGameWithdrawAccountsRequest request params
type CreateGameWithdrawAccountsRequest struct {
	UserID  uint64 `json:"userID" binding:""`
	AccountType  string `json:"accountType" binding:""` // bank, wallet
	BankName  string `json:"bankName" binding:""`
	WalletType  string `json:"walletType" binding:""`
	AccountName  string `json:"accountName" binding:""`
	AccountNo  string `json:"accountNo" binding:""`
	PhoneCountryCode  string `json:"phoneCountryCode" binding:""`
	Phone  string `json:"phone" binding:""`
	IsDefault  int `json:"isDefault" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 disabled, 3 reviewing
}

// UpdateGameWithdrawAccountsByIDRequest request params
type UpdateGameWithdrawAccountsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID  uint64 `json:"userID" binding:""`
	AccountType  string `json:"accountType" binding:""` // bank, wallet
	BankName  string `json:"bankName" binding:""`
	WalletType  string `json:"walletType" binding:""`
	AccountName  string `json:"accountName" binding:""`
	AccountNo  string `json:"accountNo" binding:""`
	PhoneCountryCode  string `json:"phoneCountryCode" binding:""`
	Phone  string `json:"phone" binding:""`
	IsDefault  int `json:"isDefault" binding:""`
	Status  int `json:"status" binding:""` // 1 active, 2 disabled, 3 reviewing
}

// GameWithdrawAccountsObjDetail detail
type GameWithdrawAccountsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID  uint64 `json:"userID"`
	AccountType  string `json:"accountType"` // bank, wallet
	BankName  string `json:"bankName"`
	WalletType  string `json:"walletType"`
	AccountName  string `json:"accountName"`
	AccountNo  string `json:"accountNo"`
	PhoneCountryCode  string `json:"phoneCountryCode"`
	Phone  string `json:"phone"`
	IsDefault  int `json:"isDefault"`
	Status  int `json:"status"` // 1 active, 2 disabled, 3 reviewing
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}


// CreateGameWithdrawAccountsReply only for api docs
type CreateGameWithdrawAccountsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameWithdrawAccountsByIDReply only for api docs
type DeleteGameWithdrawAccountsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameWithdrawAccountsByIDReply only for api docs
type UpdateGameWithdrawAccountsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameWithdrawAccountsByIDReply only for api docs
type GetGameWithdrawAccountsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWithdrawAccounts GameWithdrawAccountsObjDetail `json:"gameWithdrawAccounts"`
	} `json:"data"` // return data
}

// ListGameWithdrawAccountssRequest request params
type ListGameWithdrawAccountssRequest struct {
	query.Params
}

// ListGameWithdrawAccountssReply only for api docs
type ListGameWithdrawAccountssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameWithdrawAccountss []GameWithdrawAccountsObjDetail `json:"gameWithdrawAccountss"`
	} `json:"data"` // return data
}
