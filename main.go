package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	fmt.Println("Starting server")

	http.HandleFunc("/api", Handler)
	http.ListenAndServe(":8080", nil)

}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In the handler")

	q := r.URL.Query().Get("q")
	data, err := getData(q)
	if err != nil {
		fmt.Printf("error calling data source: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := APIResponse{
		Cache: false,
		Data:  data,
	}

	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		fmt.Printf("Error in encoding response %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getData(q string) ([]NominatimResponse, error) {

	//remove space when entering value
	escapedQ := url.PathEscape(q)
	address := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", escapedQ)

	//extract value out of the url [ http.get ]
	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}

	data := make([]NominatimResponse, 0)
	//read that resp
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
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
