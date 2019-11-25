module github.com/vugu/vugu/full-test-data/test1

replace github.com/vugu/vugu => ../..

//replace github.com/vugu/vugu/domrender => ../..

require (
	github.com/vugu/vjson v0.0.0-20191111004939-722507e863cb
	github.com/vugu/vugu v0.0.0-00010101000000-000000000000
)

go 1.13
