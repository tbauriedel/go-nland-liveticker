package scraper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly"

	"github.com/tbauriedel/go-nland-liveticker/internal/model"
)

type Scraper struct {
	Collector       *colly.Collector
	Err             error
	ReturnCode      int
	FoundOperations []model.Operation
}

func NewScraper() *Scraper {
	s := &Scraper{}

	s.Collector = colly.NewCollector(
		colly.AllowedDomains("www.kfv-online.de"),
	)

	s.Collector.OnError(func(r *colly.Response, err error) {
		s.Err = fmt.Errorf("collector failed", "error", err, "url", r.Request.URL)
		s.ReturnCode = r.StatusCode
	})

	s.Collector.OnResponse(func(r *colly.Response) {
		s.ReturnCode = r.StatusCode
	})

	return s
}

func (s *Scraper) Register() {
	s.Collector.OnHTML("table#operationList > tbody", func(HtmlElement *colly.HTMLElement) {
		HtmlElement.ForEach("tr", func(_ int, element *colly.HTMLElement) {

			t, _ := time.Parse("02.01.2006 15:04", element.ChildText("td:nth-child(1)"))

			row := model.Operation{
				Time:     t,
				Units:    streamlineUnits(element.ChildText("td:nth-child(2)")),
				District: element.ChildText("td:nth-child(3)"),
				Report:   element.ChildText("td:nth-child(4)"),
				Location: element.ChildText("td:nth-child(5)"),
			}

			s.FoundOperations = append(s.FoundOperations, row)
		})
	})
}

func streamlineUnits(units string) string {
	var specialUnits = []string{
		"ILS Lagedienst", "Kreisbrandinspektion", "THW", "UG-ÖEL",
	}

	var (
		foundUnits        []string
		foundSpecialUnits []string
	)

	for _, specialUnit := range specialUnits { // Search for special units in original string
		if strings.Contains(units, specialUnit) {
			foundSpecialUnits = append(foundSpecialUnits, specialUnit) // Add to list
			units = strings.ReplaceAll(units, specialUnit, "")         // Remove from original string
		}
	}

	// Split remaining units by "FF "
	splittedUnits := strings.Split(units, "FF ")
	for cnt, unit := range splittedUnits {
		if cnt != 0 { // First element is always empty due to split
			foundUnits = append(foundUnits, fmt.Sprintf("FF %s", unit))
		}
	}

	// Combine both lists and return
	foundUnits = append(foundUnits, foundSpecialUnits...)

	return strings.Join(foundUnits, ", ")
}

// ScrapeOperations Scrapes the latest operation from the website and returns it
func (s *Scraper) ScrapeOperations() (model.Operation, error) {
	o := model.Operation{}

	err := s.Collector.Visit("https://www.kfv-online.de/home/einsaetze")
	if err != nil || s.ReturnCode != http.StatusOK {
		return o, fmt.Errorf("error visiting page: %w", err)
	}

	// Get the latest operation
	o = s.FoundOperations[0]

	return o, nil
}
