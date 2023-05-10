package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("Starting server")

	api := NewAPI()

	http.HandleFunc("/api", api.Handler)
	http.ListenAndServe(":8080", nil)

}

func (a *api) Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In the handler")

	q := r.URL.Query().Get("q")
	data, cacheHit, err := a.getData(r.Context(), q)
	if err != nil {
		fmt.Printf("error calling data source: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := APIResponse{
		Cache: cacheHit,
		Data:  data,
	}

	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		fmt.Printf("Error in encoding response %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *api) getData(ctx context.Context, q string) ([]NominatimResponse, bool, error) {

	//is the query cached? [is the data present in cache?]
	value, err := a.cache.Get(ctx, q).Result()

	if err == redis.Nil {
		//call external DS
		escapedQ := url.PathEscape(q)
		address := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", escapedQ)

		//extract value out of the url [ http.get ]
		resp, err := http.Get(address)
		if err != nil {
			return nil, false, err
		}

		data := make([]NominatimResponse, 0)
		//read that resp
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, false, err
		}

		b, err := json.Marshal(data)
		if err != nil {
			return nil, false, err
		}

		//Set the value
		err = a.cache.Set(ctx, "q", bytes.NewBuffer(b).Bytes(), time.Second*15).Err()
		if err != nil {
			panic(err)
		}

		//return the value
		return data, false, nil

	} else if err != nil {
		return nil, false, err
	} else {
		//build response
		data := make([]NominatimResponse, 0)

		err := json.Unmarshal(bytes.NewBufferString(value).Bytes(), &data)
		if err != nil {
			return nil, false, err
		}

		//return response
		return data, true, nil

	}
}

type api struct {
	cache *redis.Client
}

func NewAPI() *api {
	redisAddress := fmt.Sprintf("%s:6379", os.Getenv("REDIS_URL"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &api{
		cache: rdb,
	}
}

//we need to register a response from the PAI, also check whether its a cache hit or miss

type APIResponse struct {
	Cache bool                `json:"cache"`
	Data  []NominatimResponse `json:"data"`
}

type NominatimResponse struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
	Icon        string   `json:"icon"`
}
