# Optimism Security Model: Batch Submission vs Dispute Game

## Overview

In Optimism, there are two different L1 submissions: **Sequencer's batch submission** and **Proposer's Dispute Game submission**. This document explains why both submissions are necessary and describes the role and security importance of each.

## 1. Two Submission Systems

### 1.1 Sequencer (op-batcher) - Batch Submission

#### Purpose
- Compress and submit L2 transactions to L1
- Ensure data availability

#### Submission Content
```go
// op-batcher submits L2 transactions to L1
func (b *Batcher) loop() {
    // Collect L2 transactions and compress into batches
    // Submit batch data to BatchInbox
}
```

#### Characteristics
- **Data Transfer**: Store actual transaction data on L1
- **Compression**: Compress multiple transactions into a single batch to save gas costs
- **Availability**: Ensure L2 data is accessible from L1

### 1.2 Proposer (op-proposer) - Dispute Game Submission

#### Purpose
- Create Fault Proof games for L2 state integrity verification
- Ensure state integrity

#### Submission Content
```go
// op-proposer submits L2 state root to L1
func (l *L2OutputSubmitter) sendTransaction(ctx context.Context, output source.Proposal) error {
    if l.Cfg.DisputeGameFactoryAddr != nil {
        candidate, err := l.ProposeL2OutputDGFTxCandidate(ctx, output)
        // Submit output root to DisputeGameFactory
    }
}
```

#### Characteristics
- **State Verification**: Submit L2 calculated state root to L1
- **Challengeable**: Allow challenges against incorrect states
- **Integrity**: Ensure correctness of L2 state

## 2. Why Are Both Submissions Necessary?

### 2.1 Why Batch Submission Alone Is Insufficient

#### Limitations of Batch Submission
```go
// Problems with batch submission only
type BatchOnlyLimitation struct {
    DataAvailability    bool   // ✅ Transaction data is on L1
    StateIntegrity      bool   // ❌ Cannot verify if state is correct
    CalculationError    bool   // ❌ Cannot detect calculation errors
    MaliciousBehavior   bool   // ❌ Cannot detect malicious behavior
}
```

#### Real Scenario: Batch Only
```
1. Sequencer submits transaction batch to L1 ✅
2. L2 performs incorrect calculation, changing state ❌
3. Batch data is correct but state is wrong
4. No one can detect this ❌
5. User asset loss occurs ❌
```

### 2.2 Necessity of Dispute Game

#### Role of Dispute Game
```go
// Security with Dispute Game
type DisputeGameSecurity struct {
    DataAvailability    bool   // ✅ Guaranteed by batch submission
    StateIntegrity      bool   // ✅ Guaranteed by Dispute Game
    CalculationError    bool   // ✅ Detectable through challenges
    MaliciousBehavior   bool   // ✅ Detectable through challenges
}
```

#### Real Scenario: With Dispute Game
```
1. Sequencer submits transaction batch to L1 ✅
2. L2 performs incorrect calculation, changing state ❌
3. Proposer submits incorrect state root to L1
4. op-challenger immediately detects and challenges ✅
5. Incorrect state is rolled back ✅
6. User assets are protected ✅
```

## 3. Cost vs Security Trade-off

### 3.1 Cost Analysis

#### Batch Submission Cost
```go
type BatchSubmissionCost struct {
    GasCost     uint64 // Gas cost for transaction data transmission
    Frequency   string // "Every batch"
    Purpose     string // "Data availability"
}
```

#### Dispute Game Submission Cost
```go
type DisputeGameCost struct {
    GasCost     uint64 // Gas cost for state root verification
    BondCost    uint64 // Bond cost
    Frequency   string // "At configured intervals"
    Purpose     string // "State integrity verification"
}
```

#### Total Cost
```go
type TotalCost struct {
    BatchSubmission    uint64 // Batch submission cost
    DisputeGame        uint64 // Dispute Game submission cost
    Total              uint64 // Sum of both costs
    SecurityLevel      string // "Complete security guarantee"
}
```

### 3.2 Security Level Comparison

#### With Batch Only
```go
type BatchOnlySecurity struct {
    DataAvailability    bool   // ✅ Guaranteed
    StateIntegrity      bool   // ❌ Not guaranteed
    FaultDetection      bool   // ❌ Impossible
    UserProtection      bool   // ❌ Limited
    SecurityLevel       string // "Partial security"
}
```

#### With Dispute Game
```go
type FullSecurity struct {
    DataAvailability    bool   // ✅ Guaranteed
    StateIntegrity      bool   // ✅ Guaranteed
    FaultDetection      bool   // ✅ Possible
    UserProtection      bool   // ✅ Complete
    SecurityLevel       string // "Complete security"
}
```

