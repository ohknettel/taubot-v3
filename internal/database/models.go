package database

import (
	"time"
	"github.com/ohknettel/taubot-v3/pkg/datatypes"
)

const (
	P_OpenAccount uint8 = iota
	P_ViewBalance 
	P_CloseAccount 
	P_TransferFunds 
	P_CreateRecurringTransfer 
	P_ManageFunds 
	P_ManageTaxBrackets 
	P_ManagePermissions 
	P_ManageEconomies 
	P_OpenSpecialAccount 
	P_LoginAsAccount 
	P_InstallPlugins 
)

const (
	AT_User uint8 = iota
	AT_Government
	AT_Corporation
	AT_Charity
)

const (
	TX_Wealth uint8 = iota
	TX_Income
	TX_VAT
	TX_Transaction
)

const (
	TT_Personal uint8 = iota
	TT_Income
	TT_Purchase
)

const (
	CUD_Create uint8 = iota
	CUD_Update
	CUD_Delete
)
	
var Models []any = []any{
	Economy{},
	Guild{},
	MinecraftIntegration{},
	Plugin{},
	Account{},
	PluginLink{},
	Transfer{},
	UserPermission{},
	Tax{},
	RecurringTransfer{},
}

type Economy struct {
	Name 			string
	ID 				datatypes.UUID 	`gorm:"primaryKey"`
	ParentGuildID 	string

	CurrencyName 	string 			`gorm:"unique"`
	CurrencyUnit 	string

	Guilds 			[]Guild
	Accounts 		[]Account
	Plugins 		[]Plugin
}

type Guild struct {
	ID			uint 			`gorm:"primaryKey"`
	GuildID 	string	
	EconomyID 	datatypes.UUID 	`gorm:"index"`
	Economy 	Economy
}

type MinecraftIntegration struct {
	UserID string `gorm:"primaryKey"`
	MinecraftToken string
}

type Plugin struct {
	ID 			datatypes.UUID 	`gorm:"primaryKey"`
	PluginName 	string
	PluginLogo 	*string

	OwnerID 	string
	EconomyID 	string 			`gorm:"index"`
	Economy 	Economy
	Links 		[]PluginLink 	`gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Account struct {
	ID 				datatypes.UUID 	`gorm:"primaryKey"`
	AccountName 	string
	AccountType 	uint8
	AccountLogo 	*string
	OwnerID 		string

	Balance 		uint
	TotalBalance 	uint
	Deleted 		bool 			`gorm:"default:false"`

	EconomyID 		string 			`gorm:"index"`
	Economy 		Economy
}

type PluginLink struct {
	ID			uint 	`gorm:"primaryKey"`
	PluginID 	datatypes.UUID 
	AccountID 	datatypes.UUID
	Enabled 	bool
}

type Transfer struct {
	TrxID 			uint `gorm:"primaryKey;autoIncrement"`
	ActorID 		string
	CreatedAt 		time.Time

	FromAccountID 	datatypes.UUID
	FromAccount		Account

	ToAccountID		datatypes.UUID
	ToAccount		Account
}

type UserPermission struct {
	EntryID 	string 			`gorm:"primaryKey"`
	UserID 		string 			`gorm:"index"`

	AccountID 	*datatypes.UUID	`gorm:"index"`
	Account 	*Account 

	EconomyID 	*datatypes.UUID	`gorm:"index"`
	Economy 	*Economy 

	PermissionID uint8
	Value 		 bool
}

type Tax struct {
	EntryID 		string 	`gorm:"primaryKey"`
	TaxName 		string
	AffectedType 	uint8
	TaxType 		uint8
	BracketStart 	uint
	BracketEnd 		uint
	Rate 			uint

	ToAccountID 	datatypes.UUID
	ToAccount 		Account `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type RecurringTransfer struct {
	EntryID 		string 			`gorm:"primaryKey"`
	ActorID 		string

	FromAccountID 	datatypes.UUID 	`gorm:"index"`
	FromAccount 	Account 		`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ToAccountID 	datatypes.UUID 	`gorm:"index"`
	ToAccount 		Account 		`gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Amount 			uint
	LastPaid 		time.Time
	PaymentInterval uint
	PaymentsLeft 	uint
}

func (Economy) TableName() string {
	return "economies"
}

func (UserPermission) TableName() string {
	return "permissions"
}

func (Tax) TableName() string {
	return "taxes"
}