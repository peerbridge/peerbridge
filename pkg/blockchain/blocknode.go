package blockchain

import (
	"errors"
)

type BlockNode struct {
	Block *Block

	Parent   *BlockNode
	Children *[]BlockNode
}

// Perform an iterative BFS to find the chain endpoint.
// This searches for nodes with the maximum height in the tree.
// If there are two or more nodes with equal max height,
// we use the node with the highest cumulative difficulty.
func (n *BlockNode) FindLongestChainEndpoint() *BlockNode {
	endpoint := n

	queue := append([]BlockNode{}, *n.Children...)
	var nextNode BlockNode

	for 0 < len(queue) {
		nextNode, queue = queue[0], queue[1:]

		// If the node has children, we don't need to evaluate
		// the height or cumulative difficulty, since the children
		// will always have height + 1
		if len(*nextNode.Children) > 0 {
			queue = append(queue, *nextNode.Children...)
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
			endpoint = &nextNode
			continue
		}

		// Otherwise, the heights are equal. We compare the
		// cumulative difficulties and choose the block with the
		// higher cumulative difficulty
		if *nextNode.Block.CumulativeDifficulty > *endpoint.Block.CumulativeDifficulty {
			endpoint = &nextNode
			continue
		}
	}

	return endpoint
}

// Get a block from the tree using its id. Note that
// this search is unidirectional, from the given node.
// This performs an iterative BFS.
func (n *BlockNode) GetBlockNodeByBlockID(id BlockID) (*BlockNode, error) {
	queue := []BlockNode{*n}
	var nextNode BlockNode

	for 0 < len(queue) {
		nextNode, queue = queue[0], queue[1:]

		if nextNode.Block.ID == id {
			// Block node found
			return &nextNode, nil
		}

		if len(*nextNode.Children) > 0 {
			// Keep searching
			queue = append(queue, *nextNode.Children...)
		}
	}

	return nil, errors.New("Block not found!")
}

// Check if the tree contains a given block (by id).
// Note that this search is unidirectional, from the
// given node. This performs an iterative BFS.
func (n *BlockNode) ContainsBlockByID(id BlockID) bool {
	_, err := n.GetBlockNodeByBlockID(id)
	return err == nil
}

// Check if the tree contains a given block. Note that
// this search is unidirectional, from the given node.
// This performs an iterative BFS.
func (n *BlockNode) ContainsBlock(b *Block) bool {
	_, err := n.GetBlockNodeByBlockID(b.ID)
	return err == nil
}

// Insert a given block into the tree. Note that this
// method will throw an error if the parent node could
// not be found. This method performs a forward
// iterative BFS.
func (n *BlockNode) InsertBlock(b *Block) (*BlockNode, error) {
	// Get the parent block node
	parentNode, err := n.GetBlockNodeByBlockID(b.ParentID)
	if err != nil {
		// Parent not found
		return nil, err
	}
	blockNode := BlockNode{
		Block:    b,
		Parent:   parentNode,
		Children: &[]BlockNode{},
	}
	*parentNode.Children = append(*parentNode.Children, blockNode)
	return &blockNode, nil
}
