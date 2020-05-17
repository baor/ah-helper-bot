package ahhelperbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"
)

// DeliveryProvider defines provider schedule
type DeliveryProvider interface {
	Get(postcode string) DeliverySchedule
}

// DefaultDeliveryProvider default implementation
type DefaultDeliveryProvider struct{}

// Get returns schedule for AH
func (p *DefaultDeliveryProvider) Get(postcode string) DeliverySchedule {
	log.Printf("Request deliveries for postcode '%s'", postcode)
	if len(postcode) == 0 {
		log.Fatal("Postcode is empty!")
	}

	c := http.Client{Timeout: 20 * time.Second}

	req := newDeliveryRequest(postcode)
	dumpReq, _ := httputil.DumpRequest(req, false)
	log.Printf("Request: %v\n", string(dumpReq))
	resp, err := c.Do(req)
	dump, _ := httputil.DumpResponse(resp, true)
	log.Printf("Response: %v\n", string(dump))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	dr := deliveryResponse{}

	err = json.NewDecoder(resp.Body).Decode(&dr)

	return convertResponseToSchedule(dr)
}

func newDeliveryRequest(postcode string) *http.Request {
	url := "https://www.ah.nl/service/rest/delegate?url=%2Fkies-moment%2Fbezorgen%2F" + postcode
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.162 Safari/537.36")
	req.Header.Add("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("referer", "https://www.ah.nl/mijnlijst")

	return req
}

// DeliveryTimeSlotBase base struct of delivery time slot. Same for time and date schedule
type DeliveryTimeSlotBase struct {
	Dl    int    `json:"dl"`
	From  string `json:"from"`
	To    string `json:"to"`
	State string `json:"state"`
}

type deliveryTimeSlot struct {
	DeliveryTimeSlotBase
	Date string `json:"-"`
}

func (s *deliveryTimeSlot) UnmarshalJSON(data []byte) error {

	var base DeliveryTimeSlotBase
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	s.Dl = base.Dl
	s.From = base.From
	s.To = base.To
	s.State = base.State

	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objmap); err != nil {
		return err
	}

	navItem, ok := objmap["navItem"]
	if !ok {
		return nil
	}
	if err := json.Unmarshal(navItem, &objmap); err != nil {
		return err
	}
	navLink, ok := objmap["link"]
	if !ok {
		return nil
	}
	if err := json.Unmarshal(navLink, &objmap); err != nil {
		return err
	}
	navHref, ok := objmap["href"]
	if !ok {
		return nil
	}

	var href string
	if err := json.Unmarshal(navHref, &href); err != nil {
		return err
	}

	r := regexp.MustCompile("/(\\d{4}-\\d{2}-\\d{2})/")
	date := r.FindStringSubmatch(href)[1]
	s.Date = date

	return nil
}

type deliveryDate struct {
	DeliveryTimeSlots []deliveryTimeSlot `json:"deliveryTimeSlots"`
	Date              string             `json:"date"`
}

type item struct {
	deliveryTimeSlots []deliveryTimeSlot
	deliveryDates     []deliveryDate
}

func (i *item) UnmarshalJSON(data []byte) error {
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objmap); err != nil {
		return err
	}

	itemTypeRaw, ok := objmap["type"]
	if !ok {
		return errors.New("no type")
	}
	i.deliveryTimeSlots = []deliveryTimeSlot{}
	i.deliveryDates = []deliveryDate{}

	var itemType string
	if err := json.Unmarshal(itemTypeRaw, &itemType); err != nil {
		return err
	}

	if itemType == "DeliveryTimeSelector" {
		emb, ok := objmap["_embedded"]
		if !ok {
			return errors.New("no _embedded")
		}
		if err := json.Unmarshal(emb, &objmap); err != nil {
			return err
		}

		slots, ok := objmap["deliveryTimeSlots"]
		if !ok {
			return errors.New("no deliveryTimeSlots")
		}

		if err := json.Unmarshal(slots, &i.deliveryTimeSlots); err != nil {
			return err
		}
		return nil
	}

	if itemType == "DeliveryDateSelector" {
		emb, ok := objmap["_embedded"]
		if !ok {
			return errors.New("no _embedded")
		}
		if err := json.Unmarshal(emb, &objmap); err != nil {
			return err
		}

		dates, ok := objmap["deliveryDates"]
		if !ok {
			return errors.New("no deliveryDates")
		}

		if err := json.Unmarshal(dates, &i.deliveryDates); err != nil {
			return err
		}

		return nil
	}

	return nil
}

type deliveryLane struct {
	items []item
}

func (d *deliveryLane) UnmarshalJSON(data []byte) error {
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objmap); err != nil {
		return err
	}

	emb, ok := objmap["_embedded"]
	if !ok {
		return errors.New("no embedded")
	}

	if err := json.Unmarshal(emb, &objmap); err != nil {
		return err
	}

	rawItems, ok := objmap["items"]
	if !ok {
		return errors.New("no items")
	}

	d.items = []item{}
	if err := json.Unmarshal(rawItems, &d.items); err != nil {
		return err
	}

	return nil
}

type deliveryResponse struct {
	lanes []deliveryLane
}

func (d *deliveryResponse) UnmarshalJSON(data []byte) error {
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objmap); err != nil {
		return err
	}

	emb, ok := objmap["_embedded"]
	if !ok {
		return errors.New("no embedded")
	}

	if err := json.Unmarshal(emb, &objmap); err != nil {
		return err
	}

	rawLanes, ok := objmap["lanes"]
	if !ok {
		return errors.New("no lanes")
	}

	d.lanes = []deliveryLane{}
	if err := json.Unmarshal(rawLanes, &d.lanes); err != nil {
		return err
	}

	return nil
}

//DeliverySchedule is used to represent time schedule by date
type DeliverySchedule map[string][]DeliveryTimeSlotBase

//convertResponseToSchedule converts response to delivery schedule
func convertResponseToSchedule(dr deliveryResponse) DeliverySchedule {
	ds := DeliverySchedule{}
	for _, line := range dr.lanes {
		for _, item := range line.items {
			for _, dd := range item.deliveryDates {
				slots := []DeliveryTimeSlotBase{}
				for _, dts := range dd.DeliveryTimeSlots {
					if dts.State == "full" {
						continue
					}
					slots = append(slots, DeliveryTimeSlotBase{
						Dl:   dts.Dl,
						From: dts.From,
						To:   dts.To,
					})

				}
				if len(slots) > 0 {
					if ds[dd.Date] == nil {
						ds[dd.Date] = []DeliveryTimeSlotBase{}
					}
					ds[dd.Date] = append(ds[dd.Date], slots...)
				}
			}
			for _, dts := range item.deliveryTimeSlots {
				if dts.State == "full" {
					continue
				}

				if ds[dts.Date] == nil {
					ds[dts.Date] = []DeliveryTimeSlotBase{}
				}
				ds[dts.Date] = append(ds[dts.Date], DeliveryTimeSlotBase{
					Dl:   dts.Dl,
					From: dts.From,
					To:   dts.To,
				})

			}
		}
	}
	return ds
}

func (ds DeliverySchedule) String() string {
	var stringBuilder strings.Builder
	for date, scheds := range ds {
		stringBuilder.WriteString(fmt.Sprintf("*%s*: ", date))
		for _, sched := range scheds {
			stringBuilder.WriteString(fmt.Sprintf("%s-%s ", sched.From, sched.To))
		}
		stringBuilder.WriteString("\n")
	}
	return stringBuilder.String()
}
