package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	baseURL     = "https://translate.googleapis.com/translate_a/single?client=gtx&dt=t&sl=%s&tl=%s&q="
	configFile  = "config.json"
	defaultLang = "en"
)

var (
	sourceLang string
	targetLang string
	reverse    bool
	setLang    bool
)

func init() {
	flag.StringVar(&sourceLang, "s", "", "Source language")
	flag.StringVar(&targetLang, "t", "", "Target language")
	flag.BoolVar(&reverse, "reverse", false, "Reverse translation")
	flag.BoolVar(&setLang, "set", false, "Set default languages")
	flag.Parse()

	if setLang {
		saveConfig(sourceLang, targetLang)
	}

	if sourceLang == "" || targetLang == "" {
		loadConfig()
	}
}

func saveConfig(source, target string) {
	config := map[string]string{
		"sourceLang": source,
		"targetLang": target,
	}
	file, err := os.Create(configFile)
	if err != nil {
		fmt.Printf("Error creating config file: %v\n", err)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
	}
}

func loadConfig() {
	config := map[string]string{}
	file, err := os.Open(configFile)
	if err != nil {
		sourceLang = defaultLang
		targetLang = defaultLang
		return
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		sourceLang = defaultLang
		targetLang = defaultLang
		return
	}

	sourceLang = config["sourceLang"]
	targetLang = config["targetLang"]
}

func translate(text string, sourceLang, targetLang string) ([]string, error) {
	query := url.QueryEscape(text)
	translateURL := fmt.Sprintf(baseURL, sourceLang, targetLang) + query
	response, err := http.Get(translateURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body error: %v", err)
	}

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal error: %v", err)
	}

	var translations []string
	if len(result) > 0 {
		sentences, ok := result[0].([]interface{})
		if ok {
			for _, sentence := range sentences {
				firstSentence, ok := sentence.([]interface{})
				if ok && len(firstSentence) > 0 {
					translatedText, ok := firstSentence[0].(string)
					if ok {
						translations = append(translations, translatedText)
					}
				}
			}
		}
	}

	if len(translations) == 0 {
		return nil, fmt.Errorf("translation not found")
	}

	return translations, nil
}

func main() {
	if len(flag.Args()) < 1 {
		fmt.Println("Please provide text to translate.")
		return
	}

	text := strings.Join(flag.Args(), " ")
	if reverse {
		sourceLang, targetLang = targetLang, sourceLang
	}
	translatedTexts, err := translate(text, sourceLang, targetLang)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, translatedText := range translatedTexts {
		fmt.Printf("- %s\n", translatedText)
	}
}
