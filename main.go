package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	// You can name these whatever you like, but the former must be populated correctly (see readme)
	APP_CONFIG  string = "pocketcache.config.json"
	EXPORT_FILE string = "pocketcache.export.json"
	config      Config
	USER        string

	// Templates for POST request bodies.
	request_body string = `{"consumer_key": "%s", "redirect_uri": "%s:authorizationFinished"}`
	access_body  string = `{"consumer_key": "%s", "code": "%s"}`
	// Note: Hard coded parameter in this template. This can be changed if somehow 100,000 is too many.
	retrieve_body string = `{"consumer_key": "%s", "access_token": "%s", "count": 100000}`

	// Endpoints
	token_url    string = "https://getpocket.com/v3/oauth/request"
	redirect_url string = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=about:blank"
	access_url   string = "https://getpocket.com/v3/oauth/authorize"
	get_url      string = "https://getpocket.com/v3/get_url"
)

type Config struct {
	APP_NAME      string `json:"APP_NAME"`
	CLIENT_KEY    string `json:"CLIENT_KEY"`
	ACCESS_TOKEN  string `json:"ACCESS_TOKEN"`
	REQUEST_TOKEN string `json:"REQUEST_TOKEN"`
}

func read_config() {
	config_path := path.Join(".", APP_CONFIG)
	raw, err := ioutil.ReadFile(config_path)
	if err != nil {
		fmt.Println("No config file found -- see readme/source for details.")
		os.Exit(1)
	}
	json.Unmarshal(raw, &config)
}

func write_config() {
	var with_nl bytes.Buffer
	config_path := path.Join(".", APP_CONFIG)

	marshalled, _ := json.MarshalIndent(config, "", "\t")
	with_nl.Write(marshalled)
	// Newlines at EOF are important.
	with_nl.WriteString("\n")

	_ = ioutil.WriteFile(config_path, with_nl.Bytes(), os.FileMode(int(0700)))
	// return conf or failure message
}

func export_data(data []byte) {
	export_path := path.Join(".", EXPORT_FILE)
	var with_nl bytes.Buffer

	// Arbitrary structure to unmarshal into
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	marshalled, _ := json.MarshalIndent(parsed, "", "\t")
	with_nl.Write(marshalled)
	// Newlines at EOF are still important.
	with_nl.WriteString("\n")
	err := ioutil.WriteFile(export_path, with_nl.Bytes(), os.FileMode(int(0700)))
	if err != nil {
		fmt.Println("Problem exporting file: %s", err)
	}
}

func main() {
	var reqbuf bytes.Buffer
	var body []byte

	// Read config file (file name found above)
	read_config()

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Program panicked -- did you skip a step?\nPanic message: %s", r)
		}
	}()

	// Request a token_url
	fmt.Println("Making request to token_url endpoint...")
	reqbuf.WriteString(fmt.Sprintf(request_body, config.CLIENT_KEY, config.APP_NAME))
	request_resp, _ := http.Post(token_url, "application/json", &reqbuf)
	// Check for abnormal status, exit if so (probs malformed post body or invalid key)
	if request_resp.Status != "200 OK" {
		fmt.Printf("Something went wrong, so here are some headers: %s\n\n", request_resp.Header)
		value, err := strconv.ParseInt(request_resp.Header["X-Error-Code"][0], 10, 32)
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(int(value))
		}
	}
	body, _ = ioutil.ReadAll(request_resp.Body)
	config.REQUEST_TOKEN = strings.Split(string(body), "=")[1]
	redirect_with_code := fmt.Sprintf(redirect_url, config.REQUEST_TOKEN)

	// We got our deets, now we have to ask the user to do something for us to validate our access_url code
	fmt.Printf(
		"Paste this link in a browser, and authorize this app: %s\npress any key to continue",
		redirect_with_code,
	)

	// Wait so they can do the thing. Use, but throw the err value away -- we dont care about this
	var input string
	_, input_err := fmt.Scanln(&input)
	if input_err != nil {
		fmt.Print("")
	}

	// Turn our request token_url into an access_url token_url
	reqbuf.WriteString(fmt.Sprintf(access_body, config.CLIENT_KEY, config.REQUEST_TOKEN))
	access_resp, _ := http.Post(access_url, "application/json", &reqbuf)
	body, _ = ioutil.ReadAll(access_resp.Body)
	parts := strings.Split(string(body), "&")
	if len(parts) > 1 {
		config.ACCESS_TOKEN = strings.Split(parts[0], "=")[1]
		USER = strings.Split(parts[1], "=")[1]
	} else {
		panic("Cannot parse access_url token_url, app likely not authorized.")
	}
	write_config()

	fmt.Printf("Authentication complete, retrieving data for %s...\n", USER)

	// Part that gets the stuff
	reqbuf.WriteString(fmt.Sprintf(retrieve_body, config.CLIENT_KEY, config.ACCESS_TOKEN))
	data_resp, _ := http.Post(get_url, "application/json", &reqbuf)
	body, _ = ioutil.ReadAll(data_resp.Body)

	fmt.Printf("Retrieved %s bytes.\nExporting to %s...\n", len(body), EXPORT_FILE)
	export_data(body)
	fmt.Println("Export complete.")
}
