package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libagent/pkg/config"
	"net/http"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/tmc/langchaingo/llms"
)

var WebReaderDefinition = llms.FunctionDefinition{
	Name: "webReader",
	Description: `Uses provided valid URL and provides a markdown text converted from html for ease of read.
		Please be sure to put a valid URL here, you can use LLM tool to extract it from query before using it in this tool.`,
	Parameters: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The valid url to read as text from.",
			},
		},
	},
}

type WebReaderArgs struct {
	URL string `json:"url"`
}

type WebReaderTool struct {
}

func (s WebReaderTool) Call(ctx context.Context, input string) (string, error) {
	webReaderArgs := WebReaderArgs{}
	if err := json.Unmarshal([]byte(input), &webReaderArgs); err != nil {
		return "", err
	}

	return processUrl(webReaderArgs.URL)
}

func init() {
	globalToolsRegistry = append(globalToolsRegistry,
		func(ctx context.Context, cfg config.Config) (*ToolData, error) {
			if cfg.WebReaderDisable {
				return nil, nil
			}
			webReaderTool := WebReaderTool{}

			return &ToolData{
				Definition: WebReaderDefinition,
				Call:       webReaderTool.Call,
			}, nil
		},
	)
}

func processUrl(u string) (string, error) {
	requiresJS, err := checkNoScript(u)
	if err != nil {
		return "", err
	}

	var body io.Reader

	if requiresJS {
		html, err := loadJSPage(u)
		if err != nil {
			return "", err
		}
		body = strings.NewReader(html)
	} else {
		resp, err := http.Get(u)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		body = resp.Body
	}

	md, err := htmltomarkdown.ConvertReader(
		body,
		converter.WithDomain(u),
	)
	return string(md), err
}

func loadJSPage(u string) (string, error) {
	browser := rod.New().MustConnect()
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{
		URL: u,
	})
	if err != nil {
		return "", err
	}
	defer page.Close()

	err = page.WaitLoad()
	if err != nil {
		return "", err
	}
	err = page.WaitDOMStable(5*time.Second, 0.75)
	if err != nil {
		return "", err
	}

	return page.HTML()
}

func checkNoScript(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, err
	}

	requiresJS := false
	doc.Find("noscript").Each(func(i int, s *goquery.Selection) {
		text := strings.ToLower(s.Text())
		if strings.Contains(text, "javascript") {
			requiresJS = true
		}
	})

	return requiresJS, nil
}
