// +build ignore

package main 

import (
	"context"
	"log"
	"time"
	"os"
	
	"github.com/chromedp/chromedp"
)

func main() {

	// remove generated files after running
	defer os.Remove("root.go")
	defer os.Remove("main_wasm.go")
	defer os.Remove("go.sum")

	ctx, cancel := chromedp.NewContext(context.Background(),chromedp.WithLogf(log.Printf))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate("http://127.0.0.1:19944/"),
		chromedp.WaitVisible("#run"),
		chromedp.Click("#run"),
		chromedp.WaitVisible("#success"),
	)
	if err != nil {
		log.Fatal(err)
	}

}
