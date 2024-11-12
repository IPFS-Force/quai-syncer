package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"quai-sync/dal"
	"time"

	"github.com/dominant-strategies/go-quai/common"
	"github.com/dominant-strategies/go-quai/core/types"
	"github.com/dominant-strategies/go-quai/quaiclient/ethclient"
	"github.com/shopspring/decimal"
)

type BlockSync struct {
	client      *ethclient.Client
	db          *dal.Database
	debug       bool
	workerCount int             // 工作协程数量
	taskChan    chan uint64     // 任务通道
	resultChan  chan syncResult // 结果通道
}

// 同步结果
type syncResult struct {
	blockNum  uint64
	blockData *dal.Block
	stats     BlockStats
	err       error
}

type BlockStats struct {
	InternalTxCount  uint
	CoinbaseTxCount  uint
	ConfirmedTxCount uint
}

// NewBlockSync 创建新的同步器实例
func NewBlockSync(nodeURL string, dbURL string, debug bool, workerCount int) (*BlockSync, error) {
	// 连接Quai节点
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to quai node: %v", err)
	}

	// 连接数据库
	db, err := dal.NewDatabase(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	if workerCount <= 0 {
		workerCount = 5 // 默认5个工作协程
	}

	return &BlockSync{
		client:      client,
		db:          db,
		debug:       debug,
		workerCount: workerCount,
		taskChan:    make(chan uint64, workerCount*2),     // 任务通道缓冲
		resultChan:  make(chan syncResult, workerCount*2), // 结果通道缓冲
	}, nil
}

// SyncBlocks 同步区块数据
func (bs *BlockSync) SyncBlocks(ctx context.Context) error {
	// 获取当前链上最新区块
	latestBlock, err := bs.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %v", err)
	}

	// 获取数据库中最新同步的区块
	lastSynced, err := bs.db.GetLatestSyncedBlock()
	if err != nil {
		return err
	}

	done := make(chan struct{})

	defer func() {
		close(bs.taskChan)   // 关闭任务通道
		close(bs.resultChan) // 关闭结果通道
		<-done               // 等待结果处理协程完成
	}()

	// 启动工作协程池
	for i := 0; i < bs.workerCount; i++ {
		go bs.worker(ctx)
	}

	// 启动结果处理协程
	resultMap := make(map[uint64]bool)
	blockDataMap := make(map[uint64]*dal.Block)
	nextBlockToConfirm := lastSynced + 1

	go func() {
		defer close(done)

		for result := range bs.resultChan {
			// 记录成功同步的区块
			resultMap[result.blockNum] = true
			blockDataMap[result.blockNum] = result.blockData // 存储区块数据

			// 按顺序确认已同步的区块
			for resultMap[nextBlockToConfirm] {
				if blockData := blockDataMap[nextBlockToConfirm]; blockData != nil {
					// 处理已确认的区块数据，保存区块和交易数据
					attempt := 0
					for {
						if err := bs.db.SaveBlock(blockData); err != nil {
							log.Printf("Failed to save block %d (attempt %d): %v", nextBlockToConfirm, attempt+1, err)
							attempt++
							time.Sleep(time.Second * time.Duration(attempt)) // 递增重试间隔
							continue
						}
						break
					}

					log.Printf("🟢 BLOCK SAVED | Block #%d",
						nextBlockToConfirm)
				}
				delete(resultMap, nextBlockToConfirm)
				delete(blockDataMap, nextBlockToConfirm) // 清理已处理的区块数据
				nextBlockToConfirm++
			}
		}
	}()

	// 分发同步任务
	for i := lastSynced + 1; i <= latestBlock; i++ {
		select {
		case <-ctx.Done():
			log.Printf("BlockSync SyncBlocks: Syncing blocks interrupted")
			return ctx.Err()
		case bs.taskChan <- i:
		}
	}

	return nil
}

func (bs *BlockSync) worker(ctx context.Context) {
	baseDelay := 4 * time.Second
	maxDelay := 32 * time.Second

	for blockNum := range bs.taskChan {
		attempt := 0

		for {
			// 如果不是第一次尝试，则等待一段时间
			if attempt > 0 {
				// 计算指数退避延迟时间: baseDelay * 2^attempt，但不超过maxDelay
				// 防止attempt过大导致位移溢出
				var delay time.Duration
				if attempt >= 5 { // 2^5 = 32, 已达到最大延迟
					delay = maxDelay
				} else {
					delay = baseDelay * time.Duration(1<<uint(attempt))
				}

				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return
				}

				log.Printf("Retrying block %d (attempt %d) after %v delay",
					blockNum, attempt+1, delay)
			}

			select {
			case <-ctx.Done():
				return
			default:
				blockData, stats, err := bs.syncBlock(ctx, blockNum)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					log.Printf("Failed to sync block %d (attempt %d): %v",
						blockNum, attempt+1, err)
					attempt++
					continue
				}

				// 成功同步，发送结果并退出循环
				select {
				case bs.resultChan <- syncResult{
					blockNum:  blockNum,
					blockData: blockData,
					stats:     stats,
					err:       nil,
				}:
				case <-ctx.Done():
					log.Printf("Context cancelled while sending result for block %d", blockNum)
					return
				}

				log.Printf("⭐ BLOCK SYNCED | Block #%d | Attempts: %d | Internal Txs: %d | Coinbase Txs: %d | Confirmed Txs: %d",
					blockNum,
					attempt+1,
					stats.InternalTxCount,
					stats.CoinbaseTxCount,
					stats.ConfirmedTxCount)
			}
			break
		}
	}
}

