package main

type Root struct {
	Count   int
	Visible bool
	Text    string
}

func (c *Root) Switch() {
	c.Visible = !c.Visible
	c.Count++
}

func (c *Root) Click(msg string) {
	if c.Text != "" {
		c.Text += " "
	}
	c.Text += msg
}
