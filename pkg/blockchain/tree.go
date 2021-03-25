package blockchain

import (
	"errors"
	"sync"

	"github.com/peerbridge/peerbridge/pkg/encryption"
)

type BlockTree struct {
	Block Block

	Parent   *BlockTree
	Children []*BlockTree

	lock sync.Mutex
}

// Perform an iterative BFS to find the chain endpoint.
// This searches for nodes with the maximum height in the tree.
// If there are two or more nodes with equal max height,
// we use the node with the highest cumulative difficulty.
func (n *BlockTree) FindLongestChainEndpoint() *BlockTree {
	n.lock.Lock()
	defer n.lock.Unlock()

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
	n.lock.Lock()
	defer n.lock.Unlock()

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

	return nil, errors.New("Block not found!")
}

// Check if the tree contains a given block (by id).
// Note that this search is unidirectional, from the
// given node. This performs an iterative BFS.
func (n *BlockTree) ContainsBlockByID(id encryption.SHA256HexString) bool {
	_, err := n.GetBlockTreeByBlockID(id)
	return err == nil
}

// Check if the tree contains a given block. Note that
// this search is unidirectional, from the given node.
// This performs an iterative BFS.
func (n *BlockTree) ContainsBlock(b *Block) bool {
	_, err := n.GetBlockTreeByBlockID(b.ID)
	return err == nil
}

// Insert a given block into the tree. Note that this
// method will throw an error if the parent node could
// not be found. This method performs a forward
// iterative BFS.
func (n *BlockTree) InsertBlock(b *Block) error {
	// Check if this block already exists
	if n.ContainsBlockByID(b.ID) {
		return errors.New("Block is already in tree!")
	}

	// Get the parent block node
	parentNode, err := n.GetBlockTreeByBlockID(*b.ParentID)
	if err != nil {
		// Parent not found
		return err
	}

	for _, child := range parentNode.Children {
		if child.Block.ID == b.ID {
			return errors.New("Block already in children!")
		}
	}

	if parentNode.Block.Height+1 != b.Height {
		panic("Parent node height should always be height - 1!")
	}
	if parentNode.Block.ID != *b.ParentID {
		panic("Parent node has wrong id!")
	}
	n.lock.Lock()
	// Add the child to the tree and link it to the parent node
	blockTree := &BlockTree{
		Block:  *b,
		Parent: parentNode,
	}
	parentNode.Children = append(parentNode.Children, blockTree)
	n.lock.Unlock()
	return nil
}

// Get the longest chain in the current tree.
func (n *BlockTree) GetLongestChain() []*BlockTree {
	endpoint := n.FindLongestChainEndpoint()

	// Compute the longest chain by going backwards
	longestChain := &[]*BlockTree{endpoint}
	parent := endpoint.Parent
	for parent != nil {
		*longestChain = append([]*BlockTree{parent}, *longestChain...)
		parent = parent.Parent
	}

	return *longestChain
}

// Get the chain leading to a specific node.
func (n *BlockTree) GetChain(id encryption.SHA256HexString) (*[]*BlockTree, error) {
	endpoint, err := n.GetBlockTreeByBlockID(id)
	if err != nil {
		return nil, err
	}
	// Compute the chain by going backwards
	chain := &[]*BlockTree{endpoint}
	parent := endpoint.Parent
	for parent != nil {
		*chain = append([]*BlockTree{parent}, *chain...)
		parent = parent.Parent
	}
	return chain, nil
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
//
// Note that this operation is in-place, which means that
// the root will be replaced by this operation, if
// the current tree exceeds the given length.
func (root *BlockTree) Chop(length int) (*BlockTree, *ChopResult, error) {
	if root.Parent != nil {
		return nil, nil, errors.New("Attempted to chop from a non-root node!")
	}

	result := &ChopResult{
		StemNodes:     &[]*BlockTree{},
		OrphanedNodes: &[]*BlockTree{},
	}

	longestChain := append([]*BlockTree{}, root.GetLongestChain()...)

	root.lock.Lock()
	var newRoot *BlockTree
	for {
		// Make one step forward in the longest chain
		newRoot, longestChain = longestChain[0], longestChain[1:]

		if len(longestChain) <= length {
			break
		}

		// All children that are not in the longest chain are
		// marked as orphaned
		for _, child := range newRoot.Children {
			if child.Block.ID != longestChain[0].Block.ID {
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
	root.lock.Unlock()

	return newRoot, result, nil
}

func (n *BlockTree) GetTransactionByID(id encryption.SHA256HexString) (*Transaction, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

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

	return nil, errors.New("Transaction not found!")
}

func (n *BlockTree) ContainsTransactionByID(id encryption.SHA256HexString) bool {
	_, err := n.GetTransactionByID(id)
	return err == nil
}

func (n *BlockTree) ContainsTransaction(t *Transaction) bool {
	_, err := n.GetTransactionByID(t.ID)
	return err == nil
}
