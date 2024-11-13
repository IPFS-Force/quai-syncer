package dal

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Block 区块
type Block struct {
	Number            uint64 `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt         time.Time
	Timestamp         time.Time `gorm:"index"`
	Hash              string    `gorm:"type:varchar(66);index"`
	ParentHash        string    `gorm:"type:varchar(66)"`
	Difficulty        string    `gorm:"type:varchar(78)"`
	PrimaryCoinbase   string    `gorm:"type:varchar(42)"`
	Location          string    `gorm:"type:varchar(10)"`
	PrimeTerminusHash string    `gorm:"type:varchar(66)"`

	// 使用 gorm:"-" 禁用关联的自动保存
	Transactions []Transaction `gorm:"-"`
}

// Transaction coinbase交易
type Transaction struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement:true"`
	CreatedAt       time.Time       `gorm:"index"`
	BlockNumber     uint64          `gorm:"index"`
	Timestamp       time.Time       `gorm:"index"`
	ConfirmedNumber uint64          `gorm:"index"` // 记录coinbase交易被确认的区块高度
	Hash            string          `gorm:"type:varchar(66);uniqueIndex"`
	To              string          `gorm:"type:varchar(42);index"`
	Value           decimal.Decimal `gorm:"type:decimal(65,0)"`
	AddressType     uint8           // 用于标识是Quai交易还是Qi交易
	OriginatingHash string          `gorm:"type:varchar(66)"`
	EtxIndex        uint16
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	var existingTx Transaction
	if err := tx.Where("hash = ?", t.Hash).First(&existingTx).Error; err == nil {
		t.ID = existingTx.ID
	}
	return nil
}
