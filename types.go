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

func ouputPageInfo(info pageInfo) string {
	return fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
		info.domain, info.url, strconv.Itoa(info.wwwRedirect),strconv.Itoa(info.rawPageLen),
		strconv.Itoa(info.tInfo.headTextLen), strconv.Itoa(info.tInfo.bodyTextLen),strconv.Itoa(info.tInfo.codeLen),
		strconv.Itoa(info.tInfo.aTagCount), strconv.Itoa(info.tInfo.aTagLen), strconv.Itoa(info.tInfo.frameCount),
		boolToString(info.tInfo.index), boolToString(info.tInfo.follow), boolToString(info.tInfo.archive),
		boolToString(info.tInfo.snippet), boolToString(info.tInfo.translate),
		boolToString(info.tInfo.imageindex), boolToString(info.tInfo.unavailable_after))
}