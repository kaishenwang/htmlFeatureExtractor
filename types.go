package main

import (
	"github.com/kwang40/zgrab/zlib"
	"fmt"
	"strconv"
)


type treeInfo struct {
	headTextLen int
	bodyTextLen int
	codeLen int
	aTagCount int
	aTagLen  int
	frameCount int
	index 	bool
	follow  bool
	archive bool
	snippet bool
	translate bool
	imageindex bool
	unavailable_after bool
}

type pageInfo struct {
	domain string
	url    string
	wwwRedirect int
	rawPageLen    int
	tInfo  treeInfo
}

func newTreeInfo() treeInfo {
	return treeInfo{
		0,
		0,
		0,
		0,
		0,
		0,
		true,
		true,
		true,
		true,
		true,
		true,
		false,
	}
}

func accumulateTreeInfo(a treeInfo, b treeInfo) treeInfo {
	return treeInfo{
		a.headTextLen + b.headTextLen,
		a.bodyTextLen + b.bodyTextLen,
		a.codeLen + b.codeLen,
		a.aTagCount + b.aTagCount,
		a.aTagLen + b.aTagLen,
		a.frameCount + b.frameCount,
		a.index && b.index,
		a.follow && b.follow,
		a.archive && b.archive,
		a.snippet && b.snippet,
		a.translate && b.translate,
		a.imageindex && b.imageindex,
		a.unavailable_after || b.unavailable_after,
	}
}

type encodedGrab struct {
	IP             string    `json:"ip"`
	Domain         string    `json:"domain,omitempty"`
	URL			   string    `json:"url,omitempty"`
	Time           string    `json:"timestamp"`
	Data           *zlib.GrabData `json:"data,omitempty"`
	Error          *string   `json:"error,omitempty"`
	ErrorComponent string    `json:"error_component,omitempty"`
}

func ouputPageInfo(info pageInfo, randDomainDns int) string {
	pageLen := float64(info.rawPageLen)
	jsCode := float64(info.tInfo.codeLen)
	readableText := float64(info.tInfo.headTextLen + info.tInfo.bodyTextLen)
	aTagLen := float64(info.tInfo.aTagLen) / float64(info.tInfo.aTagCount)
	return fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
		strconv.FormatFloat(pageLen/10000.0,'f', 6, 64),
		strconv.FormatFloat(jsCode /pageLen,'f', 6, 64),
		strconv.FormatFloat(aTagLen/100.0,'f', 6, 64),
		strconv.FormatFloat(readableText /pageLen,'f', 6, 64),
		boolToString(info.tInfo.index), boolToString(info.tInfo.follow), boolToString(info.tInfo.archive),
		strconv.Itoa(randDomainDns),
		info.domain, info.url)
}

type Result struct {
	AlteredName string        `json:"altered_name,omitempty"`
	Name        string        `json:"name,omitempty"`
	Nameserver  string        `json:"nameserver,omitempty"`
	Class       string        `json:"class,omitempty"`
	AlexaRank   int           `json:"alexa_rank,omitempty"`
	Status      string        `json:"status,omitempty"`
	Error       string        `json:"error,omitempty"`
	Timestamp   string        `json:"timestamp,omitempty"`
	Data        ALookupResult `json:"data,omitempty"`
	Trace       []interface{} `json:"-"`
}


type ALookupResult struct {
	IPv4Addresses []string `json:"ipv4_addresses,omitempty"`
	IPv6Addresses []string `json:"ipv6_addresses,omitempty"`
}