//go:build js && wasm

package js

import sjs "syscall/js"

// convertAnyToSJS converts any aliased value into its syscall/js counterpart.
func convertAnyToSJS(arg any) any {
	switch arg := arg.(type) {
	case Value:
		return sjs.Value(arg)
	case Func:
		return arg.f
	}

	return arg
}

// convertSliceToSJS takes a []any and converts every aliased value into its syscall/js counterpart.
//
// This is a bit dirty, as it may modify the given slice.
func convertSliceToSJS(args []any) []any {
	for i, arg := range args {
		switch arg := arg.(type) {
		case Value:
			args[i] = sjs.Value(arg)
		case Func:
			args[i] = arg.f
		}
	}

	return args
}

// convertMapToSJS takes a map[T]any and converts every aliased value into its syscall/js counterpart.
//
// This is a bit dirty, as it may modify the given map.
func convertMapToSJS[T comparable](args map[T]any) map[T]any {
	for k, arg := range args {
		switch arg := arg.(type) {
		case Value:
			args[k] = sjs.Value(arg)
		case Func:
			args[k] = arg.f
		}
	}

	return args
}
