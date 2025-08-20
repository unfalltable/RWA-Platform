package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/rwa-platform/data-collector/internal/config"
	"github.com/rwa-platform/data-collector/internal/kafka"
	"github.com/rwa-platform/data-collector/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BlockchainService struct {
	db      *gorm.DB
	redis   *redis.Client
	kafka   *kafka.Producer
	config  *config.Config
	clients map[string]*ethclient.Client
	logger  *logrus.Logger
}

type ChainConfig struct {
	Name    string
	RPC     string
	ChainID int64
}

func NewBlockchainService(db *gorm.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, cfg *config.Config) *BlockchainService {
	service := &BlockchainService{
		db:      db,
		redis:   redisClient,
		kafka:   kafkaProducer,
		config:  cfg,
		clients: make(map[string]*ethclient.Client),
		logger:  logrus.New(),
	}

	// 初始化区块链客户端
	service.initClients()
	
	return service
}

func (s *BlockchainService) initClients() {
	chains := []ChainConfig{
		{"ethereum", s.config.EthereumRPC, 1},
		{"arbitrum", s.config.ArbitrumRPC, 42161},
		{"base", s.config.BaseRPC, 8453},
		{"polygon", s.config.PolygonRPC, 137},
		{"bsc", s.config.BSCRPC, 56},
	}

	for _, chain := range chains {
		if chain.RPC != "" {
			client, err := ethclient.Dial(chain.RPC)
			if err != nil {
				s.logger.Errorf("Failed to connect to %s: %v", chain.Name, err)
				continue
			}
			s.clients[chain.Name] = client
			s.logger.Infof("Connected to %s blockchain", chain.Name)
		}
	}
}

func (s *BlockchainService) StartBlockchainIndexing(ctx context.Context) {
	s.logger.Info("Starting blockchain indexing service")
	
	ticker := time.NewTicker(time.Duration(s.config.BlockchainSyncInterval) * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	s.indexBlockchainData(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Blockchain indexing service stopped")
			return
		case <-ticker.C:
			s.indexBlockchainData(ctx)
		}
	}
}

func (s *BlockchainService) indexBlockchainData(ctx context.Context) {
	s.logger.Info("Starting blockchain indexing cycle")

	for chainName, client := range s.clients {
		select {
		case <-ctx.Done():
			return
		default:
			s.indexChain(ctx, chainName, client)
		}
	}

	s.logger.Info("Blockchain indexing cycle completed")
}

func (s *BlockchainService) indexChain(ctx context.Context, chainName string, client *ethclient.Client) {
	// 获取最新区块号
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get latest block for %s: %v", chainName, err)
		return
	}

	// 获取上次同步的区块号
	lastSyncedBlock := s.getLastSyncedBlock(chainName)
	
	// 如果是首次同步，从最近的100个区块开始
	if lastSyncedBlock == 0 {
		if latestBlock > 100 {
			lastSyncedBlock = latestBlock - 100
		}
	}

	// 限制每次同步的区块数量
	maxBlocks := uint64(50)
	endBlock := lastSyncedBlock + maxBlocks
	if endBlock > latestBlock {
		endBlock = latestBlock
	}

	s.logger.Infof("Indexing %s blocks from %d to %d", chainName, lastSyncedBlock+1, endBlock)

	// 逐个处理区块
	for blockNum := lastSyncedBlock + 1; blockNum <= endBlock; blockNum++ {
		select {
		case <-ctx.Done():
			return
		default:
			if err := s.processBlock(ctx, chainName, client, blockNum); err != nil {
				s.logger.Errorf("Failed to process block %d on %s: %v", blockNum, chainName, err)
				continue
			}
		}
	}

	// 更新最后同步的区块号
	s.setLastSyncedBlock(chainName, endBlock)
}

func (s *BlockchainService) processBlock(ctx context.Context, chainName string, client *ethclient.Client, blockNum uint64) error {
	// 获取区块信息
	block, err := client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
	if err != nil {
		return fmt.Errorf("failed to get block %d: %v", blockNum, err)
	}

	// 处理区块中的交易
	for _, tx := range block.Transactions() {
		if err := s.processTransaction(ctx, chainName, client, tx, block); err != nil {
			s.logger.Errorf("Failed to process transaction %s: %v", tx.Hash().Hex(), err)
			continue
		}
	}

	return nil
}

