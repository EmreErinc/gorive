package http

import (
	"bufio"
	"fmt"
	"google.golang.org/api/drive/v3"
	"gorive/pkg/auth"
	"gorive/pkg/physical"
	"net/http"
	"strings"
)

func Download(scanner *bufio.Scanner, service *drive.Service, result *drive.FileList) {
	fmt.Print("Download >> ")
	scanner.Scan()

	fileName := scanner.Text()
	for _, i := range result.Files {
		if i.Name == fileName {
			file, err := service.Files.Get(i.Id).Fields("*").Do()

			client := auth.GetClientFromFile()

			_, err = DownloadFile(client.Transport, file)
			if err != nil {
				fmt.Printf("an error occurred while '%s' dowloading\n", i.Name)
			}

			fmt.Printf("\nDownloaded : %s\n", fileName)
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

	response, _ := get(downloadUrl, t)

	// google returns 'Moved Temporarily' and we clean it
	url := strings.ReplaceAll(string(response), "<HTML>\n<HEAD>\n<TITLE>Moved Temporarily</TITLE>\n</HEAD>\n<BODY BGCOLOR=\"#FFFFFF\" TEXT=\"#000000\">\n<H1>Moved Temporarily</H1>\nThe document has moved <A HREF=\"", "")
	url = strings.ReplaceAll(url, "\">here</A>.\n</BODY>\n</HTML>", "")
	url = strings.TrimSpace(url)

	response, _ = get(url, t)
	physical.SaveAsPhysicalFile(f.Name, response)

	return string(response), nil
}
