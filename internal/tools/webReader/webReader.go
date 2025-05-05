package webreader

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func ProcessUrl(u string) (string, error) {
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
