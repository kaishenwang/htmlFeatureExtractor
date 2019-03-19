package main

import (
	"golang.org/x/net/html"
	"log"
	"bytes"
	"sync"
	"flag"
	"os"
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

func extractWorker(input <-chan string, output chan<- pageInfo, wg *sync.WaitGroup) {
	defer (*wg).Done()
	for line:= range(input) {
		grabData := encodedGrab{}
		json.Unmarshal([]byte(line), &grabData)
		if grabData.Error != nil && len(*grabData.Error) > 0 {
			continue
		}
		sTmp := html.UnescapeString(grabData.Data.HTTP.Response.BodyText)
		respBytes := []byte(sTmp)
		treeParser := bytes.NewReader(respBytes)
		doc, err := html.Parse(treeParser)
		if err != nil {
			log.Fatal(err)
		}
		treeRes := parseRoot(doc, 0)
		isRedirect := 2
		if strings.Index(grabData.URL, "www") == -1 {
			isRedirect = 0
			if grabData.Data.HTTP.RedirectResponseChain != nil {
				for _, res := range grabData.Data.HTTP.RedirectResponseChain {
					if res.Header != nil {
						for k,v := range(res.Header) {
							if k == "location" {
								for _, url := range v {
									if strings.Index(url, "www."+grabData.Domain) != -1 {
										isRedirect  = 1
									}
								}
							}
						}

					}
				}
			}
		}
		output <- pageInfo{
			grabData.Domain,
			grabData.URL,
			isRedirect,
			utf8.RuneCountInString(grabData.Data.HTTP.Response.BodyText),
			treeRes,
		}
	}
}

func outputWriter(input <-chan pageInfo, wg *sync.WaitGroup) {
	defer (*wg).Done()
	var f *os.File
	if outputFile == "" || outputFile == "-" {
		f = os.Stdout
	} else {
		var err error
		f, err = os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal("unable to open output file:", err.Error())
		}
	}
	defer f.Close()
	fieldLine := "domain,URL,wwwRedirect,rawPageLen,headTextLen,bodyTextLen,codeLen,aTagCount,aTagLen,frameCount," +
		"index,follow,archive,snippet,translate,imageindex,unavailable_after\n"
	f.WriteString(fieldLine)
	for info := range(input) {
		f.WriteString(ouputPageInfo(info))
	}
}

var (
	inputFile string
	outputFile string
)

func main() {
	flags := flag.NewFlagSet("flags", flag.ExitOnError)
	flags.StringVar(&inputFile, "input-file", "/data1/nsrg/kwang40/fullData/2019-03-03/banners.json",
		"file contained zgrab data")
	flags.StringVar(&outputFile, "output-file", "-", "file for output, stdout as default")
	flags.Parse(os.Args[1:])

	inputChan := make (chan string)
	outputChan := make (chan pageInfo)
	var outputWG sync.WaitGroup
	outputWG.Add(1)
	go outputWriter(outputChan, &outputWG)

	workerCount := 10
	var workerWG sync.WaitGroup
	workerWG.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go extractWorker(inputChan, outputChan, &workerWG)
	}

	var f *os.File
	var err error
	if f, err = os.Open(inputFile); err != nil {
		log.Fatal("unable to open input file:", err.Error())
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	buf := make([]byte, 0, 10*64*1024)
	s.Buffer(buf, 10*1024*1024)

	for s.Scan() {
		line := s.Text()
		inputChan<-line
	}
	if err := s.Err(); err != nil {
		fmt.Println(err)
	}
	close(inputChan)
	workerWG.Wait()
	close(outputChan)
	outputWG.Wait()
}

// Utility
func boolToString(b bool) string {
	if b {
		return "1"
	} else {
		return "0"
	}
}