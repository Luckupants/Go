//go:build !solution

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
)

type HandlerError struct {
	statusCode int
	error
}

type ChampInfo struct {
	Athlete string
	Age     int
	Country string
	Year    int
	Date    string
	Sport   string
	Gold    int
	Silver  int
	Bronze  int
	Total   int
}

type GetChampInfo struct {
	Athlete      string                    `json:"athlete"`
	Country      string                    `json:"country"`
	Medals       map[string]int            `json:"medals"`
	MedalsByYear map[string]map[string]int `json:"medals_by_year"`
}

func NewGetChampInfo() *GetChampInfo {
	return &GetChampInfo{Medals: make(map[string]int), MedalsByYear: make(map[string]map[string]int)}
}

func (gci *GetChampInfo) Update(champ ChampInfo) {
	gci.Athlete = champ.Athlete
	gci.Country = champ.Country
	gci.Medals["bronze"] += champ.Bronze
	gci.Medals["silver"] += champ.Silver
	gci.Medals["gold"] += champ.Gold
	gci.Medals["total"] += champ.Total
	if gci.MedalsByYear[strconv.Itoa(champ.Year)] == nil {
		gci.MedalsByYear[strconv.Itoa(champ.Year)] = make(map[string]int)
	}
	gci.MedalsByYear[strconv.Itoa(champ.Year)]["bronze"] += champ.Bronze
	gci.MedalsByYear[strconv.Itoa(champ.Year)]["silver"] += champ.Silver
	gci.MedalsByYear[strconv.Itoa(champ.Year)]["gold"] += champ.Gold
	gci.MedalsByYear[strconv.Itoa(champ.Year)]["total"] += champ.Total
}

type GetCountryInfo struct {
	Country string `json:"country"`
	Gold    int    `json:"gold"`
	Silver  int    `json:"silver"`
	Bronze  int    `json:"bronze"`
	Total   int    `json:"total"`
}

func NewGetCountryInfo() *GetCountryInfo {
	return &GetCountryInfo{}
}

func (gci *GetCountryInfo) Update(champ ChampInfo) {
	gci.Country = champ.Country
	gci.Gold += champ.Gold
	gci.Silver += champ.Silver
	gci.Bronze += champ.Bronze
	gci.Total += champ.Total
}

