package chromedp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustChromeCtx() (context.Context, context.CancelFunc) {

	debugURL := func() string {
		resp, err := http.Get("http://localhost:9222/json/version")
		if err != nil {
			panic(err)
		}

		var result map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			panic(err)
		}
		return result["webSocketDebuggerUrl"].(string)
	}()

	// t.Log(debugURL)

	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), debugURL)
	// defer cancel()

	ctx, _ := chromedp.NewContext(allocCtx) // , chromedp.WithLogf(log.Printf))
	// defer cancel()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	return ctx, cancel
}

// WaitInnerTextTrimEq will wait for the innerText of the specified element to match a specific string after whitespace trimming.
func WaitInnerTextTrimEq(sel, innerText string) chromedp.QueryAction {

	return chromedp.Query(sel, func(s *chromedp.Selector) {

		chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame, id runtime.ExecutionContextID, ids ...cdp.NodeID) ([]*cdp.Node, error) {

			nodes := make([]*cdp.Node, len(ids))
			cur.RLock()
			for i, id := range ids {
				nodes[i] = cur.Nodes[id]
				if nodes[i] == nil {
					cur.RUnlock()
					// not yet ready
					return nil, nil
				}
			}
			cur.RUnlock()

			var ret string
			err := chromedp.EvaluateAsDevTools("document.querySelector('"+sel+"').innerText", &ret).Do(ctx)
			if err != nil {
				return nodes, err
			}
			if strings.TrimSpace(ret) != innerText {
				// log.Printf("found text: %s", ret)
				return nodes, errors.New("unexpected value: " + ret)
			}

			// log.Printf("NodeValue: %#v", nodes[0])

			// return nil, errors.New("not ready yet")
			return nodes, nil
		})(s)

	})

}

func QueryNode(ref string, assert func(n *cdp.Node)) chromedp.QueryAction {
	return chromedp.QueryAfter(ref, func(ctx context.Context, id runtime.ExecutionContextID, nodes ...*cdp.Node) error {
		if len(nodes) == 0 {
			return fmt.Errorf("no %s element found", ref)
		}
		assert(nodes[0])
		return nil
	})
}
