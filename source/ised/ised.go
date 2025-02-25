package ised

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pcunning/hamcall/data"
	"github.com/pcunning/hamcall/downloader"
)

func Download(wg *sync.WaitGroup) error {
	defer wg.Done()
	fmt.Println("Downloading ISED (Canada) data")
	err := downloader.FetchHttp("ised.zip", "https://apc-cap.ic.gc.ca/datafiles/amateur_delim.zip")
	if err != nil {
		log.Fatalf("Error downloading ISED (Canada) data: %v", err)
	}
	return nil
}

func Process(calls *map[string]data.HamCall) {
	start := time.Now()
	fmt.Print("processing ISED data")

	// START

	// Step 1: Unzip the file and create ised_callsigns.txt.
	reader, err := zip.OpenReader("ised.zip")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer reader.Close()

	for _, f := range reader.File {
		if f.Name == "amateur_delim.txt" {
			destinationFile, err := os.OpenFile("ised_callsigns.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				log.Fatal(err)
				return
			}
			defer destinationFile.Close()

			zippedFile, err := f.Open()
			if err != nil {
				log.Fatal(err)
				return
			}
			defer zippedFile.Close()

			if _, err := io.Copy(destinationFile, zippedFile); err != nil {
				log.Fatal(err)
				return
			}
		}
	}

	// Step 2: Open the new file.
	f, err := os.Open("ised_callsigns.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	// Step 3: Process Each Line.
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, ";")

		if columns[0] == "callsign" {
			continue
		} else {
			name := columns[1] + " " + columns[2]
			item := data.HamCall{
				Callsign: columns[0],
				Name: name,
				FirstName: columns[1],
				LastName: columns[2],
				Address: columns[3],
				City: columns[4],
				Zip: columns[6],
			}
			(*calls)[columns[0]] = item
		}
	
	}

	// END

	fmt.Printf(" ... %s\n", time.Since(start).String())
}
