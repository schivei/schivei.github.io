package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	task := exec.Command("hugo", "serve", "--port", "65059")
	go func() {
		err := task.Run()
		if err != nil {
			panic(err)
		}
	}()
	defer func() {
		defer func() { _ = recover() }()
		_ = task.Cancel()
	}()

	time.Sleep(1000)

	// Create an exec allocator with Chrome flags to disable sandbox for CI environments
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	fmt.Print("generating './static/resume.pdf'... ")

	var resume []byte
	if err := chromedp.Run(ctx, printToPDF(`http://localhost:65059/`, &resume)); err != nil {
		fmt.Println()
		log.Fatal(err)
	}

	if err := os.WriteFile("./static/resume.pdf", resume, 0o644); err != nil {
		fmt.Println()
		log.Fatal(err)
	}

	fmt.Println("Done")
	fmt.Print("generating './static/curriculo.pdf'... ")

	var curriculo []byte
	if err := chromedp.Run(ctx, printToPDF(`http://localhost:65059/pt-br/`, &curriculo)); err != nil {
		fmt.Println()
		log.Fatal(err)
	}

	if err := os.WriteFile("./static/curriculo.pdf", curriculo, 0o644); err != nil {
		fmt.Println()
		log.Fatal(err)
	}

	fmt.Println("Done")
}

func printToPDF(urlStr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlStr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginBottom(1).
				WithMarginLeft(1).
				WithMarginRight(1).
				WithMarginTop(1).
				Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
