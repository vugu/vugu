package main

import (
    "path"
    "strings"

    "github.com/vugu/vgrouter"
    "github.com/vugu/vugu"
    "github.com/vugu/vugu/js"
)

func vuguSetup(buildEnv *vugu.BuildEnv, eventEnv vugu.EventEnv) vugu.Builder {

    app := &App{
        Router: vgrouter.New(eventEnv),
    }

    // if there is a fragment when the page is loaded we go into fragment mode
    if strings.HasPrefix(js.Global().Get("window").Get("location").Get("hash").String(), "#") {
        app.Router.SetUseFragment(true)
    } else {
        // otherwise we set the path prefix
        browserPath := path.Clean("/" + js.Global().Get("window").Get("location").Get("pathname").String())
        pathPrefix := "/" + strings.Split(strings.TrimPrefix(browserPath, "/"), "/")[0]
        app.Router.SetPathPrefix(pathPrefix)
    }

    buildEnv.SetWireFunc(func(b vugu.Builder) {
        if c, ok := b.(vgrouter.NavigatorSetter); ok {
            c.NavigatorSet(app.Router)
        }
    })

    root := &Root{}
    buildEnv.WireComponent(root)

    app.Router.MustAddRouteExact("/", vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
        root.Body = &Products{}
    }))
    app.Router.MustAddRouteExact("/create", vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
        root.Body = &Create{}
    }))

    err := app.Router.ListenForPopState()
    if err != nil {
        panic(err)
    }

    err = app.Router.Pull()
    if err != nil {
        panic(err)
    }

    return root
}

// App has shared application stuff in it - for now it's a science project.
type App struct {
    *vgrouter.Router
}
