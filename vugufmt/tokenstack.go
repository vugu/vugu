package vugufmt

import "golang.org/x/net/html"

// tokenStack is a stack of nodes.
type tokenStack []*html.Token

// pop pops the stack. It will panic if s is empty.
func (s *tokenStack) pop() *html.Token {
	i := len(*s)
	n := (*s)[i-1]
	*s = (*s)[:i-1]
	return n
}

// push inserts a node
func (s *tokenStack) push(n *html.Token) {
	i := len(*s)
	(*s) = append(*s, nil)
	(*s)[i] = n
}

// top returns the most recently pushed node, or nil if s is empty.
func (s *tokenStack) top() *html.Token {
	if i := len(*s); i > 0 {
		return (*s)[i-1]
	}
	return nil
}

// index returns the index of the top-most occurrence of n in the stack, or -1
// if n is not present.
func (s *tokenStack) index(n *html.Token) int {
	for i := len(*s) - 1; i >= 0; i-- {
		if (*s)[i] == n {
			return i
		}
	}
	return -1
}

// insert inserts a node at the given index.
func (s *tokenStack) insert(i int, n *html.Token) {
	(*s) = append(*s, nil)
	copy((*s)[i+1:], (*s)[i:])
	(*s)[i] = n
}

// remove removes a node from the stack. It is a no-op if n is not present.
func (s *tokenStack) remove(n *html.Token) {
	i := s.index(n)
	if i == -1 {
		return
	}
	copy((*s)[i:], (*s)[i+1:])
	j := len(*s) - 1
	(*s)[j] = nil
	*s = (*s)[:j]
}
