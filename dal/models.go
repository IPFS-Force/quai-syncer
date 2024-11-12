package dal

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Block 区块信息表结构
type Block struct {
	Number            uint64    `gorm:"primaryKey"`
	Hash              string    `gorm:"type:varchar(66);index"`
	ParentHash        string    `gorm:"type:varchar(66)"`
	Difficulty        string    `gorm:"type:varchar(78)"`
	PrimaryCoinbase   string    `gorm:"type:varchar(42)"`
	Location          string    `gorm:"type:varchar(10)"`
	Timestamp         time.Time `gorm:"index"`
	PrimeTerminusHash string    `gorm:"type:varchar(66)"`
	CreatedAt         time.Time

	// 使用 gorm:"-" 禁用关联的自动保存
	Transactions []Transaction `gorm:"-"`
}

// Transaction coinbase交易信息表结构
type Transaction struct {
	gorm.Model
	Hash            string          `gorm:"type:varchar(66);uniqueIndex"`
	BlockNumber     uint64          `gorm:"index"`
	To              string          `gorm:"type:varchar(42);index"`
	Value           decimal.Decimal `gorm:"type:decimal(65,0)"`
	Timestamp       time.Time       `gorm:"index"`
	EtxType         uint64          // 用于标识是否是coinbase交易
	OriginatingHash string          `gorm:"type:varchar(66);"`
	EtxIndex        uint16
	ConfirmedNumber uint64 `gorm:"index"` // 记录coinbase交易被确认的区块高度
}

// // Block 区块信息表结构
// type Block struct {
// 	gorm.Model
// 	Hash              string `gorm:"type:varchar(66)"`
// 	Number            uint64 `gorm:"primaryKey;not null"`
// 	ParentHash        string `gorm:"type:varchar(66)"`
// 	Time              uint64
// 	Difficulty        string    `gorm:"type:varchar(78)"`
// 	PrimaryCoinbase   string    `gorm:"type:varchar(42)"`
// 	Location          string    `gorm:"type:varchar(10)"`
// 	Timestamp         time.Time `gorm:"index"`
// 	PrimeTerminusHash string    `gorm:"type:varchar(66)"`

// 	// 关联
// 	Transactions []Transaction `gorm:"foreignKey:BlockNumber;references:Number"`
// }

// // Transaction coinbase交易信息表结构
// type Transaction struct {
// 	gorm.Model
// 	Hash            string    `gorm:"type:varchar(66);uniqueIndex"`
// 	BlockNumber     uint64    `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
// 	To              string    `gorm:"type:varchar(42);index"`
// 	Value           string    `gorm:"type:varchar(78)"`
// 	Timestamp       time.Time `gorm:"index"`
// 	EtxType         uint64    // 用于标识是否是coinbase交易
// 	OriginatingHash string    `gorm:"type:varchar(66)"`
// 	EtxIndex        uint16
// 	ConfirmedNumber uint64 `gorm:"index"` // 记录coinbase交易被确认的区块高度
// }