## 4. Optimism's Security Model

### 4.1 Overall Security Architecture

```go
type OptimismSecurityModel struct {
    // Data Layer
    DataLayer struct {
        Sequencer    string // "Batch submission"
        Batcher      string // "Transaction compression"
        L1Storage    string // "Data availability"
    }

    // Validation Layer
    ValidationLayer struct {
        Proposer     string // "State root submission"
        Challenger   string // "Challenge incorrect states"
        DisputeGame  string // "Fault Proof game"
    }

    // Security Layer
    SecurityLayer struct {
        DataAvailability    bool   // ✅ Guaranteed by batch
        StateIntegrity      bool   // ✅ Guaranteed by Dispute Game
        FaultProof          bool   // ✅ Guaranteed by challenge system
    }
}
```

### 4.2 Data Flow

```
L2 Transactions
    ↓
Sequencer (op-batcher) → Submit batch to L1 BatchInbox
    ↓
L2 block creation and state calculation
    ↓
Proposer (op-proposer) → Submit output root to L1 DisputeGameFactory
    ↓
op-challenger detects via L1 head subscription
    ↓
Challenge incorrect output root
    ↓
Fault Proof game progression
    ↓
Recovery to correct state
```

## 5. Comparison with Other L2 Solutions

### 5.1 ZK-Rollup

#### Characteristics
```go
type ZKRollup struct {
    Submission      string // "Submit proof only"
    Finality        string // "Immediate finality"
    Computation     string // "Complex proof generation"
    SupportedOps    string // "Specific operations only"
    Cost            string // "High proof generation cost"
}
```

#### Pros and Cons
- **Pros**: Immediate finality guarantee, high security
- **Cons**: Complex proof generation, specific operations only, high cost

### 5.2 Optimistic Rollup (Optimism)

#### Characteristics
```go
type OptimisticRollup struct {
    Submission      string // "Batch + state root submission"
    Finality        string // "Finality after challenge period"
    Computation     string // "Simple structure"
    SupportedOps    string // "All operations supported"
    Cost            string // "Two submission costs"
}
```

#### Pros and Cons
- **Pros**: All operations supported, simple structure, low cost
- **Cons**: Challenge period required, two submission costs

## 6. Real Risk Scenarios

### 6.1 Scenario 1: Malicious Sequencer

#### Situation
```
1. Malicious sequencer performs incorrect calculation on L2
2. Transaction batch is correctly submitted to L1
3. But L2 state is incorrect
```

#### With Batch Only
```
❌ No one can detect the incorrect state
❌ User asset loss occurs
❌ No recovery possible
```

#### With Dispute Game
```
✅ op-challenger detects incorrect state
✅ Immediately starts challenge
✅ Proves correct state through Fault Proof game
✅ Rolls back incorrect state
✅ Protects user assets
```

### 6.2 Scenario 2: Software Bug

#### Situation
```
1. Bug occurs in L2 node software
2. Incorrect state calculation under specific conditions
3. Batch data is normal but state is erroneous
```

#### With Batch Only
```
❌ Cannot detect incorrect state due to bug
❌ Affects all users
❌ Manual intervention required
```

#### With Dispute Game
```
✅ Automatically detects incorrect state
✅ Immediately challenges and recovers
✅ Minimizes user impact
✅ Automated recovery system
```

## 7. Conclusion

### 7.1 Core Reasons Why Dispute Game Submission Is Necessary

1. **State Integrity Verification**: Batch data alone cannot verify the correctness of L2 state
2. **Calculation Error Detection**: Detect and recover from calculation errors that may occur on L2
3. **Security Guarantee**: Complete protection against malicious behavior
4. **User Protection**: Prevent user asset loss due to incorrect state
5. **Automated Recovery**: Automatically recover system without manual intervention

### 7.2 Cost vs Security Balance

```go
type CostSecurityBalance struct {
    HigherCost     bool   // ✅ Two submission costs
    CompleteSecurity bool   // ✅ Complete security guarantee
    UserProtection  bool   // ✅ Complete user asset protection
    Automation      bool   // ✅ Automated recovery system
    Tradeoff        string // "Higher cost but complete security"
}
```

### 7.3 Optimism's Choice

Optimism chose the **"Optimistic Rollup"** model to provide:

- **All Operations Support**: EVM compatibility enabling all smart contract execution
- **Simple Structure**: Simple structure without complex proof generation
- **Complete Security**: Complete security guarantee through two submissions
- **User-Friendly**: High throughput and low cost

**In conclusion, Dispute Game submission, while costly, is an essential element that completely guarantees Optimism's security.**
