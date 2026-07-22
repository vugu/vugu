package main

import "github.com/vugu/vgrouter"

type Page2 struct {
	vgrouter.NavigatorRef
}

func (c *Page2) NavigateToPage(url string) {
	c.Navigate(url, nil)
}

func (c *Page2) NavigateToPageWithReplace(url string) {
	c.Navigate(url, nil, vgrouter.NavReplace)
}
