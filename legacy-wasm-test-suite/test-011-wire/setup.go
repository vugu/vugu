package main

import "github.com/vugu/vugu"

func vuguSetup(buildEnv *vugu.BuildEnv, eventEnv vugu.EventEnv) vugu.Builder {

	var counter Counter

	buildEnv.SetWireFunc(func(b vugu.Builder) {
		if c, ok := b.(CounterSetter); ok {
			c.CounterSet(&counter)
		}
	})

	ret := &Root{}
	buildEnv.WireComponent(ret)

	return ret
}
