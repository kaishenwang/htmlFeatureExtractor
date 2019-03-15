package main

import (
	"golang.org/x/net/html"
	"strings"
	"net/http"
	"log"
	"fmt"
	"unicode/utf8"
	"io"
	"bytes"
	"regexp"
	"io/ioutil"
)

var newLine =`
`
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
	jsCodeLen int
	rawStrLen int
	frameTagCount int
	aTagCount	int
	aTagLen 	int
}



func main() {
	resp, err := http.Get("https://www.1point3acres.com/bbs/thread-494681-1-1.html")

	respBytes, _ :=  ioutil.ReadAll(resp.Body)
	rawProcess := bytes.NewReader(respBytes)
	treeParser := bytes.NewReader(respBytes)
	otherRes := collectRawPageInfo(rawProcess)
	doc, err := html.Parse(treeParser)
	if err != nil {
		log.Fatal(err)
	}
	treeRes := parseRoot(doc, 0)
	fmt.Println(treeRes.headTextLen)
	fmt.Println(treeRes.bodyTextLen)
	if !treeRes.archive {
		fmt.Println("noarchive")
	}
	fmt.Println(otherRes.jsCodeLen)
	fmt.Println(otherRes.rawStrLen)
	fmt.Println(otherRes.frameTagCount)
	fmt.Println(otherRes.aTagCount)
	fmt.Println(otherRes.aTagLen)
}

// Utility

// countJSCode
func collectRawPageInfo(r io.Reader) otherInfo{
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	bufByte := buf.Bytes()
	reg := regexp.MustCompile(`<script`)
	scriptStarts := reg.FindAllIndex(bufByte, -1)
	reg = regexp.MustCompile(`</script>`)
	scriptEnds := reg.FindAllIndex(bufByte, -1)
	JSCodeLen := 0
	for i := 0; i < len(scriptStarts); i++ {
		JSCodeLen += utf8.RuneCount(bufByte[scriptStarts[i][0]:scriptEnds[i][1]])
	}
	reg = regexp.MustCompile(`<frame`)
	frameTagCount := len(reg.FindAllIndex(bufByte, -1))
	reg = regexp.MustCompile(`<a `)
	aTags := reg.FindAllIndex(bufByte, -1)
	aTagLen := 0
	for _,aTagStarts := range(aTags) {
		end := aTagStarts[1] + 1
		for;;end++ {
			if bufByte[end] == '>' {
				end += 1
				break
			}
		}
		aTagLen += end - aTagStarts[0]
	}

	return otherInfo{
		JSCodeLen,
		utf8.RuneCount(bufByte),
		frameTagCount,
		len(aTags),
		aTagLen,
	}
}

// Tree Parsing
func parseBodyNode(n *html.Node) treeInfo {
	res := newTreeInfo()
	if (n.Type != html.ElementNode) && (n.Type != html.TextNode) {
		return res
	}
	if (n.Type == html.ElementNode) && (n.Data == "script") {
		return res
	}
	if (n.Type == html.TextNode) {
		res.bodyTextLen += utf8.RuneCountInString(n.Data) - strings.Count(n.Data, " ") - strings.Count(n.Data, newLine)
		return res
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res =  accumulateFeatures(res,parseBodyNode(c))
	}
	return res
}

func parseHeadNode(n *html.Node, titleNode bool) treeInfo {
	res := newTreeInfo()
	if (n.Type == html.TextNode && titleNode) {
		res.headTextLen += utf8.RuneCountInString(n.Data) - strings.Count(n.Data, " ") - strings.Count(n.Data, newLine)
		return res
	}

	if (n.Type != html.ElementNode) && (n.Type != html.TextNode) {
		return res
	}
	// parse robot value
	if (n.Type == html.ElementNode) && (strings.ToLower(n.Data) == "meta") {
		attrs := n.Attr
		robotTag := false
		contents := ""
		for _,attr := range(attrs) {
			if strings.ToLower(attr.Key) == "name" {
				robotTag = strings.ToLower(attr.Val) == "robots" || strings.ToLower(attr.Val) == "googlebot"
			}
			if strings.ToLower(attr.Key) == "content" {
				contents = strings.ToLower(attr.Val)
			}
		}
		if robotTag {
			tags := strings.Split(contents, ",")
			if strings.Contains(contents, "unavailable_after") {
				res.unavailable_after = false
			}
			for _,tag := range(tags) {
				if tag == "noindex" {
					res.index = false
				} else if tag == "nofollow" {
					res.follow = false
				} else if tag == "none" {
					res.index = false
					res.follow = false
				} else if tag == "noarchive" {
					res.archive = false
				} else if tag == "nosnippet" {
					res.snippet = false
				} else if tag == "notranslate" {
					res.translate = false
				} else if tag == "noimageindex" {
					res.imageindex = false
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res =  accumulateFeatures(res, parseHeadNode(c, (n.Data == "title")))
	}
	return res
}

func parseRoot(n *html.Node, depth int) treeInfo {
	features := newTreeInfo()
	if (n.Type == html.CommentNode) {
		return features
	}

	if (n.Type == html.ElementNode) && (strings.ToLower(n.Data) == "head") {
		features = accumulateFeatures(features, parseHeadNode(n, false))
		return features
	}

	if (n.Type == html.ElementNode) && (strings.ToLower(n.Data) == "body") {
		features = accumulateFeatures(features, parseBodyNode(n))
		return features
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		tmp := parseRoot(c, depth+1)
		features = accumulateFeatures(features, tmp)
	}
	return features
}

func accumulateFeatures(a treeInfo, b treeInfo) treeInfo {
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
