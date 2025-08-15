const { ethers } = require("ethers");

/**
 * Merkle 트리 유틸리티 클래스
 */
class MerkleTree {
    constructor() {
        this.leaves = [];
        this.tree = [];
    }

    /**
     * 리프 노드 추가
     * @param {string} leaf - 리프 노드 해시
     */
    addLeaf(leaf) {
        this.leaves.push(leaf);
    }

    /**
     * 여러 리프 노드 추가
     * @param {string[]} leaves - 리프 노드 해시 배열
     */
    addLeaves(leaves) {
        this.leaves.push(...leaves);
    }

    /**
     * Merkle 트리 구축
     * @returns {string} 루트 해시
     */
    buildTree() {
        if (this.leaves.length === 0) {
            throw new Error("No leaves to build tree");
        }

        // 리프 노드를 해시로 변환
        let level = this.leaves.map(leaf =>
            ethers.utils.isHexString(leaf) ? leaf : ethers.utils.keccak256(leaf)
        );

        this.tree = [level];

        // 트리 레벨을 위로 올라가며 구축
        while (level.length > 1) {
            const nextLevel = [];

            for (let i = 0; i < level.length; i += 2) {
                const left = level[i];
                const right = i + 1 < level.length ? level[i + 1] : left;

                const combined = ethers.utils.solidityPack(
                    ['bytes32', 'bytes32'],
                    [left, right]
                );
                const hash = ethers.utils.keccak256(combined);
                nextLevel.push(hash);
            }

            level = nextLevel;
            this.tree.push(level);
        }

        return level[0]; // 루트 해시
    }

    /**
     * Merkle 증명 생성
     * @param {string} leaf - 증명할 리프 노드
     * @returns {object} Merkle 증명
     */
    getProof(leaf) {
        const leafHash = ethers.utils.isHexString(leaf) ? leaf : ethers.utils.keccak256(leaf);
        const leafIndex = this.leaves.findIndex(l =>
            (ethers.utils.isHexString(l) ? l : ethers.utils.keccak256(l)) === leafHash
        );

        if (leafIndex === -1) {
            throw new Error("Leaf not found in tree");
        }

        const proof = [];
        let index = leafIndex;

        for (let level = 0; level < this.tree.length - 1; level++) {
            const levelNodes = this.tree[level];
            const isRightNode = index % 2 === 1;
            const siblingIndex = isRightNode ? index - 1 : index + 1;

            if (siblingIndex < levelNodes.length) {
                proof.push({
                    hash: levelNodes[siblingIndex],
                    isRight: !isRightNode
                });
            }

            index = Math.floor(index / 2);
        }

        return {
            leaf: leafHash,
            proof: proof,
            root: this.tree[this.tree.length - 1][0]
        };
    }

    /**
     * Merkle 증명 검증
     * @param {string} leaf - 리프 노드
     * @param {object} proof - Merkle 증명
     * @param {string} root - 루트 해시
     * @returns {boolean} 검증 결과
     */
    static verifyProof(leaf, proof, root) {
        let computedHash = ethers.utils.isHexString(leaf) ? leaf : ethers.utils.keccak256(leaf);

        for (const proofElement of proof) {
            const left = proofElement.isRight ? proofElement.hash : computedHash;
            const right = proofElement.isRight ? computedHash : proofElement.hash;

            const combined = ethers.utils.solidityPack(
                ['bytes32', 'bytes32'],
                [left, right]
            );
            computedHash = ethers.utils.keccak256(combined);
        }

        return computedHash === root;
    }

    /**
     * 두 자식 해시로부터 부모 해시 계산
     * @param {string} leftChild - 왼쪽 자식 해시
     * @param {string} rightChild - 오른쪽 자식 해시
     * @returns {string} 부모 해시
     */
    static hashChildren(leftChild, rightChild) {
        const combined = ethers.utils.solidityPack(
            ['bytes32', 'bytes32'],
            [leftChild, rightChild]
        );
        return ethers.utils.keccak256(combined);
    }

    /**
     * 상태 트리에서 자식 노드 찾기
     * @param {string} stateRoot - 상태 루트
     * @param {object} stateTree - 상태 트리 데이터
     * @returns {object} 자식 노드들
     */
    static findChildrenForRoot(stateRoot, stateTree) {
        // 실제 구현에서는 상태 트리를 순회하여 올바른 자식들을 찾아야 함
        // 여기서는 시뮬레이션을 위한 간단한 구현

        const possibleChildren = [];

        // 상태 트리의 모든 노드 쌍을 확인
        for (let i = 0; i < stateTree.length - 1; i += 2) {
            const left = stateTree[i];
            const right = stateTree[i + 1] || stateTree[i];

            const computedRoot = this.hashChildren(left, right);

            if (computedRoot === stateRoot) {
                return { leftChild: left, rightChild: right };
            }

            possibleChildren.push({ leftChild: left, rightChild: right });
        }

        // 정확한 매치가 없으면 첫 번째 가능한 조합 반환
        if (possibleChildren.length > 0) {
            return possibleChildren[0];
        }

        // 기본값 반환
        return {
            leftChild: ethers.utils.keccak256("left"),
            rightChild: ethers.utils.keccak256("right")
        };
    }

    /**
     * 간단한 상태 트리 생성 (시뮬레이션용)
     * @param {number} size - 트리 크기
     * @returns {string[]} 상태 트리 노드들
     */
    static generateMockStateTree(size = 8) {
        const tree = [];

        for (let i = 0; i < size; i++) {
            const node = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(`node_${i}`));
            tree.push(node);
        }

        return tree;
    }

    /**
     * 트리 정보 출력
     */
    printTree() {
        console.log("Merkle Tree Structure:");
        this.tree.forEach((level, index) => {
            console.log(`Level ${index}: ${level.length} nodes`);
            level.forEach((hash, i) => {
                console.log(`  ${i}: ${hash}`);
            });
        });
    }
}

module.exports = MerkleTree;