// syncBlock 同步单个区块数据
func (bs *BlockSync) syncBlock(ctx context.Context, blockNum uint64) (*dal.Block, BlockStats, error) {
	var stats BlockStats

	block, err := bs.client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, stats, err
		}
		return nil, stats, fmt.Errorf("failed to get block %d: %v", blockNum, err)
	}

	blockData := &dal.Block{
		Number:            block.NumberU64(common.ZONE_CTX),
		Hash:              block.Hash().Hex(),
		ParentHash:        block.ParentHash(common.ZONE_CTX).Hex(),
		Difficulty:        block.Difficulty().String(),
		PrimaryCoinbase:   block.PrimaryCoinbase().Hex(),
		Location:          locationToString(block.Location()),
		Timestamp:         time.Unix(int64(block.Time()), 0),
		PrimeTerminusHash: block.PrimeTerminusHash().Hex(),
		// CreatedAt 会在保存时自动设置
	}

	txs := make([]dal.Transaction, 0)

	// 处理普通交易中的coinbase交易
	for _, tx := range block.Transactions() {
		// 只处理External类型的coinbase交易
		if tx.Type() != types.ExternalTxType || tx.EtxType() != types.CoinbaseType {
			stats.InternalTxCount++
			continue
		}

		txs = append(txs, dal.Transaction{
			Hash:            tx.Hash().Hex(),
			BlockNumber:     block.NumberU64(common.ZONE_CTX),
			To:              tx.To().Hex(),
			Value:           decimal.NewFromBigInt(tx.Value(), 0),
			Timestamp:       time.Unix(int64(block.Time()), 0),
			EtxType:         tx.EtxType(),
			OriginatingHash: tx.OriginatingTxHash().Hex(),
			EtxIndex:        tx.ETXIndex(),
		})
	}

	// 处理outboundEtxs中的coinbase交易
	for _, tx := range block.OutboundEtxs() {
		// 判断是否是coinbase交易
		if !types.IsCoinBaseTx(tx) {
			continue
		}

		// 打印coinbase交易详情
		bs.printTxDetails(tx, block.NumberU64(common.ZONE_CTX), block.Location())

		txs = append(txs, dal.Transaction{
			Hash:            tx.Hash().Hex(),
			BlockNumber:     block.NumberU64(common.ZONE_CTX),
			To:              tx.To().Hex(),
			Value:           decimal.NewFromBigInt(tx.Value(), 0),
			Timestamp:       time.Unix(int64(block.Time()), 0),
			EtxType:         tx.EtxType(),
			OriginatingHash: tx.OriginatingTxHash().Hex(),
			EtxIndex:        tx.ETXIndex(),
		})
	}

	blockData.Transactions = txs
	stats.CoinbaseTxCount = uint(len(block.OutboundEtxs()))
	stats.ConfirmedTxCount = uint(len(txs)) - stats.CoinbaseTxCount

	return blockData, stats, nil
}

// Start 启动同步器
func (bs *BlockSync) Start(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("BlockSync Ticker: Syncing blocks interrupted")
			return ctx.Err()
		case <-ticker.C:
			log.Printf("BlockSync Ticker: Syncing blocks...")
			if err := bs.SyncBlocks(ctx); err != nil {
				log.Printf("Sync error: %v", err)
			}
		}
	}
}

// 把 Location 转换为字符串
func locationToString(loc common.Location) string {
	return fmt.Sprintf("%d-%d", loc.Region(), loc.Zone())
}

// printTxDetails 打印交易详情
func (bs *BlockSync) printTxDetails(tx *types.Transaction, blockNumber uint64, location common.Location) {
	if !bs.debug {
		return
	}

	txDetails := map[string]interface{}{
		"blockNumber":     blockNumber,
		"hash":            tx.Hash().Hex(),
		"type":            tx.Type(),
		"etxType":         tx.EtxType(),
		"to":              tx.To().Hex(),
		"value":           tx.Value().String(),
		"gas":             tx.Gas(),
		"etxSender":       tx.ETXSender().Hex(),
		"originatingHash": tx.OriginatingTxHash().Hex(),
		"etxIndex":        tx.ETXIndex(),
		"size":            tx.Size(),
	}
	if tx.From(location) != nil {
		txDetails["from"] = tx.From(location).Hex()
	}

	jsonStr, err := json.MarshalIndent(txDetails, "", "  ")
	if err != nil {
		log.Printf("Error marshaling tx details: %v", err)
	} else {
		log.Printf("Coinbase Transaction Details:\n%s", string(jsonStr))
	}
}
