package batch

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// BatchChecker 인터페이스는 L1과 L2 배치 데이터를 비교하는 기능을 정의합니다.
type BatchChecker interface {
	// VerifyBatchData는 특정 배치 번호에 대해 L1과 L2 배치 데이터를 비교합니다.
	VerifyBatchData(ctx context.Context, batchNumber uint64) (*BatchComparisonResult, error)

	// GetL1BatchData는 L1 BatchInbox 컨트랙트에서 배치 데이터를 추출합니다.
	GetL1BatchData(ctx context.Context, batchNumber uint64) (*BatchData, error)

	// GetL2BatchData는 L2 체인에서 배치 데이터를 추출합니다.
	GetL2BatchData(ctx context.Context, batchNumber uint64) (*BatchData, error)

	// CompareBatchData는 두 배치 데이터를 비교합니다.
	CompareBatchData(l1Data, l2Data *BatchData) *BatchComparisonResult
}

// BatchData는 배치의 기본 정보를 담습니다.
type BatchData struct {
	BatchNumber uint64
	Hash        common.Hash
	RawData     []byte
	Timestamp   uint64
	BlockNumber uint64
}

// BatchComparisonResult는 배치 비교 결과를 담습니다.
type BatchComparisonResult struct {
	BatchNumber      uint64
	IsValid          bool
	L1Hash           common.Hash
	L2Hash           common.Hash
	HashMatch        bool
	DataMatch        bool
	Differences      []string
	Error            error
	VerificationTime int64 // milliseconds
}

// DefaultBatchChecker는 BatchChecker의 기본 구현체입니다.
type DefaultBatchChecker struct {
	logger     log.Logger
	l1Client   *ethclient.Client
	l2Client   *ethclient.Client
	batchInbox common.Address
}

// NewDefaultBatchChecker는 새로운 DefaultBatchChecker 인스턴스를 생성합니다.
func NewDefaultBatchChecker(
	logger log.Logger,
	l1Client *ethclient.Client,
	l2Client *ethclient.Client,
	batchInbox common.Address,
) *DefaultBatchChecker {
	return &DefaultBatchChecker{
		logger:     logger,
		l1Client:   l1Client,
		l2Client:   l2Client,
		batchInbox: batchInbox,
	}
}

// VerifyBatchData는 특정 배치 번호에 대해 L1과 L2 배치 데이터를 비교합니다.
func (v *DefaultBatchChecker) VerifyBatchData(ctx context.Context, batchNumber uint64) (*BatchComparisonResult, error) {
	startTime := time.Now()

	// L1 배치 데이터 추출
	l1Data, err := v.GetL1BatchData(ctx, batchNumber)
	if err != nil {
		return &BatchComparisonResult{
			BatchNumber: batchNumber,
			IsValid:     false,
			Error:       fmt.Errorf("failed to get L1 batch data: %w", err),
		}, nil
	}

	// L2 배치 데이터 추출
	l2Data, err := v.GetL2BatchData(ctx, batchNumber)
	if err != nil {
		return &BatchComparisonResult{
			BatchNumber: batchNumber,
			IsValid:     false,
			Error:       fmt.Errorf("failed to get L2 batch data: %w", err),
		}, nil
	}

	// 배치 데이터 비교
	result := v.CompareBatchData(l1Data, l2Data)
	result.VerificationTime = time.Since(startTime).Milliseconds()

	return result, nil
}

// GetL1BatchData는 L1 BatchInbox 컨트랙트에서 배치 데이터를 추출합니다.
func (v *DefaultBatchChecker) GetL1BatchData(ctx context.Context, batchNumber uint64) (*BatchData, error) {
	// BatchInbox 컨트랙트의 getBatch 함수 호출
	// 실제 구현에서는 컨트랙트 ABI를 사용하여 호출
	data, err := v.callBatchInboxGetBatch(ctx, batchNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call getBatch: %w", err)
	}

	hash := sha256.Sum256(data)

	return &BatchData{
		BatchNumber: batchNumber,
		Hash:        common.BytesToHash(hash[:]),
		RawData:     data,
		Timestamp:   uint64(time.Now().Unix()),
		BlockNumber: 0, // L1 블록 번호는 별도로 조회 필요
	}, nil
}

// GetL2BatchData는 L2 체인에서 배치 데이터를 추출합니다.
func (v *DefaultBatchVerifier) GetL2BatchData(ctx context.Context, batchNumber uint64) (*BatchData, error) {
	// L2 블록 번호를 배치 번호로 변환 (1:1 매핑 가정)
	blockNumber := batchNumber

	// L2 블록 정보 조회
	block, err := v.l2Client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get L2 block: %w", err)
	}

	// 블록의 트랜잭션들을 배치 형태로 변환
	batchData := v.convertBlockToBatchData(block)
	batchData.BatchNumber = batchNumber

	return batchData, nil
}

// CompareBatchData는 두 배치 데이터를 비교합니다.
func (v *DefaultBatchVerifier) CompareBatchData(l1Data, l2Data *BatchData) *BatchComparisonResult {
	result := &BatchComparisonResult{
		BatchNumber: l1Data.BatchNumber,
		L1Hash:      l1Data.Hash,
		L2Hash:      l2Data.Hash,
		HashMatch:   l1Data.Hash == l2Data.Hash,
		DataMatch:   false,
		Differences: []string{},
	}

	// 해시 비교
	if !result.HashMatch {
		result.Differences = append(result.Differences,
			fmt.Sprintf("Hash mismatch: L1=%s, L2=%s", l1Data.Hash.Hex(), l2Data.Hash.Hex()))
	}

	// 원본 데이터 비교
	if len(l1Data.RawData) != len(l2Data.RawData) {
		result.Differences = append(result.Differences,
			fmt.Sprintf("Data length mismatch: L1=%d bytes, L2=%d bytes",
				len(l1Data.RawData), len(l2Data.RawData)))
	} else {
		dataMatch := true
		for i, b := range l1Data.RawData {
			if i >= len(l2Data.RawData) || b != l2Data.RawData[i] {
				dataMatch = false
				break
			}
		}
		result.DataMatch = dataMatch

		if !dataMatch {
			result.Differences = append(result.Differences, "Raw data content mismatch")
		}
	}

	// 전체 유효성 판단
	result.IsValid = result.HashMatch && result.DataMatch && len(result.Differences) == 0

	return result
}

// callBatchInboxGetBatch는 BatchInbox 컨트랙트의 getBatch 함수를 호출합니다.
func (v *DefaultBatchVerifier) callBatchInboxGetBatch(ctx context.Context, batchNumber uint64) ([]byte, error) {
	// 실제 구현에서는 컨트랙트 ABI를 사용하여 호출
	// 여기서는 간단한 예시로 구현
	// TODO: 실제 컨트랙트 ABI 연동 필요

	// 임시로 더미 데이터 반환
	dummyData := []byte(fmt.Sprintf("batch_%d_data", batchNumber))
	return dummyData, nil
}

// convertBlockToBatchData는 L2 블록을 배치 데이터로 변환합니다.
func (v *DefaultBatchVerifier) convertBlockToBatchData(block *types.Block) *BatchData {
	// 블록의 트랜잭션들을 직렬화
	var batchData []byte
	for _, tx := range block.Transactions() {
		// 트랜잭션 데이터를 배치에 추가
		txData, _ := tx.MarshalBinary()
		batchData = append(batchData, txData...)
	}

	hash := sha256.Sum256(batchData)

	return &BatchData{
		Hash:        common.BytesToHash(hash[:]),
		RawData:     batchData,
		Timestamp:   block.Time(),
		BlockNumber: block.NumberU64(),
	}
}
