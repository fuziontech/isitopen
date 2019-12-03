package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type condPattern string
type roadCondition string

const (
	updateInterval = 60 * time.Second

	openPattern   condPattern = "NO TRAFFIC RESTRICTIONS ARE REPORTED FOR THIS AREA."
	chainsPattern condPattern = "CHAINS ARE REQUIRED "
	closedPattern condPattern = "CLOSED"

	open      roadCondition = "OPEN"
	chainsReq roadCondition = "CHAINS"
	closed    roadCondition = "CLOSED"
)

// HighwayStatus contains the status for the highway in memory
type HighwayStatus struct {
	Name      string        `json:"name"`
	Status    roadCondition `json:"status"`
	UpdatedAt string        `json:"updatedAt"`
}

// StatusStore This is the store for statuses for now
type StatusStore struct {
	Store map[string]HighwayStatus
}

func main() {
	// setup in memory store for status
	ss := StatusStore{
		Store: make(map[string]HighwayStatus),
	}

	// boot up and grab the deets on main two roads
	getCalTransStatus(&ss, "50")
	getCalTransStatus(&ss, "80")

	// setup the ticker for scraping caltrans website
	ticker := time.NewTicker(updateInterval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				getCalTransStatus(&ss, "50")
				getCalTransStatus(&ss, "80")
			}
		}
	}()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		// handle multiple domains
		var road string
		uri, ok := c.Get("location")
		if !ok {
			c.JSON(500, gin.H{
				"reason": "Location unknown",
			})
		}
		fiftyOpen := "is50open.com"
		eightyOpen := "is80open.com"
		switch uri.(*url.URL).Host {
		case fiftyOpen:
			road = "50"
		case eightyOpen:
			road = "80"
		default:
			road = "50"
		}
		// c.HTML(http.StatusOK, "index.html", gin.H{
		// 	"title": "Main website",
		// })
		roadStatus := ss.Store[road]
		c.JSON(http.StatusOK, roadStatus)
	})

	r.GET("/:road", func(c *gin.Context) {
		road := c.Param("road")
		roadStatus := ss.Store[road]
		c.JSON(http.StatusOK, roadStatus)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getCalTransStatus(store *StatusStore, road string) {
	resp := scrapeCalTrans(road)
	updatedAt := resp[0]
	strCond := getRoadCondition(resp[2])
	hs := HighwayStatus{
		Name:      road,
		Status:    strCond,
		UpdatedAt: updatedAt,
	}
	store.Store[road] = hs
}

func scrapeCalTrans(road string) []string {
	// get the data from caltrans
	formData := url.Values{
		"roadnumber": {road},
	}
	resp, err := http.PostForm("https://roads.dot.ca.gov/", formData)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the status element
	highwayStatus := doc.Find(".main-primary").Find("p").Text()
	lines := strings.Split(highwayStatus, "\n")
	return lines
}

func isOpen(stat roadCondition) bool {
	return stat == open
}

func getRoadCondition(resp string) roadCondition {
	switch {
	case strings.Contains(resp, string(openPattern)):
		return open
	case strings.Contains(resp, string(chainsPattern)):
		return chainsReq
	case strings.Contains(resp, string(closedPattern)):
		return closed
	default:
		return roadCondition("UNKNOWN")
	}
}
