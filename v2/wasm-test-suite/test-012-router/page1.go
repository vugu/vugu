package main

import "github.com/vugu/vgrouter"

type Page1 struct {
	vgrouter.NavigatorRef
}

func (c *Page1) NavigateToPage(url string) {
	c.Navigate(url, nil)
}

func (c *Page1) NavigateToPageWithReplace(url string) {
	c.Navigate(url, nil, vgrouter.NavReplace)
}
