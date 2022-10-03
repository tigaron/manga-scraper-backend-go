package scraper

import (
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gocolly/colly"
)

type Series struct {
	WebtoonProvider string
	SeriesId        string
	SeriesTitle     string
	SeriesCover     string
	SeriesUrl       string
	SeriesShortUrl  string
	SeriesSynopsis  string
	ScrapeDate      string
}

type SeriesList struct {
	WebtoonProvider string
	SeriesId        string
	SeriesTitle     string
	SeriesUrl       string
	ScrapeDate      string
}

type Chapter struct {
	SeriesProvider  string
	ChapterId       string
	ChapterTitle    string
	ChapterNumber   string
	ChapterDate     string
	ChapterUrl      string
	ChapterShortUrl string
	ChapterOrder    int
	ChapterPrev     string
	ChapterNext     string
	ChapterContent  []string
	ScrapeDate      string
}

func ScrapeSeriesList(message map[string]events.SQSMessageAttribute, tableName string) (*[]SeriesList, error) {
	var scrapeError error
	result := new([]SeriesList)

	collector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	collector.OnRequest(func(r *colly.Request) {
		log.Printf("Starting to scrape '%v'.", r.URL)
	})

	collector.OnResponse(func(r *colly.Response) {
		log.Printf("Got response code of '%v'.", r.StatusCode)
	})

	collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Couldn't fetch '%v'. Status code: '%v'. Error: %v\n", r.StatusCode, r.Request.URL, err)
		scrapeError = err
	})

	collector.OnHTML("div.soralist a.series", func(h *colly.HTMLElement) {
		item := SeriesList{
			WebtoonProvider: *message["Provider"].StringValue,
			SeriesId:        h.Attr("href"),
			SeriesTitle:     h.Text,
			SeriesUrl:       h.Attr("href"),
			ScrapeDate:      time.Now().Format(time.RFC3339),
		}
		*result = append(*result, item)
	})

	collector.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping '%v'.", r.Request.URL)
	})

	collector.Visit(*message["SourceUrl"].StringValue)

	return result, scrapeError
}

func ScrapeSeriesData(message map[string]events.SQSMessageAttribute, tableName string) error {
	return nil
}

func ScrapeChaptersList(message map[string]events.SQSMessageAttribute, tableName string) error {
	return nil
}

func ScrapeChaptersData(message map[string]events.SQSMessageAttribute, tableName string) error {
	return nil
}
