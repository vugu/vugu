package main

const amp = "&amp;"

type S string

func (s S) String() string { return "S-HERE:" + string(s) }
