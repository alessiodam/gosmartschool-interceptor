package interceptor

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"log"
	"os"
	"strings"
	"time"
)

func StartChromeAndCapture(smartschoolDomain string) error {
	logFile := fmt.Sprintf("./gsscap-requests/%s.gsscap", uuid.New().String())

	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing log file: %v", err)
		}
	}()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("start-maximized", true),
		chromedp.Flag("default-search-engine", "google"),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	if err := chromedp.Run(
		ctx,
		network.Enable(),
		chromedp.Navigate("https://"+smartschoolDomain),
	); err != nil {
		return fmt.Errorf("failed to navigate: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Context has ended. Stopping Chrome.")
			return nil
		default:
			var id network.RequestID
			done := make(chan bool)
			chromedp.ListenTarget(ctx, func(ev interface{}) {
				switch ev := ev.(type) {
				case *network.EventRequestWillBeSent:
					go logRequest(ctx, file, ev)
				case *network.EventResponseReceived:
					id = ev.RequestID
				case *network.EventLoadingFinished:
					if id != "" && ev.RequestID == id {
						go func() {
							defer func() {
								done <- true
							}()

							if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
								var err error
								return err
							})); err != nil {
								log.Printf("Error getting response body: %v", err)
								return
							}

							logResponse(ctx, file, ev.RequestID)
						}()
					}
				}
			})
			<-done
		}
	}
}
func logRequest(_ context.Context, file *os.File, event *network.EventRequestWillBeSent) {
	timestamp := time.Now().Format(time.RFC3339)
	req := event.Request

	postData := ""
	if req.HasPostData && len(req.PostDataEntries) > 0 {
		var sb strings.Builder
		for _, entry := range req.PostDataEntries {
			sb.WriteString(entry.Bytes)
		}
		postData = sb.String()
	}

	logEntry := fmt.Sprintf("REQUEST [%s]\n", timestamp)
	logEntry += fmt.Sprintf("URL: %s\nMethod: %s\nHeaders: %v\nPostData: %s\nReferrer Policy: %s\nRequest ID: %s\n\n",
		req.URL, req.Method, req.Headers, postData, req.ReferrerPolicy, event.RequestID)

	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("Error logging request: %v", err)
	}
}

func logResponse(ctx context.Context, file *os.File, id network.RequestID) {
	timestamp := time.Now().Format(time.RFC3339)

	var data []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		data, err = network.GetResponseBody(id).Do(ctx)
		return err
	})); err != nil {
		log.Printf("Error getting response body: %v", err)
		return
	}

	logEntry := fmt.Sprintf("RESPONSE [%s]\n", timestamp)
	logEntry += fmt.Sprintf("Body: %s\nRequest ID: %s\n\n",
		string(data), id)

	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("Error logging response: %v", err)
	}
}
