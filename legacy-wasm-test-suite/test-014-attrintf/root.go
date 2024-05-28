package main

type TestStringer struct {
	str string
}

func (s *TestStringer) String() string {
	if s == nil {
		return ""
	}
	return s.str
}

type Root struct {
	StringVar      string
	IntVar         int
	TrueVar        bool
	FalseVar       bool
	StringNilPtr   *string
	Stringer       *TestStringer
	StringerNilPtr *TestStringer
}

func (c *Root) BeforeBuild() {
	c.StringVar = "aString"
	c.IntVar = 42
	c.TrueVar = true
	c.FalseVar = false
	c.StringNilPtr = nil
	c.Stringer = &TestStringer{
		str: "myString",
	}
	c.StringerNilPtr = nil
}
