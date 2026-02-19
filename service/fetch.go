package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	url = "https://raw.githubusercontent.com/onurusluca/turkey-geo-api/refs/heads/main/data/neighborhoods.json"
)

type Neighborhood struct {
	ProvinceID int    `json:"provinceId"`
	DistrictID int    `json:"districtId"`
	ID         int    `json:"id"`
	Province   string `json:"province"`
	District   string `json:"district"`
	Name       string `json:"name"`
}

func FetchNeighborhoods(neighborhoodChan chan<- Neighborhood) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status: %s", resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)

	if _, err = decoder.Token(); err != nil {
		return fmt.Errorf("Error reading opening bracket: %v", err)
	}

	count := 0
	for decoder.More() {
		var n Neighborhood
		err := decoder.Decode(&n)
		if err != nil {
			return fmt.Errorf("Error decoding item: %v", err)
		}

		neighborhoodChan <- n
		count++
	}

	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("Error reading closing bracket: %v", err)
	}

	return nil
}
