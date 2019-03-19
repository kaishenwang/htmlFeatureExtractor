package main

import (
	"github.com/kwang40/zgrab/zlib"
)


type treeInfo struct {
	headTextLen int
	bodyTextLen int
	index 	bool
	follow  bool
	archive bool
	snippet bool
	translate bool
	imageindex bool
	unavailable_after bool
}

type otherInfo struct {
	jsCodeLen     int
	rawPageLen    int
	frameTagCount int
	aTagCount     int
	aTagLen       int
}

type pageInfo struct {
	domain string
	url    string
	wwwRedirect int
	tInfo  treeInfo
	oInfo  otherInfo
}

func newTreeInfo() treeInfo {
	return treeInfo{
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

