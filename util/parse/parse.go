package parse

import (
	"io"

	"github.com/NANNERPISS/NANNERPISS/util/parse/filter"

	"golang.org/x/net/html"
)

type ParseNode struct {
	*html.Node
}
type ParseNodes []*ParseNode

func Parse(r io.Reader) (*ParseNode, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return &ParseNode{node}, nil
}

// Interface to Node Slice
//
// bool will be false if i is not type *html.Node, *ParseNode, or ParseNodes.
func itons(i interface{}) (ParseNodes, bool) {
	switch v := i.(type) {
	case *html.Node:
		return ParseNodes{&ParseNode{v}}, true
	case *ParseNode:
		return ParseNodes{v}, true
	case ParseNodes:
		return v, true
	default:
		return nil, false
	}
}

// FindNode returns the first node that matches the provided NodeFilter.
//
// The node interface can be of type *html.Node, *ParseNode, or ParseNodes.
//
// nil will be returned if the node interface does not match any of these types.
func FindNode(node interface{}, filter filter.NodeFilter) *ParseNode {
	ns, ok := itons(node)
	if !ok {
		return nil
	}

	for _, n := range ns {
		if n == nil || n.Node == nil {
			continue
		}
		if filter(n.Node) {
			return n
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			nn := FindNode(c, filter)
			if nn != nil {
				return nn
			}
		}
	}

	return nil
}

// FindNode calls FindNode with the previous *ParseNode and the provided NodeFilter.
//
// This is a helper function to help chain function calls.
func (node *ParseNode) FindNode(filter filter.NodeFilter) *ParseNode {
	return FindNode(node, filter)
}

// FindNode calls FindNode with the previous ParseNodes and the provided NodeFilter.
//
// This is a helper function to help chain function calls.
func (node ParseNodes) FindNode(filter filter.NodeFilter) *ParseNode {
	return FindNode(node, filter)
}

// FindNodes returns all nodes that match the provided NodeFilter.
//
// The node interface can be of type *html.Node, *ParseNode, or ParseNodes.
//
// nil will be returned if the interface does not match any of these types.
func FindNodes(node interface{}, filter filter.NodeFilter) ParseNodes {
	ns, ok := itons(node)
	if !ok {
		return nil
	}

	var nodes ParseNodes
	for _, n := range ns {
		if n == nil || n.Node == nil {
			continue
		}
		if filter(n.Node) {
			nodes = append(nodes, n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			nn := FindNodes(c, filter)
			if nn != nil {
				nodes = append(nodes, nn...)
			}
		}
	}

	return nodes
}

// FindNodes calls FindNodes with the previous *ParseNode and the provided NodeFilter.
//
// This is a helper function to help chain function calls.
func (node *ParseNode) FindNodes(filter filter.NodeFilter) ParseNodes {
	return FindNodes(node, filter)
}

// FindNodes calls FindNodes with the previous ParseNodes and the provided NodeFilter.
//
// This is a helper function to help chain function calls.
func (node ParseNodes) FindNodes(filter filter.NodeFilter) ParseNodes {
	return FindNodes(node, filter)
}

// GetAttr returns the value associated with an attribute on the provided *ParseNode.
//
// The returned bool will be false if a matching attribute wasn't found.
func GetAttr(node *ParseNode, key string) (string, bool) {
	if node == nil {
		return "", false
	}

	for _, a := range node.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}

	return "", false
}

// GetAttr calls GetAttr with the previous *ParseNode and the provided string as the key.
func (node *ParseNode) GetAttr(key string) (string, bool) {
	return GetAttr(node, key)
}
