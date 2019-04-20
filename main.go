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
	"golang.org/x/net/publicsuffix"
)

func extractWorker(input <-chan string, output chan<- pageInfo, wg *sync.WaitGroup) {
	defer (*wg).Done()
	for line:= range(input) {
		if line[:4] == "null" {
			continue
		}
		grabData := encodedGrab{}
		json.Unmarshal([]byte(line), &grabData)
		if grabData.Error != nil || len(*grabData.Error) > 0 {
			continue
		}
		if grabData.Data.HTTP.Response.StatusCode == 404{
			continue
		}
		if useValidDomains{
			if _, ok := validDomains[grabData.Domain]; !ok {
				continue
			}
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
		if len(sTmp) < 2 {
			continue
		}
		output <- pageInfo{
			grabData.Domain,
			grabData.URL,
			isRedirect,
			utf8.RuneCountInString(sTmp),
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
	fieldLine := "parked,rawPageLen,jsCodeRatio,aTagLen,readbleTextRatio,index,follow,archive,randomSubdomainIP,domain,URL\n"
	f.WriteString(fieldLine)
	for info := range(input) {
		randSubDomainDNS := 0
		etldPOne,_ := publicsuffix.EffectiveTLDPlusOne(info.domain)
		if v,ok := domainDnsInfo[etldPOne]; ok {
			randSubDomainDNS = v
		}
		f.WriteString(parkLabel + "," + ouputPageInfo(info, randSubDomainDNS))
	}
}

func parseDNSInfo() {
	var f *os.File
	var err error
	ipv4Addrs := make (map[string] map[string] bool)
	if f, err = os.Open(rrFile); err == nil {
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := s.Text()
			rr := Result{}
			json.Unmarshal([]byte(line), &rr)
			if rr.Status != "NO_ERROR" {
				continue
			}
			etldPOne,parseErr := publicsuffix.EffectiveTLDPlusOne(rr.Name)
			if parseErr != nil {
				continue
			}
			if _, ok := ipv4Addrs[etldPOne]; !ok {
				ipv4Addrs[etldPOne] = make(map[string] bool)
			}
			for _,ipv4 := range(rr.Data.IPv4Addresses) {
				ipv4Addrs[etldPOne][ipv4] = true
			}
		}
	}
	domainDnsInfo = make(map[string] int)
	if f, err = os.Open(randSubRRFile); err == nil {
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := s.Text()
			rr := Result{}
			json.Unmarshal([]byte(line), &rr)
			if rr.Status != "NO_ERROR" {
				continue
			}
			etldPOne,parseErr := publicsuffix.EffectiveTLDPlusOne(rr.Name)
			if parseErr != nil {
				continue
			}
			if _, ok := ipv4Addrs[etldPOne]; !ok {
				domainDnsInfo[etldPOne] = 1
			}
			for _,ipv4 := range(rr.Data.IPv4Addresses) {
				if _,ok := ipv4Addrs[etldPOne][ipv4]; ok {
					domainDnsInfo[etldPOne] = 2
					break
				}
			}
		}
	}

}
var (
	parkLabel string
	inputFile string
	outputFile string
	rrFile string
	randSubRRFile string
	validDomainsFile string
	useValidDomains bool
	validDomains map[string] bool
	domainDnsInfo map[string]int

)

func main() {
	flags := flag.NewFlagSet("flags", flag.ExitOnError)
	flags.StringVar(&parkLabel, "park-label", "1",
		"use 0 or 1 to indicate if the domains in the input file are parked")
	flags.StringVar(&inputFile, "input-file", "/data1/nsrg/kwang40/fullData/2019-03-11/banners.json",
		"file contained zgrab data")
	flags.StringVar(&rrFile, "rr-file", "",
		"RR file for domains")
	flags.StringVar(&randSubRRFile, "randRR-file", "",
		"RR file for random subdomains")
	flags.StringVar(&outputFile, "output-file", "-", "file for output, stdout as default")
	flags.StringVar(&validDomainsFile, "valid-domains", "",
		"file contains valid domains, default is none")
	flags.Parse(os.Args[1:])

	parseDNSInfo()
	// Fulfill validDomains
	useValidDomains = false
	if len(validDomainsFile) > 0 {
		var f *os.File
		var err error
		validDomains = make (map[string] bool)
		if f, err = os.Open(validDomainsFile); err == nil {
			s := bufio.NewScanner(f)
			useValidDomains = true
			for s.Scan() {
				line := s.Text()
				validDomains[strings.TrimSuffix(line, "\n")] = true
			}
		}
		f.Close()

	}



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