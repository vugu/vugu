package vugu

// func TestParse2(t *testing.T) {

// 	tmpDir, err := ioutil.TempDir("", "TestParse2")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(tmpDir)

// 	p := &ParserGo{
// 		PackageName: "main",
// 		StructType:  "Root",
// 		// TagName:       "demo-comp",
// 		// DataType: "DemoCompData",
// 		OutDir:  tmpDir,
// 		OutFile: "root.go",
// 	}

// 	err = p.Parse(bytes.NewReader([]byte(`

// <html>
// 	<head>
// 		<title>This is the title</title>
// 	</head>
// 	<body>
// 		<div id="testdiv">
// 			This is a test.

// 			<ul>
// 				<li vg-for='_, i := range someNums'>Testing <span vg-html="i"></span></li>
// 			</ul>

// 		</div>
// 		<script>
// 		alert("HELLO!");
// 		</script>
// 	</body>
// </html>

// <style>
// body { background: green; }
// </style>

// <script type="application/x-go">
// func init() {}

// var someNums = []int{5,6,7,8}

// </script>

// `)), "root.vugu")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "root.go"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	log.Printf("OUTPUT: \n%s", b)

// }
