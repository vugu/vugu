package main

type Counter struct {
	c int
}

func (c *Counter) NextCount() int {
	c.c++
	return c.c
}

type CounterRef struct{ *Counter }
type CounterSetter interface{ CounterSet(c *Counter) }

func (cr *CounterRef) CounterSet(c *Counter) { cr.Counter = c }
