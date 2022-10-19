package eth

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/bor"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
)

// GetRootHash returns root hash for given start and end block
func (b *EthAPIBackend) GetRootHash(ctx context.Context, starBlockNr uint64, endBlockNr uint64) (string, error) {
	var api *bor.API
	for _, _api := range b.eth.Engine().APIs(b.eth.BlockChain()) {
		if _api.Namespace == "bor" {
			api = _api.Service.(*bor.API)
		}
	}

	if api == nil {
		return "", errors.New("Only available in Bor engine")
	}

	root, err := api.GetRootHash(starBlockNr, endBlockNr)
	if err != nil {
		return "", err
	}

	return root, nil
}

// GetRootHash returns root hash for given start and end block
func (b *EthAPIBackend) GetVoteOnRootHash(ctx context.Context, starBlockNr uint64, endBlockNr uint64, rootHash string, milestoneId string) (bool, error) {
	var api *bor.API
	for _, _api := range b.eth.Engine().APIs(b.eth.BlockChain()) {
		if _api.Namespace == "bor" {
			api = _api.Service.(*bor.API)
		}
	}

	if api == nil {
		return false, errors.New("Only available in Bor engine")
	}

	downloader := b.eth.handler.downloader

	log.Warn("In the GetRootHash-Locking the mutex")
	isLocked := downloader.Lock(endBlockNr)

	if !isLocked {
		log.Warn("In the GetRootHash-Unlocking the mutex")
		downloader.Unlock(false, "", common.Hash{})

		return false, errors.New("Previous sprint is still in locked state")
	}

	root, err := api.GetRootHash(starBlockNr, endBlockNr)

	if err != nil {
		log.Warn("In the GetRootHash-Unlocking the mutex")
		downloader.Unlock(false, "", common.Hash{})

		return false, err
	}

	if root != rootHash {
		log.Warn("In the GetRootHash-Unlocking the mutex")
		downloader.Unlock(false, "", common.Hash{})

		return false, errors.New("RootHash mismatch+" + root + rootHash)
	}

	endBlock := b.eth.blockchain.GetBlockByNumber(endBlockNr)
	endBlockHash := endBlock.Hash()
	log.Warn("In the GetRootHash-Unlocking the mutex")
	downloader.Unlock(true, milestoneId, endBlockHash)

	return true, nil
}

// GetBorBlockReceipt returns bor block receipt
func (b *EthAPIBackend) GetBorBlockReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	receipt := b.eth.blockchain.GetBorReceiptByHash(hash)
	if receipt == nil {
		return nil, ethereum.NotFound
	}

	return receipt, nil
}

// GetBorBlockLogs returns bor block logs
func (b *EthAPIBackend) GetBorBlockLogs(ctx context.Context, hash common.Hash) ([]*types.Log, error) {
	receipt := b.eth.blockchain.GetBorReceiptByHash(hash)
	if receipt == nil {
		return nil, nil
	}
	return receipt.Logs, nil
}

// GetBorBlockTransaction returns bor block tx
func (b *EthAPIBackend) GetBorBlockTransaction(ctx context.Context, hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadBorTransaction(b.eth.ChainDb(), hash)
	return tx, blockHash, blockNumber, index, nil
}

func (b *EthAPIBackend) GetBorBlockTransactionWithBlockHash(ctx context.Context, txHash common.Hash, blockHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadBorTransactionWithBlockHash(b.eth.ChainDb(), txHash, blockHash)
	return tx, blockHash, blockNumber, index, nil
}

// SubscribeStateSyncEvent subscribes to state sync event
func (b *EthAPIBackend) SubscribeStateSyncEvent(ch chan<- core.StateSyncEvent) event.Subscription {
	return b.eth.BlockChain().SubscribeStateSyncEvent(ch)
}

// SubscribeChain2HeadEvent subscribes to reorg/head/fork event
func (b *EthAPIBackend) SubscribeChain2HeadEvent(ch chan<- core.Chain2HeadEvent) event.Subscription {
	return b.eth.BlockChain().SubscribeChain2HeadEvent(ch)
}
