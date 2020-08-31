package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	service, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	countPtr := flag.Int64("count", 20, "will fetch item count")
	flag.Parse()

	fmt.Printf("Fetching %d items\n", *countPtr)

	fetch(service, *countPtr, "")

}

func fetch(service *drive.Service, count int64, nextPageToken string) {
	result, err := service.Files.
		List(). // what will you do
		PageSize(count). // how many item will fetch
		Fields("*"). // which fields will fetch. ex: nextPageToken, files(id, name) etc.
		PageToken(nextPageToken). // which page will fetch
		Do() // do it!!
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	var tree map[string][]string
	fmt.Println("Files:")
	if len(result.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		tree = make(map[string][]string)
		for _, i := range result.Files {
			if strings.Contains(i.MimeType, "folder") {
				tree[i.Id] = append(tree[i.Name], "| "+i.Name)
			} else if i.Parents != nil {
				tree[i.Parents[0]] = append(tree[i.Parents[0]], "->\t"+strconv.Itoa(int(i.Size))+" kb\t"+i.Name)
			}
		}
	}

	for _, value := range tree {
		for _, v := range value {
			fmt.Printf("%s\n", v)
		}
	}

	if result.NextPageToken != "" {
		fmt.Printf("\nContinue to fetch %d item, Press Enter...\n", count)

		for {
			// only read single characters, the rest will be ignored!!
			consoleReader := bufio.NewReaderSize(os.Stdin, 1)
			fmt.Print(">")
			input, _ := consoleReader.ReadByte()

			ascii := input

			// ESC = 27 and Ctrl-C = 3
			if ascii == 27 || ascii == 3 {
				fmt.Println("Exiting...")
				os.Exit(0)
			}

			if ascii == 10 {
				fetch(service, count, result.NextPageToken)
			}
		}
	}
}