func main() {
	port := flag.String("port", "8000", "port to start server on")
	pathToData := flag.String("data", "./olympics/testdata/olympicWinners.json", "path to json with data on olympic champions")
	flag.Parse()
	data, err := os.ReadFile(*pathToData)
	if err != nil {
		fmt.Printf("error reading data %v", err)
		return
	}
	var info []ChampInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		fmt.Println("invalid data", err)
		return
	}
	server := http.NewServeMux()
	wrapErrorReply := func(h func(w http.ResponseWriter, r *http.Request) ([]byte, error)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			var err HandlerError
			toWrite, er := h(w, r)
			if er != nil {
				errors.As(er, &err)
				w.WriteHeader(err.statusCode)
				if err.error != nil {
					w.Header().Set("error", err.Error())
				}
			}
			_, _ = w.Write(toWrite)
		}
	}
	getInfoHandler := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
		name := r.URL.Query().Get("name")
		ans := NewGetChampInfo()
		for i := len(info) - 1; i >= 0; i-- {
			champ := info[i]
			if champ.Athlete == name {
				ans.Update(champ)
			}
		}
		if ans.Athlete == "" {
			return nil, HandlerError{statusCode: http.StatusNotFound, error: errors.New("no such name in data")}
		}
		toWrite, err := json.Marshal(ans)
		if err != nil {
			return nil, HandlerError{statusCode: http.StatusInternalServerError, error: err}
		}
		w.Header().Set("Content-Type", "application/json")
		return toWrite, nil
	}
	server.HandleFunc("GET /athlete-info", wrapErrorReply(getInfoHandler))
	getTopAthletesHandler := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
		sport := r.URL.Query().Get("sport")
		limitData := r.URL.Query().Get("limit")
		var limit int
		if limitData == "" {
			limit = 3
		} else {
			limit, err = strconv.Atoi(limitData)
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusBadRequest, error: err}
			}
		}
		champInfos := make(map[string]*GetChampInfo)
		for _, champ := range info {
			if champ.Sport == sport {
				if _, ok := champInfos[champ.Athlete]; !ok {
					champInfos[champ.Athlete] = NewGetChampInfo()
				}
				champInfos[champ.Athlete].Update(champ)
			}
		}
		if len(champInfos) == 0 {
			return nil, HandlerError{statusCode: http.StatusNotFound, error: errors.New("no such sport in data")}
		}
		var results []*GetChampInfo
		for _, value := range champInfos {
			results = append(results, value)
		}
		sort.Slice(results, func(i, j int) bool {
			if results[i].Medals["gold"] != results[j].Medals["gold"] {
				return results[i].Medals["gold"] > results[j].Medals["gold"]
			}
			if results[i].Medals["silver"] != results[j].Medals["silver"] {
				return results[i].Medals["silver"] > results[j].Medals["silver"]
			}
			if results[i].Medals["bronze"] != results[j].Medals["bronze"] {
				return results[i].Medals["bronze"] > results[j].Medals["bronze"]
			}
			return results[i].Athlete < results[j].Athlete
		})
		results = results[:min(limit, len(results))]
		toWrite, err := json.Marshal(results)
		if err != nil {
			return nil, HandlerError{statusCode: http.StatusInternalServerError, error: err}
		}
		w.Header().Set("Content-Type", "application/json")
		return toWrite, nil
	}
	server.HandleFunc("GET /top-athletes-in-sport", wrapErrorReply(getTopAthletesHandler))
	getTopCountriesHandler := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
		yearData := r.URL.Query().Get("year")
		if yearData == "" {
			return nil, HandlerError{statusCode: http.StatusBadRequest, error: errors.New("you need to specify data")}
		}
		year, err := strconv.Atoi(yearData)
		if err != nil {
			return nil, HandlerError{statusCode: http.StatusBadRequest, error: err}
		}
		limitData := r.URL.Query().Get("limit")
		var limit int
		if limitData == "" {
			limit = 3
		} else {
			limit, err = strconv.Atoi(limitData)
			if err != nil {
				return nil, HandlerError{statusCode: http.StatusBadRequest, error: err}
			}
		}
		countryInfos := make(map[string]*GetCountryInfo)
		for _, champ := range info {
			if champ.Year == year {
				if _, ok := countryInfos[champ.Country]; !ok {
					countryInfos[champ.Country] = NewGetCountryInfo()
				}
				countryInfos[champ.Country].Update(champ)
			}
		}
		if len(countryInfos) == 0 {
			return nil, HandlerError{statusCode: http.StatusNotFound, error: errors.New("no such year found in data")}
		}
		var results []*GetCountryInfo
		for _, value := range countryInfos {
			results = append(results, value)
		}
		sort.Slice(results, func(i, j int) bool {
			if results[i].Gold != results[j].Gold {
				return results[i].Gold > results[j].Gold
			}
			if results[i].Silver != results[j].Silver {
				return results[i].Silver > results[j].Silver
			}
			if results[i].Bronze != results[j].Bronze {
				return results[i].Bronze > results[j].Bronze
			}
			return results[i].Country < results[j].Country
		})

		results = results[:min(limit, len(results))]
		toWrite, err := json.Marshal(results)
		if err != nil {
			return nil, err
		}
		w.Header().Set("Content-Type", "application/json")
		return toWrite, nil
	}
	server.HandleFunc("GET /top-countries-in-year", wrapErrorReply(getTopCountriesHandler))
	_ = http.ListenAndServe(":"+*port, server)
}
