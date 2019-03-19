package main

import (
	"strings"
	"golang.org/x/net/html"
	"unicode/utf8"
)

// Tree Parsing

func parseBodyNode(n *html.Node) treeInfo {
	res := newTreeInfo()
	if (n.Type != html.ElementNode) && (n.Type != html.TextNode) {
		return res
	}
	if (n.Type == html.TextNode) {
		res.bodyTextLen += utf8.RuneCountInString(n.Data) - strings.Count(n.Data, " ") - 2*strings.Count(n.Data,
			"\\n")
		return res
	}
	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "script":
			res.codeLen += countNodeLen(n)
		case "frame":
			res.frameCount += 1
		case "a":
			res.aTagCount += 1
			res.aTagLen += countNodeLen(n)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res =  accumulateTreeInfo(res,parseBodyNode(c))
	}
	return res
}

func parseHeadNode(n *html.Node, titleNode bool) treeInfo {
	res := newTreeInfo()
	if (n.Type == html.TextNode && titleNode) {
		res.headTextLen += utf8.RuneCountInString(n.Data) - strings.Count(n.Data, " ") - 2*strings.Count(n.Data,
			"\\n")
		return res
	}
	if (n.Type != html.ElementNode) && (n.Type != html.TextNode) {
		return res
	}

	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "script":
			res.codeLen += countNodeLen(n)
		case "frame":
			res.frameCount += 1
		case "a":
			res.aTagCount += 1
			res.aTagLen += countNodeLen(n)
		case "meta":
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

func countNodeLen(n *html.Node) int {
	len := 0
	for _,attr := range(n.Attr) {
		len += utf8.RuneCountInString(attr.Key) + utf8.RuneCountInString(attr.Val) + 1
	}

	len += utf8.RuneCountInString(n.Data)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		len += countNodeLen(c)
	}
	return len
}

