package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const prefix string = "!weather"

type Weather struct {
	Location struct {
		Name      string  `json:"name"`
		Region    string  `json:"region"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
		Timezone  string  `json:"tz_id"`
		LocalTime string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdated string  `json:"last_updated"`
		TempC       float64 `json:"temp_c"`
		TempF       float64 `json:"temp_f"`
		IsDay       int     `json:"is_day"`
		Condition   struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMph      float64 `json:"wind_mph"`
		WindKph      float64 `json:"wind_kph"`
		WindDegree   int     `json:"wind_degree"`
		WindDir      string  `json:"wind_dir"`
		PressureMb   float64 `json:"pressure_mb"`
		PressureIn   float64 `json:"pressure_in"`
		Humidity     int     `json:"humidity"`
		Cloud        int     `json:"cloud"`
		FeelsLikeC   float64 `json:"feelslike_c"`
		FeelsLikeF   float64 `json:"feelslike_f"`
		VisibilityKm float64 `json:"vis_km"`
		VisibilityMi float64 `json:"vis_miles"`
		UVIndex      float64 `json:"uv"`
		GustMph      float64 `json:"gust_mph"`
		GustKph      float64 `json:"gust_kph"`
	} `json:"current"`
}

func main() {
	sess, err := discordgo.New(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		args := strings.Split(m.Content, " ")
		if args[0] != prefix {
			return
		}

		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Please provide a city name.")
			return
		}

		if len(args) == 2 {
			city := args[1]
			temp, err := getWeather(city)
			if err != nil {
				log.Println("Error getting the current weather: ", err)
				s.ChannelMessageSend(m.ChannelID, "Failed to get weather data for "+city)
				return
			}
			message := fmt.Sprintf("The weather in %s is %.2f°F", city, temp)
			s.ChannelMessageSend(m.ChannelID, message)
		} else if len(args) == 3 {
			city := args[1] + " " + args[2]
			city_url := args[1] + "%20" + args[2]
			temp, err := getWeather(city_url)
			if err != nil {
				log.Println("Error getting the current weather: ", err)
				s.ChannelMessageSend(m.ChannelID, "Failed to get weather data for "+city)
				return
			}
			message := fmt.Sprintf("The weather in %s is %.2f°F", city, temp)
			s.ChannelMessageSend(m.ChannelID, message)
		}

	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("the bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func getWeather(city string) (float64, error) {
	requestURL := os.Getenv("API_URL")

	res, err := http.Get(requestURL + city)
	if err != nil {
		fmt.Print(err.Error())
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("Weather API not available, status code: %d", res.StatusCode)
	}

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var weather Weather
	err = json.Unmarshal(resData, &weather)
	if err != nil {
		return 0, err
	}

	temperature := weather.Current.TempF
	return temperature, nil
}
