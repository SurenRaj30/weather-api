package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// "strconv"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// data struc
type Weather struct {
	QueryCost 		int  `json:"queryCost"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	ResolvedAddress string  `json:"resolvedAddress"`
	Description     string  `json:"description"`
}

// for incoming request body
type CityRequest struct {
	City string `json:"city"`
}

var redisClient *redis.Client

// Function to initialize the Redis client
func initRedis() {
	// redisHost := os.Getenv("REDIS_HOST")
	// redisPort := os.Getenv("REDIS_PORT")
	// redisDBStr := os.Getenv("REDIS_DB")

	// Convert redisDB from string to int
	// redisDB, err := strconv.Atoi(redisDBStr)
	// if err != nil {
	// 	log.Fatalf("Invalid Redis DB value: %v", err)
	// }

	// create new connection to Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0, // Set the converted DB number
	})

	
	// Check if Redis is reachable
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis");

	err = redisClient.Set(context.Background(), "mykey", "myvalue", 0).Err()
	if err != nil {
		log.Fatalf("Error setting key in Redis: %v", err)
	}

	// Test getting a key from Redis
	value, err := redisClient.Get(context.Background(), "mykey").Result()
	if err != nil {
		log.Fatalf("Error getting key from Redis: %v", err)
	}
	fmt.Printf("Value of mykey: %s\n", value)
	
}


func weatherHandler(w http.ResponseWriter, r *http.Request)  {
	//** First Step: Read and process user's request

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	
	// Unmarshal the JSON request body into the CityRequest struct (json to Object)
	var cityRequest CityRequest
	err = json.Unmarshal(body, &cityRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the city from the request
	city := cityRequest.City
	if city == "" {
		http.Error(w, "City parameter is required", http.StatusBadRequest)
		return
	}

	//** Second Step: Check redis cache

	// Replace spaces with underscores for the Redis key
    redisKey := strings.ReplaceAll(city, " ", "_")
    fmt.Println("Querying Redis for city:", redisKey)

	// check redis cache, creates empty context
	ctx := context.Background();
	cachedData, err := redisClient.Get(ctx, city).Result();

	// key does not exist in redis, call third-party API
	if err == redis.Nil {
		//** Third Step: Build URL to fetch data from Visual Crossing API

		apiKey := os.Getenv("WEATHER_API_KEY")
		if apiKey == "" {
			http.Error(w, "API key is missing", http.StatusInternalServerError)
			return
		}

	
		// build the query based on api key and city name
		query := fmt.Sprintf("https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/%s?unitGroup=us&key=%s&contentType=json", city, apiKey);

		//build request object, store it in req variable
		req, err := http.NewRequest(http.MethodGet, query, nil);
		if err != nil {
			log.Fatalf("Failed to create request object for /GET endpoint: %v", err)
		}

		// add required header, in this case it would be json format
		req.Header.Add("Content-Type", "application/json; charset=utf-8");

		//http client
		client := &http.Client{};
		//sends http request based on the response, store it in resp variable
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to send HTTP request: %v", err)
		}

		// reading response body
		respBody, err := ioutil.ReadAll(resp.Body);
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}
		// Always close the response body
		defer resp.Body.Close()

		//** Fourth Step: Save response to redis cache
		// Unmarshal the JSON response into the Weather struct
		var weather_redis Weather;
		err = json.Unmarshal(respBody, &weather_redis)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}
		
		// Convert the struct back to JSON for display
		jsonData, err := json.MarshalIndent(weather_redis, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}
		err = redisClient.Set(ctx, city, jsonData, 10*time.Minute).Err()
		if err != nil {
			http.Error(w, "Failed to store data in Redis", http.StatusInternalServerError)
			return
		}
		// Use the response from the API
		cachedData = string(jsonData) 
	}else if err != nil {
		log.Println("Error retrieving data from Redis:", err) // Log the specific error
		http.Error(w, "Error retrieving data from Redis", http.StatusInternalServerError)
		return
	} else {
		log.Println("cached data %s", cachedData);
	}

	//** Fifth Step: Construct client's JSON response

	// Unmarshal the JSON response into the Weather struct
	var weather Weather;
	err = json.Unmarshal([]byte(cachedData), &weather)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}
	
	// Convert the struct back to JSON for display (object to json)
	jsonData, err := json.MarshalIndent(weather, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	// Set response headers to indicate JSON content
	w.Header().Set("Content-Type", "application/json");
	w.WriteHeader(http.StatusOK);
	w.Write(jsonData);
	
}

func main() {
	// load env variable
	err := godotenv.Load();
	if err != nil {
        log.Fatal("Error loading .env file")
    }

	// Initialize Redis
	initRedis();

	// call the weater handler GET endpoint
	http.HandleFunc("/get", weatherHandler);

	// Start the server on port 8080
	fmt.Println("Server is running on port 9000...")
	log.Fatal(http.ListenAndServe(":9000", nil))
}