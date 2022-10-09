package scraper

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
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

type TSReaderScript struct {
	PrevURL string `json:"prevUrl"`
	NextURL string `json:"nextUrl"`
	Sources []struct {
		Images []string `json:"images"`
	} `json:"sources"`
}

func ScrapeSeriesList(provider *string, sourceUrl *string) ([]SeriesNew, error) {
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
		log.Fatalf("Couldn't fetch '%v'. Status code: '%v'. Error: %v\n", r.StatusCode, r.Request.URL, err.Error())
		scrapeError = err
	})

	collector.OnHTML("div.soralist a.series", func(h *colly.HTMLElement) {
		urlArr := strings.Split(h.Attr("href"), "/")
		reUrl := regexp.MustCompile(`^\d+-?`)
		item := SeriesNew{
			WebtoonProvider: *provider,
			SeriesId:        reUrl.ReplaceAllString(urlArr[len(urlArr)-2], ""),
			SeriesTitle:     strings.TrimSpace(h.Text),
			SeriesUrl:       h.Attr("href"),
			ScrapeDate:      time.Now().Format(time.RFC3339),
		}
		*result = append(*result, item)
	})

	collector.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping '%v'.", r.Request.URL)
	})

	collector.Visit(*sourceUrl)

	return *result, scrapeError
}

func ScrapeSeriesData(provider *string, sourceUrl *string) (*SeriesKey, *SeriesUpdate, error) {
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
		log.Fatalf("Couldn't fetch '%v'. Status code: '%v'. Error: %v\n", r.StatusCode, r.Request.URL, err.Error())
		scrapeError = err
	})

	collector.OnHTML("html", func(h *colly.HTMLElement) {
		urlArr := strings.Split(*sourceUrl, "/")
		reUrl := regexp.MustCompile(`^\d+-?`)
		key = SeriesKey{
			WebtoonProvider: *provider,
			SeriesId:        reUrl.ReplaceAllString(urlArr[len(urlArr)-2], ""),
		}
		reSynopsis := regexp.MustCompile(`\n`)
		data = SeriesUpdate{
			SeriesCover:    h.ChildAttr("div.thumb img", "src"),
			SeriesShortUrl: h.ChildAttr("link[rel='shortlink']", "href"),
			SeriesSynopsis: reSynopsis.ReplaceAllString(strings.TrimSpace(h.ChildText("div.entry-content")), "<br />"),
			ScrapeDate:     time.Now().Format(time.RFC3339),
		}
	})

	collector.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping '%v'.", r.Request.URL)
	})

	collector.Visit(*sourceUrl)

	return &key, &data, scrapeError
}

func ScrapeChapterList(provider *string, sourceUrl *string) ([]ChapterNew, error) {
	var result = new([]ChapterNew)
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
		log.Fatalf("Couldn't fetch '%v'. Status code: '%v'. Error: %v\n", r.StatusCode, r.Request.URL, err.Error())
		scrapeError = err
	})

	collector.OnHTML("div.eplister a", func(h *colly.HTMLElement) {
		urlArr := strings.Split(h.Attr("href"), "/")
		reUrl := regexp.MustCompile(`^\d+-?`)
		reTitle := regexp.MustCompile(`\n`)
		reOrder := regexp.MustCompile(`\d+`)
		orderRaw, _ := h.DOM.ParentsFiltered("li").Attr("data-num")
		orderValue, _ := strconv.Atoi(reOrder.FindString(orderRaw))
		item := ChapterNew{
			SeriesProvider:    *provider,
			ChapterId:         reUrl.ReplaceAllString(urlArr[len(urlArr)-2], ""),
			ChapterShortTitle: reTitle.ReplaceAllString(strings.TrimSpace(h.ChildText("span.chapternum")), " "),
			ChapterDate:       strings.TrimSpace(h.ChildText("span.chapterdate")),
			ChapterUrl:        h.Attr("href"),
			ChapterOrder:      orderValue,
			ScrapeDate:        time.Now().Format(time.RFC3339),
		}
		*result = append(*result, item)
	})

	collector.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping '%v'.", r.Request.URL)
	})

	collector.Visit(*sourceUrl)

	return *result, scrapeError
}

func ScrapeChapterData(provider *string, sourceUrl *string) (*ChapterKey, *ChapterUpdate, error) {
	var key ChapterKey
	var data ChapterUpdate
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
		log.Fatalf("Couldn't fetch '%v'. Status code: '%v'. Error: %v\n", r.StatusCode, r.Request.URL, err.Error())
		scrapeError = err
	})

	collector.OnHTML("html", func(h *colly.HTMLElement) {
		urlArr := strings.Split(*sourceUrl, "/")
		reUrl := regexp.MustCompile(`^\d+-?`)
		key = ChapterKey{
			SeriesProvider: *provider,
			ChapterId:      reUrl.ReplaceAllString(urlArr[len(urlArr)-2], ""),
		}
		tsReaderScript := h.ChildText("script:contains('ts_reader.run')")
		reScript := regexp.MustCompile(`^ts_reader.run\((.*)\);`)
		tsReaderRaw := reScript.FindStringSubmatch(tsReaderScript)
		var tsReaderValue TSReaderScript
		err := json.Unmarshal([]byte(tsReaderRaw[1]), &tsReaderValue)
		if err != nil {
			log.Fatalf("Couldn't unmarshal ts_reader script. Error: %v\n", err.Error())
		}
		prevUrl := strings.Split(tsReaderValue.PrevURL, "/")
		nextUrl := strings.Split(tsReaderValue.NextURL, "/")
		var prevUrlVal string
		var nextUrlVal string
		if len(prevUrl) > 1 {
			prevUrlVal = prevUrl[len(prevUrl)-2]
		}
		if len(nextUrl) > 1 {
			nextUrlVal = nextUrl[len(nextUrl)-2]
		}
		data = ChapterUpdate{
			ChapterTitle:    strings.TrimSpace(h.ChildText("h1.entry-title")),
			ChapterShortUrl: h.ChildAttr("link[rel='shortlink']", "href"),
			ChapterPrev:     reUrl.ReplaceAllString(prevUrlVal, ""),
			ChapterNext:     reUrl.ReplaceAllString(nextUrlVal, ""),
			ChapterContent:  tsReaderValue.Sources[0].Images,
			ScrapeDate:      time.Now().Format(time.RFC3339),
		}
	})

	collector.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping '%v'.", r.Request.URL)
	})

	collector.Visit(*sourceUrl)

	return &key, &data, scrapeError
}
