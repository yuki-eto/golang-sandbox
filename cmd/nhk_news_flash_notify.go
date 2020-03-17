package main

import (
	"encoding/xml"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/martinlindhe/notify"
)

type NHKNewsFlash struct {
	XMLName xml.Name `xml:"flashNews"`
	Flag    string   `xml:"flag,attr"`
	PubDate string   `xml:"pubDate,attr"`
	Report  []struct {
		Category string `xml:"category,attr"`
		Date     string `xml:"date,attr"`
		Link     string `xml:"link,attr"`
		Line     string `xml:"line"`
	} `xml:"report"`
}

var (
	lastPubDate *time.Time
)

func main() {
	if err := parseXML(); err != nil {
		panic(err)
	}

	ticker := time.NewTicker(10 * time.Second)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			if err := parseXML(); err != nil {
				panic(err)
			}
		case <-sig:
			log.Print("interrupted")
			return
		}
	}
}

func parseXML() error {
	resp, err := http.Get("https://www3.nhk.or.jp/sokuho/news/sokuho_news.xml")
	if err != nil {
		return err
	}
	reader := resp.Body
	defer reader.Close()

	dec := xml.NewDecoder(reader)
	var news NHKNewsFlash
	if err := dec.Decode(&news); err != nil {
		return err
	}

	pubDate, err := time.Parse(time.RFC1123Z, news.PubDate)
	if err != nil {
		return err
	}
	if lastPubDate == nil || !lastPubDate.Equal(pubDate) {
		lastPubDate = &pubDate
	} else {
		log.Print("not updated")
		return nil
	}

	for _, r := range news.Report {
		notify.Alert("NHKNewsFlash", "", r.Line, "")
	}
	return nil
}
