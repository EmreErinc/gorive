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
	"path"
	"runtime"
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
		// append parameters to file tree
		for _, i := range result.Files {
			if strings.Contains(i.MimeType, "folder") {
				tree[i.Id] = append(tree[i.Name], "| "+i.Name)
			} else if i.Parents != nil {
				tree[i.Parents[0]] = append(tree[i.Parents[0]], "->\t"+strconv.Itoa(int(i.Size))+" kb\t"+i.Name)
			}
		}
	}

	// write file tree
	for _, value := range tree {
		for _, v := range value {
			fmt.Printf("%s\n", v)
		}
	}

	// continuous
	if result.NextPageToken != "" {
		fmt.Printf("\nContinue to fetch %d item, \nPress Enter...\n", count)
		scanner := bufio.NewScanner(os.Stdin)
		for {
			// only read single characters, the rest will be ignored!!
			consoleReader := bufio.NewReaderSize(os.Stdin, 1)
			fmt.Print("\n>")
			input, _ := consoleReader.ReadByte()
			ascii := input

			// ESC = 27 and Ctrl-C = 3
			if ascii == 27 || ascii == 3 {
				fmt.Println("Exiting...")
				os.Exit(0)
			}

			if ascii == 100 {
				fmt.Print("Download >> ")
				scanner.Scan()

				fileName := scanner.Text()
				for _, i := range result.Files {
					//fmt.Printf("file name : %s ---- desired file name : %s\n",i.Name, file)
					if i.Name == fileName {
						file, err := service.Files.Get(i.Id).Fields("*").Do()

						b, err := ioutil.ReadFile("credentials.json")
						if err != nil {
							log.Fatalf("Unable to read client secret file: %v", err)
						}

						config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
						if err != nil {
							log.Fatalf("Unable to parse client secret file to config: %v", err)
						}
						client := getClient(config)

						dlFile, err := DownloadFile(client.Transport, file)
						if err != nil {
							fmt.Printf("an error occurred while '%s' dowloading\n", i.Name)
						}

						fmt.Printf("Download : %s", dlFile)
					}
				}
			}

			if ascii == 10 {
				fetch(service, count, result.NextPageToken)
			}
		}
	}
}

func DownloadFile(t http.RoundTripper, f *drive.File) (string, error) {
	// t parameter should use an oauth.Transport
	downloadUrl := f.WebContentLink
	if downloadUrl == "" {
		// If there is no downloadUrl, there is no body
		fmt.Printf("An error occurred: File is not downloadable")
		return "", nil
	}
	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}
	resp, err := t.RoundTrip(req)
	// Make sure we close the Body later
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}

	//res := strings.TrimLeft(string(body), "<HEAD>\n<TITLE>Moved Temporarily</TITLE>\n</HEAD>\n<BODY BGCOLOR=\"#FFFFFF\" TEXT=\"#000000\">\n<H1>Moved Temporarily</H1>\nThe document has moved <A HREF=\"")
	//fmt.Println(res)

	// Moved Temporarily
	//xml := strings.NewReader(string(body))
	//result, err := xj.Convert(xml)
	//if err != nil {
	//	panic("An error occurred while xml parsing")
	//}



	//result1 :=strings.ReplaceAll(result.String(), "{\"HTML\": {\"HEAD\": {\"TITLE\": \"Moved Temporarily\"}, \"BODY\": {\"#content\": \".\", \"A\": {\"#content\": \"here\", \"-HREF\": \"", "")
	//result1 = strings.ReplaceAll(result1, "\"}, \"-BGCOLOR\": \"#FFFFFF\", \"-TEXT\": \"#000000\", \"H1\": \"Moved Temporarily\"}}}", "")
	//

	//req, err = http.NewRequest("GET", result1, nil)
	//if err != nil {
	//	fmt.Printf("An error occurred: %v\n", err)
	//	return "", err
	//}
	//resp, err = t.RoundTrip(req)
	//// Make sure we close the Body later
	//defer resp.Body.Close()
	//if err != nil {
	//	fmt.Printf("An error occurred: %v\n", err)
	//	return "", err
	//}
	//body, err = ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	fmt.Printf("An error occurred: %v\n", err)
	//	return "", err
	//}

	path := RootDirectory() + "/" + f.Name
	err = ioutil.WriteFile(path, body, 0644)
	if err != nil {
		panic(err)
	}

	return string(body), nil
}

func RootDirectory() string {
	_, file, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(file))
}
