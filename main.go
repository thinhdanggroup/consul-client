package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

type Config struct {
	ConsulTarget string
	Key          string
	Token        string
	Timeout      int
	Destination  string
}

type ConsulResponse struct {
	Key   string
	Value string
}

func checkErr(desc string, err error) {
	if err != nil {
		log.Fatalf("Fail %s \n %s", desc, err)
	}
}

func parseConfig(config []byte) ConsulResponse {
	rawData := string(config)
	splitData := strings.Split(rawData, "]")
	rawData = strings.Replace(splitData[0], "[", "", -1)
	// fmt.Printf("%s", rawData)

	var data ConsulResponse

	err := json.Unmarshal([]byte(rawData), &data)
	checkErr("unmarshall config. \nRawData: "+rawData, err)
	return data
}

func getKey(config *Config) {
	tr := &http.Transport{
		IdleConnTimeout:    time.Duration(config.Timeout) * time.Second,
		DisableCompression: true,
	}

	client := &http.Client{Transport: tr}

	target := config.ConsulTarget + config.Key
	req, err := http.NewRequest("GET", target, nil)
	checkErr("request connect consul", err)

	req.Header.Add("Authorization", config.Token)

	resp, err := client.Do(req)
	checkErr("request get key in consul", err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	checkErr("read response from consul", err)

	// parse response
	response := parseConfig(body)
	data, err := b64.StdEncoding.DecodeString(response.Value)
	checkErr("decode key", err)

	// write file
	err = ioutil.WriteFile(config.Destination, data, 0644)
	checkErr("write config to file", err)
}

func main() {
	config := &Config{}
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "get",
				Aliases: []string{"g"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "consul-target", Aliases: []string{"t"}, Value: "http://10.60.45.6:8500", Destination: &config.ConsulTarget},
					&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Value: "zas/zas/bootstrap/dev/stable", Destination: &config.Key},
					&cli.StringFlag{Name: "token", Aliases: []string{"a"}, Value: "Bearer f16f8d68-c181-7d44-e90c-377390b8314e", Destination: &config.Token},
					&cli.IntFlag{Name: "timeout-in-sec", Value: 5, Destination: &config.Timeout},
					&cli.StringFlag{Name: "destination", Aliases: []string{"d"}, Value: "file.config", Destination: &config.Destination},
				},
				Usage: "Get Key in Consul",
				Action: func(c *cli.Context) error {
					getKey(config)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
