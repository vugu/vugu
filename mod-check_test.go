package vugu

import "testing"

func TestModChecker(t *testing.T) {

}

// type testRowList struct {
// 	rows []testRow
// 	n int
// }

// func (l *testRowList)Add(tr testRow) {
// }

type testRow struct {
	A string  `vugu:"modcheck"`
	B *string `vugu:"modcheck"`
}

func newTestRowMap() *testRowMap {
	return &testRowMap{m: make(map[string]*testRow)}
}

func (trm *testRowMap) ModCheck(mt *ModTracker, oldData interface{}) (isModified bool, newData interface{}) {
	oldn, _ := oldData.(int)
	return oldn == trm.n, trm.n
}

func (trm *testRowMap) Set(key string, v *testRow) {
	trm.m[key] = v
	trm.n++
}

func (trm *testRowMap) Get(key string) *testRow {
	return trm.m[key]
}

type testRowMap struct {
	m map[string]*testRow
	n int
}
