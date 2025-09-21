package webreader

import (
	"net/http"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
)

func ProcessUrl(u string) (string, error) {
	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	md, err := htmltomarkdown.ConvertReader(
		resp.Body,
		converter.WithDomain(u),
	)
	return string(md), err
}
