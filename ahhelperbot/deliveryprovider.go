package ahhelperbot

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type DeliveryProvider interface {
	Get(postcode string) string
}

type DefaultDeliveryProvider struct{}

func (p *DefaultDeliveryProvider) Get(postcode string) string {

	c := http.Client{Timeout: 20 * time.Second}

	req := newDeliveryRequest(postcode)
	resp, err := c.Do(req)
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%q", dump)

	return ""
}

func newDeliveryRequest(postcode string) *http.Request {
	url := "https://www.ah.nl/service/rest/delegate?url=%2Fkies-moment%2Fbezorgen%2F" + postcode
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("authority", "www.ah.nl")
	req.Header.Add("accept", "application/json")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.162 Safari/537.36")
	req.Header.Add("x-order-mode", "false")
	req.Header.Add("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("x-breakpoint", "medium")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("referer", "https://www.ah.nl/mijnlijst")
	req.Header.Add("accept-language", "en-US,en;q=0.9,nl-NL;q=0.8,nl;q=0.7,ru-RU;q=0.6,ru;q=0.5")

	return req
}
