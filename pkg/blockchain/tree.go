package blockchain

import (
	"errors"

	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

var (
	ErrTransactionInTreeNotFound = errors.New("Transaction in tree not found!")
	ErrBlockInTreeNotFound       = errors.New("Block in tree not found!")
	ErrBlockAlreadyInTree        = errors.New("Block is already in tree!")
	ErrParentBlockNotInTree      = errors.New("Parent block is not in tree!")
	ErrAttemptChopNonRoot        = errors.New("Attempted to chop from a non-root node!")
)

type BlockTree struct {
	Block Block

	Parent   *BlockTree
	Children []*BlockTree
}

// Perform an iterative BFS to find the chain endpoints.
func (n *BlockTree) FindEndpoints() []*BlockTree {
	endpoints := []*BlockTree{}

	queue := append([]*BlockTree{}, n)
	var nextNode *BlockTree

	for 0 < len(queue) {
		nextNode, queue = queue[0], queue[1:]

		if len(nextNode.Children) > 0 {
			queue = append(queue, nextNode.Children...)
			continue
		}

		// Node has no children
		endpoints = append(endpoints, nextNode)
	}

	return endpoints
}

// Perform an iterative BFS to find the chain endpoint.
// This searches for nodes with the maximum height in the tree.
// If there are two or more nodes with equal max height,
// we use the node with the highest cumulative difficulty.
func (n *BlockTree) FindLongestChainEndpoint() *BlockTree {
	endpoint := n

	queue := append([]*BlockTree{}, n.Children...)
	var nextNode *BlockTree

	for 0 < len(queue) {
		nextNode, queue = queue[0], queue[1:]

		// If the node has children, we don't need to evaluate
		// the height or cumulative difficulty, since the children
		// will always have height + 1
		if len(nextNode.Children) > 0 {
			queue = append(queue, nextNode.Children...)
			continue
		}

		// If the found node height is less than the already
		// known highest node, the found node is not the chain
		// endpoint
		if nextNode.Block.Height < endpoint.Block.Height {
			continue
		}

		// If the height is higher than the already known highest
		// node, then this node could be our chain endpoint
		if nextNode.Block.Height > endpoint.Block.Height {
			endpoint = nextNode
			continue
		}

		// Otherwise, the heights are equal. We compare the
		// cumulative difficulties and choose the block with the
		// higher cumulative difficulty
		if nextNode.Block.CumulativeDifficulty > endpoint.Block.CumulativeDifficulty {
			endpoint = nextNode
			continue
		}
	}

	return endpoint
}

// Get a block from the tree using its id. Note that
// this search is unidirectional, from the given node.
// This performs an iterative BFS.
func (n *BlockTree) GetBlockTreeByBlockID(id encryption.SHA256HexString) (*BlockTree, error) {
	queue := []*BlockTree{n}
	var nextNode *BlockTree

	for 0 < len(queue) {
		nextNode, queue = queue[0], queue[1:]

		if nextNode.Block.ID == id {
			// Block node found
			return nextNode, nil
		}

		if len(nextNode.Children) > 0 {
			// Keep searching
			queue = append(queue, nextNode.Children...)
		}
	}

	return nil, ErrBlockInTreeNotFound
}

// Check if the tree contains a given block (by id).
// Note that this search is unidirectional, from the
// given node. This performs an iterative BFS.
func (n *BlockTree) ContainsBlockByID(id encryption.SHA256HexString) bool {
	_, err := n.GetBlockTreeByBlockID(id)
	return err == nil
}

// Insert a given block into the tree. Note that this
// method will throw an error if the parent node could
// not be found. This method performs a forward
// iterative BFS.
func (n *BlockTree) InsertBlock(b *Block) error {
	// Check if this block already exists
	if n.ContainsBlockByID(b.ID) {
		return ErrBlockAlreadyInTree
	}

	// Get the parent block node
	parentNode, err := n.GetBlockTreeByBlockID(*b.ParentID)
	if err != nil {
		// Parent not found
		return ErrParentBlockNotInTree
	}

	for _, child := range parentNode.Children {
		if child.Block.ID == b.ID {
			return ErrBlockAlreadyInTree
		}
	}

	if parentNode.Block.Height+1 != b.Height {
		panic("Parent node height should always be height - 1!")
	}
	if parentNode.Block.ID != *b.ParentID {
		panic("Parent node has wrong id!")
	}
	// Add the child to the tree and link it to the parent node
	blockTree := &BlockTree{
		Block:  *b,
		Parent: parentNode,
	}
	parentNode.Children = append(parentNode.Children, blockTree)
	return nil
}

// Get the longest branch in the current tree.
func (n *BlockTree) GetLongestBranch() []*BlockTree {
	endpoint := n.FindLongestChainEndpoint()

	// Compute the longest chain by going backwards
	longestBranch := &[]*BlockTree{endpoint}
	parent := endpoint.Parent
	for parent != nil {
		*longestBranch = append([]*BlockTree{parent}, *longestBranch...)
		parent = parent.Parent
	}

	return *longestBranch
}

// Get the branch leading to a specific node.
func (n *BlockTree) GetBranch(id encryption.SHA256HexString) (*[]*BlockTree, error) {
	endpoint, err := n.GetBlockTreeByBlockID(id)
	if err != nil {
		return nil, err
	}
	// Compute the chain by going backwards
	branch := &[]*BlockTree{endpoint}
	parent := endpoint.Parent
	for parent != nil {
		*branch = append([]*BlockTree{parent}, *branch...)
		parent = parent.Parent
	}
	return branch, nil
}

// The result of a chop operation.
type ChopResult struct {
	// The stem nodes which belong to the longest chain.
	StemNodes *[]*BlockTree
	// The orphaned nodes which belong to a shorter side chain.
	OrphanedNodes *[]*BlockTree
}

// Chop a block node's tree to a given length.
// This will chop off all nodes from the root side
// of the tree, until the given length is reached.
// As a result, the stem nodes (belonging to the longest
// chain) will be returned, as well as orphaned nodes
// from shorter side chains.
func (root *BlockTree) Chop(length int) (*BlockTree, *ChopResult, error) {
	if root.Parent != nil {
		return nil, nil, ErrAttemptChopNonRoot
	}

	result := &ChopResult{
		StemNodes:     &[]*BlockTree{},
		OrphanedNodes: &[]*BlockTree{},
	}

	longestBranch := append([]*BlockTree{}, root.GetLongestBranch()...)

	var newRoot *BlockTree
	for {
		// Make one step forward in the longest chain
		newRoot, longestBranch = longestBranch[0], longestBranch[1:]

		if len(longestBranch) <= length {
			break
		}

		// All children that are not in the longest chain are
		// marked as orphaned
		for _, child := range newRoot.Children {
			if child.Block.ID != longestBranch[0].Block.ID {
				*result.OrphanedNodes = append(*result.OrphanedNodes, child)
			}
			// Detach the child from its parent
			child.Parent = nil
		}

		// Detach the parent from its children
		newRoot.Children = nil
		newRoot.Parent = nil

		// The only child that is in the longest chain gets
		// into the "stem" nodes
		*result.StemNodes = append(*result.StemNodes, newRoot)
	}

	return newRoot, result, nil
}

func (n *BlockTree) GetTransactionByID(id encryption.SHA256HexString) (*Transaction, error) {
	queue := []*BlockTree{n}
	var nextNode *BlockTree

	for 0 < len(queue) {
		nextNode, queue = queue[0], queue[1:]

		for _, t := range nextNode.Block.Transactions {
			if id == t.ID {
				// Transaction found
				return &t, nil
			}
		}

		if len(nextNode.Children) > 0 {
			// Keep searching
			queue = append(queue, nextNode.Children...)
		}
	}

	return nil, ErrTransactionInTreeNotFound
}

func (n *BlockTree) ContainsTransactionByID(id encryption.SHA256HexString) bool {
	_, err := n.GetTransactionByID(id)
	return err == nil
}

func (n *BlockTree) GetLongestChainTransactionsForAccount(account secp256k1.PublicKeyHexString) []Transaction {
	transactions := []Transaction{}
	longestChain := n.GetLongestBranch()
	for _, cn := range longestChain {
		for _, t := range cn.Block.Transactions {
			if t.Receiver == account || t.Sender == account {
				transactions = append(transactions, t)
			}
		}
	}
	return transactions
}

// Compute the stake of an account.
func (n *BlockTree) Stake(
	p secp256k1.PublicKeyHexString,
	fromBlockID encryption.SHA256HexString,
	fromIsInclusive bool,
	toBlockID encryption.SHA256HexString,
	toIsInclusive bool,
) (*int64, error) {
	// Get the `from` block in the tree
	fromBlock, err := n.GetBlockTreeByBlockID(fromBlockID)
	if err != nil {
		return nil, err
	}

	// Get the chain until the `to` block in the tree
	branch, err := fromBlock.GetBranch(toBlockID)
	if err != nil {
		return nil, err
	}

	stake := int64(0)

	for _, chainNode := range *branch {
		block := chainNode.Block

		if block.ID == fromBlockID && !fromIsInclusive {
			continue
		}

		if block.ID == toBlockID && !toIsInclusive {
			continue
		}

		if block.Creator == p {
			stake += 100 // Block reward

			for _, t := range block.Transactions {
				// FIXME: Theoretically, this could overflow
				// with very high fees
				stake += int64(t.Fee) // Transaction fee grant
			}
		}

		for _, t := range block.Transactions {
			if t.Sender == p {
				// FIXME: Theoretically, this could overflow
				// with very high fees or balances
				stake -= int64(t.Balance)
				stake -= int64(t.Fee)
			}
			if t.Receiver == p {
				// FIXME: Theoretically, this could overflow
				// with very high balances
				stake += int64(t.Balance)
			}
		}
	}

	return &stake, nil
}
