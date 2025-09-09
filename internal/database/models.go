package database

import (
	"github.com/google/uuid"
	"time"
	"gorm.io/gorm"
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
	ID 				string 		`gorm:"primaryKey"`
	ParentGuildID 	string

	CurrencyName 	string 		`gorm:"unique"`
	CurrencyUnit 	string

	Guilds 			[]Guild 	`gorm:"foreignKey:EconomyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Accounts 		[]Account 	`gorm:"foreignKey:EconomyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Plugins 		[]Plugin 	`gorm:"foreignKey:EconomyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Guild struct {
	ID			uint 	`gorm:"primaryKey"`
	GuildID 	string	
	EconomyID 	string 	`gorm:"index"`
	Economy 	Economy
}

type MinecraftIntegration struct {
	UserID string `gorm:"primaryKey"`
	MinecraftToken string
}

type Plugin struct {
	ID 			string 			`gorm:"primaryKey"`
	PluginName 	string
	PluginLogo 	*string

	OwnerID 	string
	EconomyID 	string 			`gorm:"index"`
	Economy 	Economy
	Links 		[]PluginLink 	`gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Account struct {
	ID 				string 	`gorm:"primaryKey"`
	AccountName 	string
	AccountType 	uint8
	AccountLogo 	*string
	OwnerID 		string

	Balance 		uint
	TotalBalance 	uint
	Deleted 		bool 	`gorm:"default:false"`

	EconomyID 		string 	`gorm:"index"`
	Economy 		Economy
}

type PluginLink struct {
	ID			uint 	`gorm:"primaryKey"`
	PluginID 	string 
	AccountID 	string
	Enabled 	bool
}

type Transfer struct {
	TrxID 			uint `gorm:"primaryKey;autoIncrement"`
	ActorID 		string
	CreatedAt 		time.Time

	FromAccountID 	uint8
	FromAccount		Account

	ToAccountID		uint8
	ToAccount		Account
}

type UserPermission struct {
	EntryID 	string 	`gorm:"primaryKey"`
	UserID 		string 	`gorm:"index"`

	AccountID 	*string	`gorm:"index"`
	Account 	*Account 

	EconomyID 	*string	`gorm:"index"`
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

	ToAccountID 	string
	ToAccount 		Account `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type RecurringTransfer struct {
	EntryID 		string 	`gorm:"primaryKey"`
	ActorID 		string

	FromAccountID 	string 	`gorm:"index"`
	FromAccount 	Account `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ToAccountID 	string 	`gorm:"index"`
	ToAccount 		Account `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

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

func (e *Economy) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New().String()
	return
}

func (a *Account) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return
}

func (p *Plugin) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New().String()
	return
}