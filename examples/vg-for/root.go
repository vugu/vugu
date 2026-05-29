package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
)

const numberOfDaysInWeek = 7

const (
	Monday    = "Monday"
	Tuesday   = "Tuesday"
	Wednesday = "Wednesday"
	Thursday  = "Thursday"
	Friday    = "Friday"
	Saturday  = "Saturday"
	Sunday    = "Sunday"
	First     = "1st"
	Second    = "2nd"
	Third     = "3rd"
	Fourth    = "4th"
	Fifth     = "5th"
	Sixth     = "6th"
	Last      = "Last"
)

// Root is the root component of the application in the DOM
// This example has two components an array and a map that hold the days of the week and a set of booleans to control the visibility of each for loop example
// Note that no fields are exported. This means the Javascript side cannot access these fields directly, instead it must use the getters
// see the Toggle* and Show* exported functions.
// This is much more robust approach as it allows the Root object to be restructured with less chance of breaking the JS side.
type Root struct {
	showForILoop     bool
	showKVLoop       bool
	showShortCutLoop bool
	showSortedLoop   bool
	showIteratorLoop bool
	daysInWeek       [numberOfDaysInWeek]string
	daysInWeekIth    map[string]string
}

// Init is a component lifecycle method. It is called when the component is created.
// We use it is to initialise the array and the map
// See:
// https://www.vugu.org/doc/components
func (c *Root) Init() {
	c.daysInWeek = [numberOfDaysInWeek]string{Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday}
	c.daysInWeekIth = map[string]string{First: Monday, Second: Tuesday, Third: Wednesday, Fourth: Thursday, Fifth: Friday, Sixth: Saturday, Last: Sunday}
}

// Toggle the visibility of the for i loop
func (c *Root) ToggleForILoop() bool {
	c.showForILoop = !c.showForILoop
	return c.showForILoop
}

// Return the visibility of the for i loop
func (c *Root) ShowForILoop() bool {
	return c.showForILoop
}

// return the value of the numberOfDaysInWeek constant
func (c *Root) DaysInWeek() int {
	return numberOfDaysInWeek
}

// Return the name of the ith day of the week (zero based)
func (c *Root) DayOfWeek(i int) string {
	return c.daysInWeek[i]
}

// Roggle the visibility of the key value loop
func (c *Root) ToggleKVLoop() bool {
	c.showKVLoop = !c.showKVLoop
	return c.showForILoop
}

// Return the visibility of the key value loop
func (c *Root) ShowKVLoop() bool {
	return c.showKVLoop
}

// Return the map that maps te 1st, 2nd etc day of the week to the day
func (c *Root) DaysInWeekIth() map[string]string {
	return c.daysInWeekIth
}

// Toggle the visibility of the shortcut version of the key value loop
func (c *Root) ToggleShortCutLoop() bool {
	c.showShortCutLoop = !c.showShortCutLoop
	return c.showShortCutLoop
}

// Return the visibility of the shortcut version of the key value loop
func (c *Root) ShowShortCutLoop() bool {
	return c.showShortCutLoop
}

// Toggle the visibility of the sorted key value loop
func (c *Root) ToggleSortedLoop() bool {
	c.showSortedLoop = !c.showSortedLoop
	return c.showSortedLoop
}

// Return the visibility of the sorted key value loop
func (c *Root) ShowSortedLoop() bool {
	return c.showSortedLoop
}

// Return a slice containing the days of the week sorted by key (1st, 2nd etc) order
func (c *Root) DaysInWeekIthSorted() []string {
	return slices.Sorted(maps.Keys(c.daysInWeekIth))
}

// Toggle the visibility of the range over an iterator loop
func (c *Root) ToggleIteratorLoop() bool {
	c.showIteratorLoop = !c.showIteratorLoop
	return c.showIteratorLoop
}

// Return the visibility of the range over an iterator loop
func (c *Root) ShowIteratorLoop() bool {
	return c.showIteratorLoop
}

// Create and return a single use iterator that iterates over the days of the week.
func (c *Root) DaysInWeekIterator() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i, v := range c.daysInWeek {
			fmt.Printf("i: %d, v:%s\n", i, v)
			if !yield(i, v) {
				return
			}
		}
	}
}