func (s *BlockchainService) processTransaction(ctx context.Context, chainName string, client *ethclient.Client, tx *types.Transaction, block *types.Block) error {
	// 获取交易收据
	receipt, err := client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	// 创建交易记录
	transaction := &models.BlockchainTransaction{
		Chain:            chainName,
		Hash:             tx.Hash().Hex(),
		BlockNumber:      block.NumberU64(),
		BlockHash:        block.Hash().Hex(),
		TransactionIndex: receipt.TransactionIndex,
		FromAddress:      s.getFromAddress(tx),
		Value:            tx.Value().String(),
		GasUsed:          &receipt.GasUsed,
		Status:           &receipt.Status,
		Timestamp:        time.Unix(int64(block.Time()), 0),
	}

	if tx.To() != nil {
		toAddr := tx.To().Hex()
		transaction.ToAddress = &toAddr
	}

	if receipt.ContractAddress != (common.Address{}) {
		contractAddr := receipt.ContractAddress.Hex()
		transaction.ContractAddress = &contractAddr
	}

	gasPrice := tx.GasPrice()
	if gasPrice != nil {
		gasPriceStr := gasPrice.String()
		transaction.GasPrice = &gasPriceStr
	}

	// 序列化日志
	if len(receipt.Logs) > 0 {
		logsData, err := json.Marshal(receipt.Logs)
		if err == nil {
			transaction.Logs = logsData
		}
	}

	// 保存交易到数据库
	if err := s.db.Create(transaction).Error; err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("failed to save transaction: %v", err)
		}
	}

	// 处理代币转账事件
	s.processTokenTransfers(chainName, tx, receipt, block)

	// 发布到Kafka
	s.publishTransactionEvent(transaction)

	return nil
}

func (s *BlockchainService) processTokenTransfers(chainName string, tx *types.Transaction, receipt *types.Receipt, block *types.Block) {
	// ERC-20 Transfer事件的签名
	transferEventSignature := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	for _, log := range receipt.Logs {
		if len(log.Topics) >= 3 && log.Topics[0] == transferEventSignature {
			// 解析Transfer事件
			transfer := &models.TokenTransfer{
				Chain:           chainName,
				TransactionHash: tx.Hash().Hex(),
				LogIndex:        log.Index,
				ContractAddress: log.Address.Hex(),
				FromAddress:     common.HexToAddress(log.Topics[1].Hex()).Hex(),
				ToAddress:       common.HexToAddress(log.Topics[2].Hex()).Hex(),
				Value:           new(big.Int).SetBytes(log.Data).String(),
				BlockNumber:     block.NumberU64(),
				Timestamp:       time.Unix(int64(block.Time()), 0),
			}

			// 保存代币转账记录
			if err := s.db.Create(transfer).Error; err != nil {
				if !strings.Contains(err.Error(), "duplicate key") {
					s.logger.Errorf("Failed to save token transfer: %v", err)
				}
			}

			// 发布代币转账事件
			s.publishTokenTransferEvent(transfer)
		}
	}
}

func (s *BlockchainService) getFromAddress(tx *types.Transaction) string {
	// 从交易中恢复发送者地址
	signer := types.LatestSignerForChainID(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return ""
	}
	return from.Hex()
}

func (s *BlockchainService) getLastSyncedBlock(chainName string) uint64 {
	key := fmt.Sprintf("last_synced_block:%s", chainName)
	result, err := s.redis.Get(context.Background(), key).Result()
	if err != nil {
		return 0
	}

	var blockNum uint64
	fmt.Sscanf(result, "%d", &blockNum)
	return blockNum
}

func (s *BlockchainService) setLastSyncedBlock(chainName string, blockNum uint64) {
	key := fmt.Sprintf("last_synced_block:%s", chainName)
	s.redis.Set(context.Background(), key, fmt.Sprintf("%d", blockNum), 0)
}

func (s *BlockchainService) publishTransactionEvent(transaction *models.BlockchainTransaction) {
	message := map[string]interface{}{
		"type":         "blockchain_transaction",
		"chain":        transaction.Chain,
		"hash":         transaction.Hash,
		"block_number": transaction.BlockNumber,
		"from_address": transaction.FromAddress,
		"to_address":   transaction.ToAddress,
		"value":        transaction.Value,
		"timestamp":    transaction.Timestamp.Unix(),
	}

	if err := s.kafka.PublishMessage("blockchain-events", transaction.Hash, message); err != nil {
		s.logger.Errorf("Failed to publish transaction event: %v", err)
	}
}

func (s *BlockchainService) publishTokenTransferEvent(transfer *models.TokenTransfer) {
	message := map[string]interface{}{
		"type":             "token_transfer",
		"chain":            transfer.Chain,
		"transaction_hash": transfer.TransactionHash,
		"contract_address": transfer.ContractAddress,
		"from_address":     transfer.FromAddress,
		"to_address":       transfer.ToAddress,
		"value":            transfer.Value,
		"block_number":     transfer.BlockNumber,
		"timestamp":        transfer.Timestamp.Unix(),
	}

	if err := s.kafka.PublishMessage("token-transfers", transfer.TransactionHash, message); err != nil {
		s.logger.Errorf("Failed to publish token transfer event: %v", err)
	}
}

func (s *BlockchainService) GetAssetInfo(contractAddress string) (*models.Asset, error) {
	var asset models.Asset
	if err := s.db.Where("contracts @> ?", fmt.Sprintf(`[{"address": "%s"}]`, contractAddress)).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (s *BlockchainService) GetTransaction(hash string) (*models.BlockchainTransaction, error) {
	var transaction models.BlockchainTransaction
	if err := s.db.Where("hash = ?", hash).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (s *BlockchainService) TriggerSync() error {
	go s.indexBlockchainData(context.Background())
	return nil
}
