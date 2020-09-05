/* 
MIT License

Copyright (c) 2020 Ashwin Godbole

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"strconv"
)

// States 
const (
	normal int = iota
	escape
	bold
	italics
	underline
	link
	alttext
	url
	image
	code
	heading
)

// map of names to characters, for flexibility
var blockChars = map[string]byte {
	"esc" : '\\',
	"escapeBegin" : '{',
	"escapeEnd" : '}',
	"bold" : '*',
	"italics" : '/',
	"underline" : '_',
	"line" : '-',
	"newline" : ';',
	"link" : '@',
	"image" : '!',
	"code" : '`',
	"heading" : '#',
}

// Stack to keep track of the open states
var openStates []int

func lastState() int {
	return openStates[len(openStates)-1]
}

func pushState(state int) {
	openStates = append(openStates, state)
}

func remLastState() {
	openStates = openStates[0:len(openStates)-1]
}

var ch byte
var source string
var current = 0
var peeked = 1
var hNum int

var preserveNewLines = false

func next() {
	current++
	peeked++
}

func curr() byte {
	return source[current]
}

func peek() byte {
	return source[peeked]
}

func peek2() byte {
	return source[peeked+1]
}

func compile(src string) string {
	source = src
	var compiled string 
	for ; current < len(source); next() {
		ch = curr()
		switch ch {
			case blockChars["esc"]:
				next()
				ch = curr()
				compiled += string(ch)

			case blockChars["escapeBegin"]:
				next()
				ch = curr()
				for ch != blockChars["escapeEnd"] && current < len(source) {
					if ch == blockChars["esc"] {
						next()
						ch = curr()	
					}
					compiled += string(ch)
					next()
					ch = curr()
					if ch == blockChars["escapeEnd"] {
						break
					}
				}
	
			case blockChars["bold"]:
				if len(openStates) > 0 && lastState() == bold {
					compiled += "</strong>"
					remLastState()
				} else {
					compiled += "<strong>"
					pushState(bold)
				}
		
			case blockChars["italics"]:
				if len(openStates) > 0 && lastState() == italics {
					compiled += "</em>"
					remLastState()
				} else {
					compiled += "<em>"
					pushState(italics)
				}
			
			case blockChars["underline"]:
				if len(openStates) > 0 && lastState() == underline {
					compiled += "</u>"
					remLastState()
				} else {
					compiled += "<u>"
					pushState(underline)
				}
			
			case blockChars["line"]:
				if peek() == blockChars["line"] && peek2() == blockChars["line"] {
					if len(openStates) == 0 {
						next()
						next()
						ch = curr()
						compiled += "<hr>"
					} else {
						compiled += string(ch)
					}
				} else {
						compiled += string(ch)
				}

			case blockChars["newline"]:
				if peek() == blockChars["newline"] {
					next()
					compiled += "<br>"
				} else {
					compiled += string(ch)
				}

			case blockChars["heading"]:
				hNum = 1
				next()
				ch = curr()
				for ch == '#' {
					hNum++
					next()
					ch = curr()
					if hNum == 6 {
						break
					}
				}
				current--
				peeked--
				compiled += "<h" + strconv.Itoa(hNum) + ">"
				pushState(heading)

			case '\n':
				if len(openStates) > 0 && lastState() == heading {
					remLastState()
					compiled += "</h" + strconv.Itoa(hNum) + ">"
					hNum = 0
				} else if preserveNewLines {
					compiled += "<br>"
				} else {
					compiled += string(ch)
				}
		
			case blockChars["link"]:
				next()
				ch = curr()
				if ch == '[' {
				compiled += "\n<a href=\""
				var url string
				var alt string
				next()
				ch = curr()
				for ch != ']' && current < len(source) {
					alt += string(ch)
					next()
					ch = curr()
				}
				next()
				ch = curr()
				if ch != '(' {
					return ""
				}
				next()
				ch = curr()
				paropen := 0
				for ch != ')' && current < len(source) {
					if ch == '(' {
							paropen++
				}
					url += string(ch)
					next()
					ch = curr()
					for ch == ')' && paropen != 0 {
						url += string(ch)
						next()
						ch = curr()
						paropen--
					}
				}
				next()
				ch = curr()
				compiled += url
				compiled += "\">"
				if alt== "" {
					compiled += url
				} else {
					compiled += alt
				}
				compiled += "</a>\n"
			}
			
			case blockChars["image"]: 
				next()
				ch = curr()
				if ch == '[' {
					compiled += "\n<img src=\""
					var url string
					var alt string 
					var w string
					var h string
					next()
					ch = curr()
					for ch != ']' && current < len(source) {
						if ch == ':' && peek() == ':' {
							next()
							next()
							ch = curr()
							for ch != ':' || peek() != ':' {
								w += string(ch)
								next()
								ch = curr()
							}
							next()
							next()
							ch = curr()
							for ch != ']' && current < len(source) {
								h += string(ch)
								next()
								ch = curr()
							}
							break
						}
						alt += string(ch)
						next()
						ch = curr()
					}
					next()
					ch = curr()
					if ch != '(' {
						return ""
					}
					next()
					ch = curr()
					paropen := 0
					for ch != ')' && current < len(source) {
						if ch == '(' {
							paropen++
						}
						url += string(ch)
						next()
						ch = curr()
						for ch == ')' && paropen != 0 {
							url += string(ch)
							next()
							ch = curr()
							paropen--
						}
					}
					next()
					ch = curr()
					compiled += url
					compiled += "\" alt=\""
					if alt== "" {
						compiled += url
					} else {
						compiled += alt
					}
					compiled += "\""
					if w!= "" {
						compiled += " width=\""
						compiled += w
						compiled += string('"')
					}
					if h!= "" {
						compiled += " height=\""
						compiled += h
						compiled += string('"')
					}
					compiled += ">\n"
				}			
			
			case blockChars["code"]:
				next()
				ch = curr()
				compiled += "<pre>"
				for ch != blockChars["code"] && current < len(source) {
					if ch == blockChars["esc"] {
						next()
						ch = curr()	
						compiled += string(ch)
						next()
						ch = curr()
					}
					compiled += string(ch)
					next()
					ch = curr()
					if ch == blockChars["code"] || current >= len(source) {
						break
					}
				}
				compiled += "</pre>"
			default:
				compiled += string(ch)
		}
	}
	return compiled
}

//export convert
func convert(input string) string {
	htmlBeg := `
<!doctype HTML>
<html>
<head>
<link rel="stylesheet" href="styles.css">
</head>
<body>
<div id="content">
`
	htmlEnd :=`
</div>
</body>
</html>
`	
	output := htmlBeg + compile(input) + htmlEnd
	return output
}

func main() {
		a := convert("*bold*")
}
