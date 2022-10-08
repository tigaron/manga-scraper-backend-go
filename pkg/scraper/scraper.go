package scraper

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type SeriesKey struct {
	WebtoonProvider string
	SeriesId        string
}

type SeriesUpdate struct {
	SeriesCover    string
	SeriesShortUrl string
	SeriesSynopsis string
	ScrapeDate     string
}

type SeriesNew struct {
	WebtoonProvider string
	SeriesId        string
	SeriesTitle     string
	SeriesUrl       string
	ScrapeDate      string
}

type ChapterKey struct {
	SeriesProvider string
	ChapterId      string
}

type ChapterUpdate struct {
	ChapterTitle    string
	ChapterShortUrl string
	ChapterPrev     string
	ChapterNext     string
	ChapterContent  []string
	ScrapeDate      string
}

type ChapterNew struct {
	SeriesProvider    string
	ChapterId         string
	ChapterShortTitle string
	ChapterDate       string
	ChapterUrl        string
	ChapterOrder      int
	ScrapeDate        string
}

func ScrapeSeriesList(provider *string, sourceUrl *string, tableName string) (*[]SeriesNew, error) {
	var result = new([]SeriesNew)
	var scrapeError error

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
		urlArr := strings.Split(h.Attr("href"), "/")
		re := regexp.MustCompile(`^\d+-?`)
		item := SeriesNew{
			WebtoonProvider: *provider,
			SeriesId:        re.ReplaceAllString(urlArr[len(urlArr)-2], ""),
			SeriesTitle:     h.Text,
			SeriesUrl:       h.Attr("href"),
			ScrapeDate:      time.Now().Format(time.RFC3339),
		}
		*result = append(*result, item)
	})

	collector.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping '%v'.", r.Request.URL)
	})

	collector.Visit(*sourceUrl)

	return result, scrapeError
}

func ScrapeSeriesData(provider *string, sourceUrl *string, tableName string) (*SeriesKey, *SeriesUpdate, error) {
	var key SeriesKey
	var data SeriesUpdate
	var scrapeError error

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

	collector.OnHTML("html", func(h *colly.HTMLElement) {
		urlArr := strings.Split(*sourceUrl, "/")
		re := regexp.MustCompile(`^\d+-?`)
		key = SeriesKey{
			WebtoonProvider: *provider,
			SeriesId:        re.ReplaceAllString(urlArr[len(urlArr)-2], ""),
		}
		data = SeriesUpdate{
			SeriesCover:    h.ChildAttr("div.thumb img", "src"),
			SeriesShortUrl: h.ChildAttr("link[rel='shortlink']", "href"),
			SeriesSynopsis: h.ChildText("div.entry-content"), // TODO do something about the \n or div
			ScrapeDate:     time.Now().Format(time.RFC3339),
		}
	})

	collector.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping '%v'.", r.Request.URL)
	})

	collector.Visit(*sourceUrl)

	return &key, &data, scrapeError
}

func ScrapeChaptersList(provider *string, sourceUrl *string, tableName string) error {
	return nil
}

func ScrapeChaptersData(provider *string, sourceUrl *string, tableName string) error {
	return nil
}
