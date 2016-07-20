package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/redis.v2"

	"github.com/montanaflynn/stats"
)

var (
	redisHost string
	redisPort string
	mongoHost string
	mongoPort string

	consumers string

	mean   float64
	stdDev float64

	domain = "www.youtube.com"
)

type VideoStats struct {
	URL      string    `json:"url"`
	ZScore   float64   `json:"z_score"`
	Views    int64     `json:"views"`
	Analyzed time.Time `jsont:"analyzed"`
}

func main() {
	if err := loadEnvironmentVariables(); err != nil {
		log.Fatalln("Error loading environment variables:", err.Error())
	}

	// connect to redis db
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := client.Ping().Result(); err != nil {
		log.Fatalln("Unable to connect to Redis:", err.Error())
	}

	session, err := mgo.Dial(fmt.Sprintf("%s:%s", mongoHost, mongoPort))
	if err != nil {
		log.Fatalln("Unable to connect to MongoDB:", err.Error())
	}
	defer session.Close()

	// calculate the main, std dev for z-scores of videos
	routines, err := strconv.Atoi(consumers)
	if err != nil {
		log.Fatalln("Unable to parse consumers into int:", err.Error())
	}

	views := make([]float64, 0)
	videos := make([]*VideoStats, 0)

	var wg sync.WaitGroup
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(client *redis.Client, session *mgo.Session) {
			defer wg.Done()

			for url, err := client.RPop(domain).Result(); err != nil; {
				stat, err := calculateVideoStatsFromURL(url)
				if err != nil {
					log.Printf("Error getting video stats for %s: %s")
				} else if err = insertIntoMongo(stat, session); err != nil {
					log.Println("Error inserting video stats in to MongoDB:", err.Error())
				} else {
					views = append(views, float64(stat.Views))
					videos = append(videos, stat)
				}
			}
		}(client, session)
	}

	wg.Wait()

	mean, err = stats.Mean(views)
	if err != nil {
		log.Println("Error calculating mean from view data set:", err.Error())
	}

	stdDev, err = stats.StandardDeviation(views)
	if err != nil {
		log.Println("Error calculating std dev from view data set:", err.Error())
	}

	updateVideoStatsWithZScores(videos, session)
}

func calculateVideoStatsFromURL(url string) (*VideoStats, error) {
	// grab the response from the site...
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		tokenizer := html.NewTokenizer(resp.Body)

		depth := 0
		for {
			tt := tokenizer.Next()
			switch tt {
			case html.ErrorToken:
				// should only happen when we hit EOF
			case html.TextToken:
				if depth > 0 {
					fields := strings.Fields(string(tokenizer.Text()))
					for _, f := range fields {
						if f == "views" {
							log.Println("URL Views: %s", fields[0])
							strippedStr := strings.Replace(fields[0], ",", "", -1)

							val, err := strconv.ParseInt(strippedStr, 10, 64)
							if err != nil {
								log.Println("Error parsing string to int:", err.Error())
							} else if val > 10000 {
								return &VideoStats{
									URL:      url,
									Views:    val,
									ZScore:   0.0,
									Analyzed: time.Now(),
								}, nil
							} else {
								return nil, errors.New("Video views less than 10000")
							}
						}
					}
				}
			case html.StartTagToken, html.EndTagToken:
				tn, _ := tokenizer.TagName()
				if len(tn) == 3 && string(tn[:3]) == "div" {
					if tt == html.StartTagToken {
						depth++
					} else {
						depth--
					}
				}
			} // end switch
		} // end for
	} // end if

	return nil, errors.New("Response body was nil")
}

func updateVideoStatsWithZScores(videos []*VideoStats, session *mgo.Session) {
	c := session.DB("video_stats").C("stats")

	for _, video := range videos {
		// lookup the video url in mongo
		var result VideoStats
		err := c.Find(bson.M{"url": video.URL}).One(&result)
		if err != nil {
			log.Println("Error looking up video in mongo", err.Error())
		}

		score := getZScore(result.Views)

		video.ZScore = score

		err = c.Update(bson.M{"url": video.URL}, video)
		if err != nil {
			log.Println("Unable to update video stat record", err.Error())
		}
	}
}

func insertIntoMongo(stat *VideoStats, session *mgo.Session) error {
	c := session.DB("video_stats").C("stats")
	if err := c.Insert(stat); err != nil {
		return err
	}

	return nil
}

func getZScore(views int64) float64 {
	return (float64(views) - mean) / stdDev
}

func loadEnvironmentVariables() error {
	redisHost = os.Getenv("REDIS_PORT_6379_TCP_ADDR")
	redisPort = os.Getenv("REDIS_PORT_6379_TCP_PORT")

	mongoHost = os.Getenv("MONGODB_PORT_27017_TCP_ADDR")
	mongoPort = os.Getenv("MONGODB_PORT_27017_TCP_PORT")

	consumers = os.Getenv("GO_ROUTINE_REDIS_CONSUMERS")

	if redisHost == "" || redisPort == "" {
		return errors.New("Unable to load Redis environment variables")
	} else if mongoHost == "" || mongoPort == "" {
		return errors.New("Unable to load MongoDB environment variables")
	} else if consumers == "" {
		return errors.New("Unable to load go routine redis consumers")
	}

	return nil
}
