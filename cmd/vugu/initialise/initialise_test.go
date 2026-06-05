package initialise

import "testing"

func TestCleanTemplateData(t *testing.T) {
	// fill the template
	d := IndexHTMLData{}

	// WasmExecJSDir tests
	d.WasmExecJSDir = "/end/in/slash/"
	d.cleanTemplateData()

	if d.WasmExecJSDir != "/end/in/slash" {
		t.Fatal("Failed to remove training slash from \"/end/in/slash/\"")
	}

	d.WasmExecJSDir = ""
	d.cleanTemplateData()

	if d.WasmExecJSDir != "" {
		t.Fatalf("Failed to return an empty string when WasmExecJSDir is empty. Returned %q", d.WasmExecJSDir)
	}

	d.WasmExecJSDir = "."
	d.cleanTemplateData()

	if d.WasmExecJSDir != "" {
		t.Fatalf("Failed to return empty string when WasmExecJSDir was a period. Returned %q", d.WasmExecJSDir)
	}

	d.WasmExecJSDir = ".."
	d.cleanTemplateData()

	if d.WasmExecJSDir != ".." {
		t.Fatalf("Failed to leave unchanged when WasmExecJSDir was a .. Returned %q", d.WasmExecJSDir)
	}

	d.WasmExecJSDir = "/"
	d.cleanTemplateData()

	if d.WasmExecJSDir != "" {
		t.Fatalf("Failed to return empty string when WasmExecJSDir was a /. Returned %q", d.WasmExecJSDir)
	}

	// WasmMainDir tests
	d.WasmMainDir = "/end/in/slash/"
	d.cleanTemplateData()

	if d.WasmMainDir != "/end/in/slash" {
		t.Fatal("Failed to remove training slash from \"/end/in/slash/\"")
	}

	d.WasmMainDir = ""
	d.cleanTemplateData()

	if d.WasmMainDir != "" {
		t.Fatalf("Failed to return an empty string when WasmMainDir is empty. Returned %q", d.WasmMainDir)
	}

	d.WasmMainDir = "."
	d.cleanTemplateData()

	if d.WasmMainDir != "" {
		t.Fatalf("Failed to return empty string when WasmMainDir was a period. Returned %q", d.WasmMainDir)
	}

	d.WasmMainDir = ".."
	d.cleanTemplateData()

	if d.WasmMainDir != ".." {
		t.Fatalf("Failed to leave unchanged when WasmMainDir was a .. Returned %q", d.WasmMainDir)
	}

	d.WasmMainDir = "/"
	d.cleanTemplateData()

	if d.WasmMainDir != "" {
		t.Fatalf("Failed to return empty string when WasmMainDir was a /. Returned %q", d.WasmMainDir)
	}

	// WasmBinaryName tests
	d.WasmBinaryName = "/main.js/"
	d.cleanTemplateData()

	if d.WasmBinaryName != "main.js" {
		t.Fatal("Failed to remove training slash from \"/main.js/\"")
	}

	d.WasmBinaryName = "main.js/"
	d.cleanTemplateData()

	if d.WasmBinaryName != "main.js" {
		t.Fatal("Failed to remove training slash from \"main.js/\"")
	}

	d.WasmBinaryName = "."
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName when WasmMainDir was a period. Returned %q", d.WasmBinaryName)
	}

	d.WasmBinaryName = ".."
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName WasmMainDir was a .. Returned %q", d.WasmBinaryName)
	}

	d.WasmBinaryName = "/"
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName when WasmMainDir was a /. Returned %q", d.WasmBinaryName)
	}

	d.WasmBinaryName = ""
	d.cleanTemplateData()

	if d.WasmBinaryName != defaultWasmBinaryName {
		t.Fatalf("Failed to return defaultWasmBinaryName when WasmMainDir was an empty string. Returned %q", d.WasmBinaryName)
	}

}
