package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

type Config struct {
	Port         int
	Rules        []Rule
	CIDROverride []string `json:"cidr_override"`
	CIDRs        []*net.IPNet
}

type Rule struct {
	Command  string
	Criteria []Criteria
}

type Criteria struct {
	Type       string
	Owner      string
	Repository string
}

func loadConfig(fileName string) Config {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Error loading config file: ", err)
	}
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Invalid config file: ", err)
	}

	//get valid CIDRs from Github
	if len(config.CIDROverride) != 0 {
		config.CIDRs = parseCIDRs(config.CIDROverride)
	} else {
		config.CIDRs = getGithubCIDRs()
	}
	fmt.Println("CIDRs: ", config.CIDRs)
	return config
}

func getGithubCIDRs() []*net.IPNet {
	//request CIDRs from Github
	resp, err := http.Get("https://api.github.com/meta")
	if err != nil {
		log.Fatal("Could not load Github CIDRs")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var data struct {
		Hooks []string //we only really care about the Hooks
	}
	json.Unmarshal(body, &data)

	//convert the response into net.IPNet slice
	cidrs := parseCIDRs(data.Hooks)

	return cidrs
}