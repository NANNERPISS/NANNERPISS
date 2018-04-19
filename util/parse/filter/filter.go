package filter

import (
	"golang.org/x/net/html"
)

// NodeFilter should returns true when the provided node matches the filter criteria.
type NodeFilter func(*html.Node) bool

// Chain combines multiple NodeFilters into a single NodeFilter.
func Chain(filters ...NodeFilter) NodeFilter {
	return func(n *html.Node) bool {
		for _, f := range filters {
			if !f(n) {
				return false
			}
		}

		return true
	}
}

// Type returns a NodeFilter that will match nodes that have the provided NodeType.
func Type(nodeType html.NodeType) NodeFilter {
	return func(n *html.Node) bool {
		return n.Type == nodeType
	}
}

// Type chains the previous NodeFilter to a Type NodeFilter.
func (f NodeFilter) Type(nodeType html.NodeType) NodeFilter {
	return Chain(f, Type(nodeType))
}

// Tag returns a NodeFilter that will match nodes that have the provided tag type.
func Tag(tag string) NodeFilter {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag
	}
}

// Tag chains the previous NodeFilter to a Tag NodeFilter.
func (f NodeFilter) Tag(tag string) NodeFilter {
	return Chain(f, Tag(tag))
}

// AttrKey returns a NodeFilter that will match nodes that have an attribute that matches the provided key.
func AttrKey(key string) NodeFilter {
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				return true
			}
		}

		return false
	}
}

// AttrKey chains the previous NodeFilter to an AttrKey NodeFilter.
func (f NodeFilter) AttrKey(key string) NodeFilter {
	return Chain(f, AttrKey(key))
}

// Attr returns a NodeFilter that will match nodes that have an attribute that matches the provided key and value.
func Attr(key string, val string) NodeFilter {
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				if a.Val == val {
					return true
				}
			}
		}

		return false
	}
}

// Attr chains the previous NodeFilter to an Attr NodeFilter.
func (f NodeFilter) Attr(key string, val string) NodeFilter {
	return Chain(f, Attr(key, val))
}

// ID returns a NodeFilter that will match nodes that have the provided ID attribute.
func ID(id string) NodeFilter {
	return Attr("id", id)
}

// ID chains the previous NodeFilter to an ID NodeFilter.
func (f NodeFilter) ID(id string) NodeFilter {
	return Chain(f, ID(id))
}

// Class returns a NodeFilter that will match nodes that have the provided class attribute.
func Class(class string) NodeFilter {
	return Attr("class", class)
}

// Class chains the previous NodeFilter to a Class NodeFilter.
func (f NodeFilter) Class(class string) NodeFilter {
	return Chain(f, Class(class))
}
