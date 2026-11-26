package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateGameUserWalletsRequest request params
type CreateGameUserWalletsRequest struct {
	UserID              uint64 `json:"userID" binding:""`
	Currency            string `json:"currency" binding:""`
	CashBalance         string `json:"cashBalance" binding:""`
	BonusBalance        string `json:"bonusBalance" binding:""`
	WithdrawableBalance string `json:"withdrawableBalance" binding:""`
	FrozenBalance       string `json:"frozenBalance" binding:""`
}

// UpdateGameUserWalletsByIDRequest request params
type UpdateGameUserWalletsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	UserID              uint64 `json:"userID" binding:""`
	Currency            string `json:"currency" binding:""`
	CashBalance         string `json:"cashBalance" binding:""`
	BonusBalance        string `json:"bonusBalance" binding:""`
	WithdrawableBalance string `json:"withdrawableBalance" binding:""`
	FrozenBalance       string `json:"frozenBalance" binding:""`
}

// GameUserWalletsObjDetail detail
type GameUserWalletsObjDetail struct {
	ID uint64 `json:"id"` // convert to uint64 id

	UserID              uint64     `json:"userID"`
	Currency            string     `json:"currency"`
	CashBalance         string     `json:"cashBalance"`
	BonusBalance        string     `json:"bonusBalance"`
	WithdrawableBalance string     `json:"withdrawableBalance"`
	FrozenBalance       string     `json:"frozenBalance"`
	CreatedAt           *time.Time `json:"createdAt"`
	UpdatedAt           *time.Time `json:"updatedAt"`
}

// CreateGameUserWalletsReply only for api docs
type CreateGameUserWalletsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// DeleteGameUserWalletsByIDReply only for api docs
type DeleteGameUserWalletsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// UpdateGameUserWalletsByIDReply only for api docs
type UpdateGameUserWalletsByIDReply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// GetGameUserWalletsByIDReply only for api docs
type GetGameUserWalletsByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserWallets GameUserWalletsObjDetail `json:"gameUserWallets"`
	} `json:"data"` // return data
}

// ListGameUserWalletssRequest request params
type ListGameUserWalletssRequest struct {
	query.Params
}

// ListGameUserWalletssReply only for api docs
type ListGameUserWalletssReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		GameUserWalletss []GameUserWalletsObjDetail `json:"gameUserWalletss"`
	} `json:"data"` // return data
}
