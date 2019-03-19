package main

import (
	"strings"
	"golang.org/x/net/html"
	"io"
	"bytes"
	"regexp"
	"unicode/utf8"
)

// Collect info from raw string
func collectRawPageInfo(r io.Reader) otherInfo{
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	bufByte := buf.Bytes()
	reg := regexp.MustCompile(`<script`)
	scriptStarts := reg.FindAllIndex(bufByte, -1)
	reg = regexp.MustCompile(`</script>`)
	scriptEnds := reg.FindAllIndex(bufByte, -1)
	JSCodeLen := 0
	ei := 0
	// There are two cases: 1. <script XXX> 2. <script XXX> XXX </script>
	for si:= 0; si < len(scriptStarts); si++ {
		// Case 2
		if ei < len(scriptEnds) && (si + 1 == len(scriptStarts) || scriptStarts[si+1][0] > scriptEnds[ei][0]) {
			JSCodeLen += utf8.RuneCount(bufByte[scriptStarts[si][0]:scriptEnds[ei][1]])
			ei += 1
		} else { // case 1
			var tmpEnd int
			tmpEnd  = scriptStarts[si][1] + 1
			for ;tmpEnd < len(bufByte) && bufByte[tmpEnd] != '>'; tmpEnd++{
			}
			JSCodeLen += utf8.RuneCount(bufByte[scriptStarts[si][0]:tmpEnd])
		}
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
var newLine =`
`

func parseBodyNode(n *html.Node) treeInfo {
	res := newTreeInfo()
	if (n.Type != html.ElementNode) && (n.Type != html.TextNode) {
		return res
	}
	if (n.Type == html.ElementNode) && (n.Data == "script") {
		return res
	}
	if (n.Type == html.TextNode) {
		res.bodyTextLen += utf8.RuneCountInString(n.Data) - strings.Count(n.Data, " ") - strings.Count(n.Data,
			newLine)
		return res
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res =  accumulateTreeInfo(res,parseBodyNode(c))
	}
	return res
}

func parseHeadNode(n *html.Node, titleNode bool) treeInfo {
	res := newTreeInfo()
	if (n.Type == html.TextNode && titleNode) {
		res.headTextLen += utf8.RuneCountInString(n.Data) - strings.Count(n.Data, " ") - strings.Count(n.Data,
			newLine)
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
		res =  accumulateTreeInfo(res, parseHeadNode(c, (n.Data == "title")))
	}
	return res
}

func parseRoot(n *html.Node, depth int) treeInfo {
	features := newTreeInfo()
	if (n.Type == html.CommentNode) {
		return features
	}

	if (n.Type == html.ElementNode) && (strings.ToLower(n.Data) == "head") {
		features = accumulateTreeInfo(features, parseHeadNode(n, false))
		return features
	}

	if (n.Type == html.ElementNode) && (strings.ToLower(n.Data) == "body") {
		features = accumulateTreeInfo(features, parseBodyNode(n))
		return features
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		tmp := parseRoot(c, depth+1)
		features = accumulateTreeInfo(features, tmp)
	}
	return features
}