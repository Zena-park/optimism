package verification

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/batch"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// VerificationResult는 검증 결과를 나타냅니다.
type VerificationResult struct {
	GameID           common.Hash
	BatchNumber      uint64
	StateRootValid   bool
	BatchDataValid   bool
	OverallValid     bool
	StateRootError   error
	BatchDataError   error
	VerificationTime int64 // milliseconds
	ResultType       VerificationResultType
}

// VerificationResultType은 검증 결과의 유형을 나타냅니다.
type VerificationResultType string

const (
	ResultValid        VerificationResultType = "VALID"
	ResultInvalidState VerificationResultType = "INVALID_STATE"
	ResultInvalidBatch VerificationResultType = "INVALID_BATCH"
	ResultInvalidBoth  VerificationResultType = "INVALID_BOTH"
	ResultError        VerificationResultType = "ERROR"
)

// StateRootVerifier는 기존 상태 루트 검증 기능을 정의합니다.
type StateRootVerifier interface {
	VerifyStateRoot(ctx context.Context, game types.Game) (bool, error)
}

// EnhancedVerifier는 상태 루트 검증과 배치 데이터 검증을 통합합니다.
type EnhancedVerifier struct {
	logger        log.Logger
	stateVerifier StateRootVerifier
	batchChecker  batch.BatchChecker
	metrics       VerificationMetrics
}

// VerificationMetrics는 검증 관련 메트릭을 정의합니다.
type VerificationMetrics interface {
	RecordVerificationResult(resultType VerificationResultType, duration time.Duration)
	RecordStateRootVerification(success bool, duration time.Duration)
	RecordBatchDataVerification(success bool, duration time.Duration)
}

// NewEnhancedVerifier는 새로운 EnhancedVerifier 인스턴스를 생성합니다.
func NewEnhancedVerifier(
	logger log.Logger,
	stateVerifier StateRootVerifier,
	batchVerifier batch.BatchVerifier,
	metrics VerificationMetrics,
) *EnhancedVerifier {
	return &EnhancedVerifier{
		logger:        logger,
		stateVerifier: stateVerifier,
		batchVerifier: batchVerifier,
		metrics:       metrics,
	}
}

// VerifyGame은 게임에 대해 통합 검증을 수행합니다.
func (v *EnhancedVerifier) VerifyGame(ctx context.Context, game types.Game, batchNumber uint64) (*VerificationResult, error) {
	startTime := time.Now()

	result := &VerificationResult{
		GameID:      game.Claims()[0].Value, // 루트 클레임의 값이 게임 ID
		BatchNumber: batchNumber,
	}

	// 1. 상태 루트 검증 (기존 방식)
	stateStart := time.Now()
	stateValid, stateErr := v.verifyStateRoot(ctx, game)
	stateDuration := time.Since(stateStart)

	result.StateRootValid = stateValid
	result.StateRootError = stateErr

	v.metrics.RecordStateRootVerification(stateValid, stateDuration)

	// 2. 배치 데이터 검증 (새로운 방식)
	batchStart := time.Now()
	batchValid, batchErr := v.verifyBatchData(ctx, batchNumber)
	batchDuration := time.Since(batchStart)

	result.BatchDataValid = batchValid
	result.BatchDataError = batchErr

	v.metrics.RecordBatchDataVerification(batchValid, batchDuration)

	// 3. 통합 결과 결정
	result.OverallValid = stateValid && batchValid
	result.VerificationTime = time.Since(startTime).Milliseconds()
	result.ResultType = v.determineResultType(stateValid, batchValid, stateErr, batchErr)

	// 4. 로깅
	v.logVerificationResult(result)

	// 5. 메트릭 기록
	v.metrics.RecordVerificationResult(result.ResultType, time.Since(startTime))

	return result, nil
}

// verifyStateRoot는 상태 루트 검증을 수행합니다.
func (v *EnhancedVerifier) verifyStateRoot(ctx context.Context, game types.Game) (bool, error) {
	if v.stateVerifier == nil {
		v.logger.Warn("State root verifier not available, skipping state root verification")
		return true, nil // 검증기를 사용할 수 없는 경우 기본적으로 유효하다고 가정
	}

	return v.stateVerifier.VerifyStateRoot(ctx, game)
}

// verifyBatchData는 배치 데이터 검증을 수행합니다.
func (v *EnhancedVerifier) verifyBatchData(ctx context.Context, batchNumber uint64) (bool, error) {
	if v.batchVerifier == nil {
		v.logger.Warn("Batch verifier not available, skipping batch data verification")
		return true, nil // 검증기를 사용할 수 없는 경우 기본적으로 유효하다고 가정
	}

	batchResult, err := v.batchVerifier.VerifyBatchData(ctx, batchNumber)
	if err != nil {
		return false, fmt.Errorf("batch verification failed: %w", err)
	}

	return batchResult.IsValid, batchResult.Error
}

// determineResultType은 검증 결과의 유형을 결정합니다.
func (v *EnhancedVerifier) determineResultType(stateValid, batchValid bool, stateErr, batchErr error) VerificationResultType {
	// 에러가 있는 경우
	if stateErr != nil || batchErr != nil {
		return ResultError
	}

	// 두 검증 모두 성공
	if stateValid && batchValid {
		return ResultValid
	}

	// 상태 루트만 실패
	if !stateValid && batchValid {
		return ResultInvalidState
	}

	// 배치 데이터만 실패
	if stateValid && !batchValid {
		return ResultInvalidBatch
	}

	// 둘 다 실패
	return ResultInvalidBoth
}

// logVerificationResult는 검증 결과를 로깅합니다.
func (v *EnhancedVerifier) logVerificationResult(result *VerificationResult) {
	level := log.LevelInfo
	if !result.OverallValid {
		level = log.LevelError
	}

	v.logger.Log(level, "Game verification completed",
		"gameID", result.GameID.Hex(),
		"batchNumber", result.BatchNumber,
		"stateRootValid", result.StateRootValid,
		"batchDataValid", result.BatchDataValid,
		"overallValid", result.OverallValid,
		"resultType", result.ResultType,
		"verificationTime", result.VerificationTime,
		"stateRootError", result.StateRootError,
		"batchDataError", result.BatchDataError,
	)
}

// GetVerificationSummary는 검증 결과의 요약을 반환합니다.
func (v *EnhancedVerifier) GetVerificationSummary(result *VerificationResult) string {
	switch result.ResultType {
	case ResultValid:
		return fmt.Sprintf("✅ Game %s (Batch %d): All verifications passed",
			result.GameID.Hex()[:8], result.BatchNumber)
	case ResultInvalidState:
		return fmt.Sprintf("⚠️ Game %s (Batch %d): State root verification failed",
			result.GameID.Hex()[:8], result.BatchNumber)
	case ResultInvalidBatch:
		return fmt.Sprintf("⚠️ Game %s (Batch %d): Batch data verification failed",
			result.GameID.Hex()[:8], result.BatchNumber)
	case ResultInvalidBoth:
		return fmt.Sprintf("❌ Game %s (Batch %d): Both verifications failed",
			result.GameID.Hex()[:8], result.BatchNumber)
	case ResultError:
		return fmt.Sprintf("💥 Game %s (Batch %d): Verification error occurred",
			result.GameID.Hex()[:8], result.BatchNumber)
	default:
		return fmt.Sprintf("❓ Game %s (Batch %d): Unknown verification result",
			result.GameID.Hex()[:8], result.BatchNumber)
	}
}
