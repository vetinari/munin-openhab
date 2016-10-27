package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ErrMissingSymlink = fmt.Errorf("missing symlink")
var itemFilter *regexp.Regexp
var version = "0.1"

func main() {
	mode := "print"
	if len(os.Args) != 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "autoconf":
		fmt.Fprintf(os.Stdout, "no\n")

	case "config":
		itemName, err := getItemName()
		if err != nil {
			die("failed to get item name: %s", err)
		}
		item, err := fetchItem(itemName)
		if err != nil {
			die("failed to fetch item: %s", err)
		}
		printConfig(item)

	case "version":
		fmt.Fprintf(os.Stdout, "openhab_ munin plugin v%s\n", version)

	case "print":
		itemName, err := getItemName()
		if err != nil {
			die("failed to get item name: %s", err)
		}
		item, err := fetchItem(itemName)
		if err != nil {
			die("failed to fetch item: %s", err)
		}
		printValues(item)

	default:
		die("unknown command: %s\n", mode)
	}
}

func getItemName() (string, error) {
	parts := strings.SplitN(filepath.Base(os.Args[0]), "_", 2)
	if len(parts) == 1 || parts[1] == "" {
		return "", ErrMissingSymlink
	}
	return parts[1], nil
}

type Item struct {
	Type    string  `json:"type"`
	Name    string  `json:"name"`
	State   string  `json:"state"`
	Members []*Item `json:"members"`
}

func openHABURL(itemName string) string {
	baseURL := os.Getenv("server")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	return baseURL + "/rest/items/" + itemName
}

func fetchItem(name string) (*Item, error) {
	req, err := http.NewRequest("GET", openHABURL(name), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}
	req.Header.Set("Accept", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call openHAB")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("call to openHAB returned non OK status %d (%s)", res.StatusCode, res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}

	var item *Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse JSON")
	}
	return item, nil
}

func printConfig(item *Item) {
	re := os.Getenv("item_filter")
	if re != "" {
		itemFilter = regexp.MustCompile(re)
	}

	fmt.Fprintf(os.Stdout, "graph_title %s\n", getEnv("title", item))
	if val := getEnv("graph_args", item); val != "" {
		fmt.Fprintf(os.Stdout, "graph_args %s\n", val)
	}
	if val := getEnv("graph_scale", item); val != "" {
		fmt.Fprintf(os.Stdout, "graph_scale %s\n", val)
	}
	fmt.Fprintf(os.Stdout, "graph_category sensors\n")
	if val := getEnv("vlabel", item); val != "" {
		fmt.Fprintf(os.Stdout, "graph_vlabel %s\n", val)
	}
	if item.Members == nil {
		fmt.Fprintf(os.Stdout, "%s.label %s\n", item.Name, getEnv("label", item))
		fmt.Fprintf(os.Stdout, "%s.draw %s\n", item.Name, getEnv("draw", item))
	} else {
		for _, member := range item.Members {
			if !filtered(member) {
				fmt.Fprintf(os.Stdout, "%s.label %s\n", member.Name, getEnv("label", member))
				fmt.Fprintf(os.Stdout, "%s.draw %s\n", member.Name, getEnv("draw", member))
			}
		}
	}
}

func printValues(item *Item) {
	re := os.Getenv("item_filter")
	if re != "" {
		itemFilter = regexp.MustCompile(re)
	}

	if item.Members == nil {
		printItem(item)
	} else {
		for _, member := range item.Members {
			if member.Type != "GroupItem" {
				if !filtered(member) {
					printItem(member)
				}
			}
		}
	}
}

func filtered(item *Item) bool {
	if itemFilter == nil {
		return false
	}
	return itemFilter.MatchString(item.Name)
}

func printItem(item *Item) {
	switch item.Type {
	case "NumberItem", "DimmerItem":
	case "StringItem", "GroupItem":
		return
	case "SwitchItem", "ContactItem":
		switch strings.ToLower(item.State) {
		case "on", "open":
			item.State = "1"
		case "off", "closed":
			item.State = "0"
		default:
			item.State = "0"
		}
	}
	fmt.Fprintf(os.Stdout, "%s.value %s\n", item.Name, item.State)
}

func die(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, os.Args[0]+": "+format, v...)
	os.Exit(1)
}

func getEnv(name string, item *Item) string {
	if val := os.Getenv(fmt.Sprintf("%s_%s", name, item.Name)); val != "" {
		return val
	}
	if val := os.Getenv(name); val != "" {
		return val
	}
	switch name {
	case "graph_args", "graph_scale", "vlabel":
		// no defaults
		return ""
	case "draw":
		return "LINE"
	case "label", "title":
		return item.Name
	default:
		return ""
	}
}
