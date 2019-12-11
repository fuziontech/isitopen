package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

type condPattern string
type roadCondition string

const (
	updateInterval = 60 * time.Second

	openPattern         condPattern = "NO TRAFFIC RESTRICTIONS ARE REPORTED FOR THIS AREA."
	chainsPattern       condPattern = "CHAINS ARE REQUIRED "
	advisoryPattern     condPattern = "ADVISORY"
	closedPattern       condPattern = "CLOSED"
	constructionPattern condPattern = "CONSTRUCTION"

	open         roadCondition = "OPEN"
	chainsReq    roadCondition = "CHAINS"
	closed       roadCondition = "CLOSED"
	advisory     roadCondition = "ADVISORY"
	construction roadCondition = "CONSTRUCTION"
)

// HighwayStatus contains the status for the highway in memory
type HighwayStatus struct {
	Name        string        `json:"name"`
	Status      roadCondition `json:"status"`
	Description string        `json:"description"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

// StatusStore This is the store for statuses for now
type StatusStore struct {
	Store map[string]HighwayStatus
}

func scrape(ss *StatusStore) {
	getCalTransStatus(ss, "50")
	getCalTransStatus(ss, "80")
	getCalTransStatus(ss, "88")
}

func main() {
	// setup in memory store for status
	ss := StatusStore{
		Store: make(map[string]HighwayStatus),
	}

	// boot up and grab the deets on main two roads
	scrape(&ss)

	// setup the ticker for scraping caltrans website
	ticker := time.NewTicker(updateInterval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				scrape(&ss)
			}
		}
	}()

	r := gin.Default()
	// configure to automatically detect scheme and host
	// - use http when default scheme cannot be determined
	// - use localhost:8080 when default host cannot be determined
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		// handle multiple domains
		var road string
		url := c.Request.Host

		// test to see where this request is coming from
		fiftyOpen := "is50open.com"
		eightyOpen := "is80open.com"
		eighteight := "is88open.com"
		switch url {
		case fiftyOpen:
			road = "50"
		case eightyOpen:
			road = "80"
		case eighteight:
			road = "88"
		default:
			road = "50"
		}
		roadStatus := ss.Store[road]
		c.HTML(http.StatusOK, "index.html", gin.H{
			"status": roadStatus,
		})
	})

	r.GET("/v1/road/:road", func(c *gin.Context) {
		road := c.Param("road")
		roadStatus := ss.Store[road]
		c.JSON(http.StatusOK, roadStatus)
	})

	r.Static("/static", "./static")

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// for parsing the calendar date
// map[ordinal]cardinal
var ordinals = map[string]string{
	"1st": "1", "2nd": "2", "3rd": "3", "4th": "4", "5th": "5",
	"6th": "6", "7th": "7", "8th": "8", "9th": "9", "10th": "10",
	"11th": "11", "12th": "12", "13th": "13", "14th": "14", "15th": "15",
	"16th": "16", "17th": "17", "18th": "18", "19th": "19", "20th": "20",
	"21st": "21", "22nd": "22", "23rd": "23", "24th": "24", "25th": "25",
	"26th": "26", "27th": "27", "28th": "28", "29th": "29", "30th": "30",
	"31st": "31",
}

func getCalTransStatus(store *StatusStore, road string) {
	resp := scrapeCalTrans(road)
	longResponse := strings.Join(resp[2:], "\n")
	strCond := getRoadCondition(longResponse)

	// parse updatedAt from string from DOT
	ua := strings.Split(resp[0], ",")
	layout := "January 2, 2006 at 03:04 PM"
	uas := ua[1] + "," + ua[2]
	uas = strings.TrimSpace(uas)
	uas = strings.Trim(uas, ".")

	for k, v := range ordinals {
		uas = strings.Replace(uas, k, v, 1)
	}

	updatedAt, err := time.Parse(layout, uas)
	if err != nil {
		log.Fatalln(err)
	}

	hs := HighwayStatus{
		Name:        road,
		Status:      strCond,
		Description: longResponse,
		UpdatedAt:   updatedAt,
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
	case strings.Contains(resp, string(advisoryPattern)):
		return advisory
	case strings.Contains(resp, string(constructionPattern)):
		return construction
	case strings.Contains(resp, string(closedPattern)):
		return closed
	default:
		return roadCondition("UNKNOWN")
	}
}
