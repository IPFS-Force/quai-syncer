package dal

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(dsn string, logLevel logger.LogLevel) (*Database, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	database := &Database{db: db}

	// 迁移 Block 表
	if err := db.AutoMigrate(&Block{}); err != nil {
		return nil, fmt.Errorf("failed to migrate block table: %v", err)
	}

	// 迁移 Transaction 表
	if err = db.AutoMigrate(&Transaction{}); err != nil {
		return nil, fmt.Errorf("failed to migrate transaction table: %v", err)
	}

	return database, nil
}

func (d *Database) GetLatestSyncedBlock() (uint64, error) {
	var block Block
	result := d.db.Select("number").Order("number desc").First(&block)
	if result.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if result.Error != nil {
		return 0, fmt.Errorf("failed to get latest synced block: %v", result.Error)
	}
	return block.Number, nil
}

// SaveBlock 保存区块和交易数据
func (d *Database) SaveBlock(block *Block) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		// 保存区块时忽略 Transactions 关联
		if err := tx.Create(block).Error; err != nil {
			return fmt.Errorf("failed to save block: %v", err)
		}

		if len(block.Transactions) == 0 {
			return nil
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hash"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"confirmed_number": gorm.Expr("EXCLUDED.block_number"),
				"value":            gorm.Expr("EXCLUDED.value"),
				"timestamp":        gorm.Expr("EXCLUDED.timestamp"),
			}),
		}).CreateInBatches(block.Transactions, 100).Error; err != nil {
			return fmt.Errorf("failed to save transactions: %v", err)
		}

		return nil
	})
}
