package http

import (
	"bufio"
	"fmt"
	"google.golang.org/api/drive/v3"
	"log"
	"os"
	"strconv"
	"strings"
)

func Fetch(service *drive.Service, count int64, nextPageToken string) {
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
	fmt.Println("«Files»")
	if len(result.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		tree = make(map[string][]string)
		// append parameters to file tree
		for _, i := range result.Files {
			if strings.Contains(i.MimeType, "folder") {
				tree[i.Id] = append(tree[i.Name], "| "+i.Name)
			} else if i.Parents != nil {
				tree[i.Parents[0]] = append(tree[i.Parents[0]], "->\t("+strconv.Itoa(int(i.Size))+" kb)\t\t"+i.Name)
			}
		}
	}

	// write file tree
	for _, value := range tree {
		for _, v := range value {
			fmt.Printf("%s\n", v)
		}
	}

	// continuous fetch
	if result.NextPageToken != "" {
		fmt.Println("\n* Select Operation :")
		fmt.Printf("\t* Continue to fetch %d item, \t\tPress Enter\n", count)
		fmt.Println("\t* Download file with given name, \tPress d")
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
				Download(scanner, service, result)
			}

			if ascii == 10 {
				Fetch(service, count, result.NextPageToken)
			}
		}
	}
}
