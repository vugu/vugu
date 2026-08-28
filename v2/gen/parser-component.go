package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vugu/html"
	"github.com/vugu/vugu/v2"
)

// visitNodeComponentElement handles an element that is a call to a component
func (p *ParserGo) visitNodeComponentElement(state *parseGoState, n *html.Node) error {
	// components are just different so we handle all of our own vg-for vg-if and everything else

	// vg-for
	if v, _ := vgForExpr(n); v.expr != "" {
		if err := p.emitForExpr(state, n); err != nil {
			return err
		}
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// vg-if
	ife := vgIfExpr(n)
	if ife != "" {
		fmt.Fprintf(&state.buildBuf, "if %s {\n", ife)
		defer fmt.Fprintf(&state.buildBuf, "}\n")
	}

	// nodeName := n.OrigData // use original case of element
	// nodeNameParts := strings.Split(nodeName, ":")
	// if len(nodeNameParts) != 2 {
	// 	return fmt.Errorf("invalid component tag name %q must contain exactly one colon", nodeName)
	// }

	// // x.Y or just Y depending on if in same package
	// typeExpr := strings.Join(nodeNameParts, ".")
	// pkgPrefix := nodeNameParts[0] + "." // needed so we can calc pkg name for pkg.WhateverEvent
	// if nodeNameParts[0] == p.PackageName {
	// 	typeExpr = nodeNameParts[1]
	// 	pkgPrefix = ""
	// }

	// look for vg-type vg-pkg and vg-struct tags
	typeExpr := ""
	pkgPrefix := ""
	var vgStructFound bool
	fmt.Printf("OrigData = %s\n", n.OrigData)
	if n.OrigData == "vg-type" { // ths will need reworked once vugu/html is removed
		fmt.Printf("vg-type found\n")
		for _, a := range n.Attr {
			// vg-pkg is optional - we can infer the current package
			if a.Key == "vg-pkg" {
				if a.Val == p.PackageName { // vg-pkg matches the current package so there is prefix
					pkgPrefix = ""
				} else {
					pkgPrefix = a.Val + "."
				}
			}
			// vg-struct is mandatory
			if a.Key == "vg-struct" {
				typeExpr = a.Val
				vgStructFound = true
			}
		}
		if !vgStructFound {
			return fmt.Errorf("vg-type found but no vg-struct attribute")
		}
		fmt.Printf("vg-pkg = %s vg-struct = %s\n", pkgPrefix, typeExpr)
	}

	compKeyID := compHashCounted(p.StructType + "." + n.OrigData)

	fmt.Fprintf(&state.buildBuf, "{\n")
	defer fmt.Fprintf(&state.buildBuf, "}\n")

	keyExpr := vgKeyExpr(n)
	if keyExpr != "" {
		fmt.Fprintf(&state.buildBuf, "vgcompKey := vugu.MakeCompKey(0x%X^vgin.CurrentPositionHash(), %s)\n", compKeyID, keyExpr)
	} else {
		fmt.Fprintf(&state.buildBuf, "vgcompKey := vugu.MakeCompKey(0x%X^vgin.CurrentPositionHash(), vgiterkey)\n", compKeyID)
	}
	fmt.Fprintf(&state.buildBuf, "// ask BuildEnv for prior instance of this specific component\n")
	fmt.Fprintf(&state.buildBuf, "vgcomp, _ := vgin.BuildEnv.CachedComponent(vgcompKey).(*%s)\n", typeExpr)
	fmt.Fprintf(&state.buildBuf, "if vgcomp == nil {\n")
	fmt.Fprintf(&state.buildBuf, "// create new one if needed\n")
	fmt.Fprintf(&state.buildBuf, "vgcomp = new(%s)\n", typeExpr)
	fmt.Fprintf(&state.buildBuf, "vgin.BuildEnv.WireComponent(vgcomp)\n")
	fmt.Fprintf(&state.buildBuf, "}\n")
	fmt.Fprintf(&state.buildBuf, "vgin.BuildEnv.UseComponent(vgcompKey, vgcomp) // ensure we can use this in the cache next time around\n")

	// now that we have vgcomp with the right type and a correct value, we can declare the vg-var if specified
	if vgv := vgVarExpr(n); vgv != "" {
		fmt.Fprintf(&state.buildBuf, "var %s = vgcomp // vg-var\n", vgv)

		// NOTE: It's a bit too much to have "unused variable" errors coming from a Vugu code-generated file,
		// too far off the beaten path of making "type-safe HTML templates with Go".  It makes sense with
		// hand-written Go code but I don't think so here.
		fmt.Fprintf(&state.buildBuf, "_ = %s\n", vgv) // avoid unused var error
	}

	didAttrMap := false

	// look for vg-field
	// Do we support c.Field=blah or just Field=blah. Only the later ATM
	var vgFieldFound bool
	for _, a := range n.Attr {
		if a.Key == "vg-field" {
			fmt.Fprintf(&state.buildBuf, "// %s = \"%s\"\n", a.Key, a.Val)
			fmt.Fprintf(&state.buildBuf, "vgcomp.%s\n", a.Val)
			vgFieldFound = true
			break
		}
	}

	if !vgFieldFound {
		// dynamic attrs
		dynExprMap, dynExprMapKeys := dynamicVGAttrExpr(n)
		for _, k := range dynExprMapKeys {
			fmt.Printf("DYNAMIC ATTR\n")
			// if k == "" {
			// 	return fmt.Errorf("invalid empty dynamic attribute name on component %#v", n)
			// }

			valExpr := dynExprMap[k]

			// if starts with upper case, it's a field name
			if hasUpperFirst(k) {
				// we now ignore this - its replaced by vg-field
				// fmt.Fprintf(&state.buildBuf, "vgcomp.%s = %s\n", k, valExpr)
			} else {
				// otherwise we use an "AttrMap"
				if !didAttrMap {
					didAttrMap = true
					fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap = make(map[string]interface{}, 8)\n")
				}
				fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap[%q] = %s\n", k, valExpr)
			}

		}
	}

	// static attrs
	vgAttrs := staticVGAttr(n.Attr)
	for _, a := range vgAttrs {
		fmt.Printf("STATIC ATTR\n")
		// if starts with upper case, it's a field name
		if hasUpperFirst(a.Key) {
			fmt.Fprintf(&state.buildBuf, "vgcomp.%s = %q\n", a.Key, a.Val)
		} else {
			// otherwise we use an "AttrMap"
			if !didAttrMap {
				didAttrMap = true
				fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap = make(map[string]interface{}, 8)\n")
			}
			fmt.Fprintf(&state.buildBuf, "vgcomp.AttrMap[%q] = %q\n", a.Key, a.Val)
		}
	}

	// component events
	// NOTE: We keep component events really simple and the @ is just a thin wrapper around a field assignment:
	//     <pkg:Comp @Something="log.Println(event)"></pkg:Comp>
	// is shorthand for:
	//     <pkg:Comp :Something='func(event pkg.SomethingEvent) { log.Println(event) }'></pkg:Comp>
	//
	// I considered using the handler interface function approach for this, but it would mean
	// SomethingHandlerFunc would have to exist as a type, with a SomethingHandle method, which
	// implements a SomethingHandler interface, so the type of Comp.Something could be SomethingHandler,
	// and the emitted code could be vgcomp.Something = pkg.SomethingHandlerFunc(func...)
	// But that's two additional types and a method for every event.  I'm very concerned that it will
	// make component events feel crufty and arduous to implement (unless we could find a good way
	// to automatically generate those when missing - that's a possibility - actually I think
	// I'm going to try this, see https://github.com/vugu/vugu/v2/issues/128).
	// But this this way with a func you can just do
	// type SomethingEvent struct { /* whatever relevant data */ } and then define your field on
	// your component as Something func(SomethingEvent) - still type-safe but very straightforward.
	// So far it seems like the best approach.

	eventMap, eventKeys := vgEventExprs(n)
	for _, k := range eventKeys {
		expr := eventMap[k]
		// fmt.Fprintf(&state.buildBuf, "vgcomp.%s = func(event %s%sEvent){%s}\n", k, pkgPrefix, k, expr)
		// switched to using interfaces
		fmt.Fprintf(&state.buildBuf, "vgcomp.%s = %s%sFunc(func(event %s%sEvent){%s})\n", k, pkgPrefix, k, pkgPrefix, k, expr)
	}

	// NOTE: vugugen:slot might come in really handy, have to work out the types involved - update: as it stands, this won't be needed.

	// slots:

	// scan children and see if it's default slot mode or vg-slot tags
	foundTagSlot, foundDefSlot := false, false
	var foundTagSlotNames []string
	for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

		// non-ws text means default slot
		if childN.Type == html.TextNode {
			if strings.TrimSpace(childN.Data) != "" {
				foundDefSlot = true
			}
			continue
		}

		// ignore comments
		if childN.Type == html.CommentNode {
			continue
		}

		// should only be element at this point
		if childN.Type != html.ElementNode {
			return fmt.Errorf("in tag %q unexpected node found where slot expected: %#v", n.Data, childN)
		}

		if childN.Data == "vg-slot" {
			foundTagSlot = true
			name := strings.TrimSpace(vgSlotName(childN))
			if name != "" {
				foundTagSlotNames = append(foundTagSlotNames, name)
			}
		} else {
			foundDefSlot = true
		}
	}

	// now process slot(s) appropriately according to format
	switch {

	case foundTagSlot && foundDefSlot:
		return fmt.Errorf("in tag %q found both vg-slot and other markup, only one or the other is allowed", n.Data)

	case foundTagSlot:

		// NOTE:
		// <vg-slot name="X"> will assign to vgcomp.X
		// <vg-slot name='X[Y]'> will assume X is of type map[string]Builder and create the map and then assign with X[Y] =

		// find any names with map expressions and clear the maps
		sort.Strings(foundTagSlotNames)
		slotMapInited := make(map[string]bool)
		for _, slotName := range foundTagSlotNames {
			slotNameParts := strings.Split(slotName, "[") // check for map expr
			if len(slotNameParts) > 1 {                   // if map
				if slotMapInited[slotNameParts[0]] { // if not already initialized
					continue
				}
				slotMapInited[slotNameParts[0]] = true

				// if nil create map, otherwise reuse
				fmt.Fprintf(&state.buildBuf, "if vgcomp.%s == nil {\n", slotNameParts[0])
				fmt.Fprintf(&state.buildBuf, "    vgcomp.%s = make(map[string]vugu.Builder)\n", slotNameParts[0])
				fmt.Fprintf(&state.buildBuf, "} else {\n")
				fmt.Fprintf(&state.buildBuf, "    for k := range vgcomp.%s { delete(vgcomp.%s, k) }\n", slotNameParts[0], slotNameParts[0])
				fmt.Fprintf(&state.buildBuf, "}\n")
			}
		}

		// iterate over children
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

			// ignore white space and coments
			if childN.Type == html.CommentNode ||
				(childN.Type == html.TextNode && strings.TrimSpace(childN.Data) == "") {
				continue
			}

			if childN.Type != html.ElementNode { // should be impossible from foundTagSlot check above, just making sure
				panic(fmt.Errorf("unexpected non-element found where vg-slot should be: %#v", childN))
			}

			if childN.Data != "vg-slot" { // should also be imposible
				panic(fmt.Errorf("unexpected element found where vg-slot should be: %#v", childN))
			}

			slotName := strings.TrimSpace(vgSlotName(childN))
			if slotName == "" {
				return fmt.Errorf("found vg-slot tag without a 'name' attribute, the name is required")
			}

			fmt.Fprintf(&state.buildBuf, "vgcomp.%s = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {\n", slotName)
			fmt.Fprintf(&state.buildBuf, "vgn := &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type:vugu.VGNodeType(%d)}}\n", vugu.ElementNode)
			fmt.Fprintf(&state.buildBuf, "vgout = &vugu.BuildOut{}\n")
			fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn)\n")
			fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n")
			fmt.Fprintf(&state.buildBuf, "\n")

			// iterate over children and do the usual with each one
			for innerChildN := childN.FirstChild; innerChildN != nil; innerChildN = innerChildN.NextSibling {
				err := p.visitDefaultByType(state, innerChildN)
				if err != nil {
					return err
				}
			}

			fmt.Fprintf(&state.buildBuf, "return\n")
			fmt.Fprintf(&state.buildBuf, "})\n")

		}

	case foundDefSlot:
		fmt.Fprintf(&state.buildBuf, "vgcomp.DefaultSlot = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *vugu.BuildOut) {\n")
		// vgn is the equivalent of a vg-template tag and becomes the contents of vgout.Out and the vgparent
		fmt.Fprintf(&state.buildBuf, "vgn := &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Type: vugu.VGNodeType(%d)}}\n", vugu.ElementNode)
		fmt.Fprintf(&state.buildBuf, "vgout = &vugu.BuildOut{}\n")
		fmt.Fprintf(&state.buildBuf, "vgout.Out = append(vgout.Out, vgn)\n")
		fmt.Fprintf(&state.buildBuf, "vgparent := vgn; _ = vgparent\n")
		fmt.Fprintf(&state.buildBuf, "\n")

		// iterate over children and do the usual with each one
		for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {
			err := p.visitDefaultByType(state, childN)
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(&state.buildBuf, "return\n")
		fmt.Fprintf(&state.buildBuf, "})\n")

	default:
		// nothing meaningful inside this component tag
	}

	// // keep track of contents for default slot
	// var defSlotNodes []*html.Node
	// defSlotMode := false // start off not in default slot mode and look for <vg-slot> tags

	// // loop over all component children
	// for childN := n.FirstChild; childN != nil; childN = childN.NextSibling {

	// 	if !defSlotMode {

	// 		// anything not an element just add to the list for default
	// 		if childN.Type != html.ElementNode {
	// 			defSlotNodes = append(defSlotNodes, childN)
	// 			continue
	// 		}

	// 		if childN.Data == "vg-slot" {

	// 		}

	// 	}

	// }

	// ignore whitespace
	// first non-slot, non-ws child, assume "DefaultSlot" (or whatever name) and consume rest of children
	// if vg-slot, then consume with specified name
	// <vg-slot name="SomeSlot"> <!-- field name syntax
	// <vg-slot name='SomeDynaSlot' index='"Row.FirstName"'> <!-- expression syntax, HM, NO
	// <vg-slot index='SomeDynaSlot["Row.FirstName"]'> <!-- maybe this - still annoying that we have to limit it to a map expression, but whatever
	// emit vgcomp.SlotName = vugu.NewBuilderFunc(func(vgin *vugu.BuildIn) (vgout *BuildOut, vgerr error) { ... })
	// and descend into children

	fmt.Fprintf(&state.buildBuf, "vgout.Components = append(vgout.Components, vgcomp)\n")
	fmt.Fprintf(&state.buildBuf, "vgn = &vugu.VGNode{VGNodeCommonCore: vugu.VGNodeCommonCore{Component:vgcomp}}\n")
	fmt.Fprintf(&state.buildBuf, "vgparent.AppendChild(vgn)\n")

	return nil
	// return fmt.Errorf("component tag not yet supported (%q)", nodeName)
}
