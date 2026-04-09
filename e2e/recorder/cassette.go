package main

import (
	"encoding/json"
	"os"
)

type CassetteRequest struct {
	Method       string `json:"method"`
	URL          string `json:"url"`
	Body         string `json:"body"`
	CanonicalKey string `json:"canonical_key"`
}

type CassetteResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Interaction struct {
	Request  CassetteRequest  `json:"request"`
	Response CassetteResponse `json:"response"`
}

type Cassette struct {
	Interactions []Interaction `json:"interactions"`
}

func loadCassette(path string) (*Cassette, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Cassette
	return &c, json.Unmarshal(data, &c)
}

func saveCassette(path string, c *Cassette) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func canonicalKey(method, url, body string) string {
	return method + "|" + url + "|" + body
}
