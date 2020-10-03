//Package embed is an extension based on goldmark-emoji(http://github.com/yuin/goldmark-emoji) for goldmark(http://github.com/yuin/goldmark)
package embed

import (
	"fmt"
	"html"

	geast "goldmark-embed/ast"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type embedParser struct {
}

var defaultEmbedParser = &embedParser{}

// NewParser returns a new parser.InlineParser that can parse embed
func NewParser() parser.InlineParser {
	return defaultEmbedParser
}

//Trigger returns the the required bytes to trigger this parser
func (s *embedParser) Trigger() []byte {
	return []byte{'!'} //[](
}

//Parse is called when goldmark detects this objects Trigger()
func (s *embedParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	if len(line) < 5 { //Not even long enough for ![]()
		return nil
	}
	if line[1] != '[' || line[2] != ']' || line[3] != '(' { //Verify beginning ![](
		return nil
	}

	i := 4 //Start after the first parenthesis ![](
	//Then parse the line until we either give up due to improper format, or when we have the data needed for the node
	//Format of this node should be "<mimetype> <url> <mimetype> <url> <mimetype> <url>", at least once and repeated any number of times optionally quoted around mimetype and url
	var embedData []geast.EmbedData
	currentEmbed := geast.EmbedData{}
	currentData := []byte{}
	inQuote := false
	inMime := true
	for ; i < len(line); i++ {
		c := line[i]
		//These characters are invalid. If we run into these, end parsing
		if c == '\r' || c == '\n' {
			//Invalid node
			return nil
		}

		//If we run into ) while not in a quoted string, then we can end cleanly/valid node
		if !inQuote && c == ')' {
			if len(currentData) > 0 && !inMime {
				//Clean end, but we have to save final node
				//Otherwise we are in url, and we should add the completed embed to the embed slice
				currentEmbed.URL = currentData
				embedData = append(embedData, currentEmbed)
				//Reset for next entry
				currentEmbed = geast.EmbedData{}
				currentData = []byte{}
				inMime = true
				break
			}
			if len(currentData) == 0 && inMime {
				//Clean end
				break
			} else {
				return nil //Invalid node as we ended with an incomplete mime/url pair
			}
		}

		//Skip whitespace while we are not in quote and at the start of a new data entry. This makes ![](     mime url) valid
		if (c == ' ' || c == '\t') && !inQuote && len(currentData) == 0 {
			continue //Skip this character
		}
		//However, if we already have some data, not in a quote, and we hit whitespace, then we should end the current item
		if (c == ' ' || c == '\t') && !inQuote {
			if inMime {
				inMime = false //Next element should be url
				//Switch to parsing URL
				currentEmbed.MIMEType = currentData
				currentData = []byte{}
			} else {
				//Otherwise we are in url, and we should add the completed embed to the embed slice
				currentEmbed.URL = currentData
				embedData = append(embedData, currentEmbed)
				//Reset for next entry
				currentEmbed = geast.EmbedData{}
				currentData = []byte{}
				inMime = true
			}

			continue //Skip the rest of the processing for this character
		}

		//Handle quoting
		if c == '"' || c == '\'' {
			if inQuote {
				inQuote = false //End quote, add the current item or fail
				if len(currentData) == 0 {
					return nil //Invalid data, quoted string was empty
				}
				if inMime {
					inMime = false //Next element should be url
					//Switch to parsing URL
					currentEmbed.MIMEType = currentData
					currentData = []byte{}
				} else {
					//Otherwise we are in url, and we should add the completed embed to the embed slice
					currentEmbed.URL = currentData
					embedData = append(embedData, currentEmbed)
					//Reset for next entry
					currentEmbed = geast.EmbedData{}
					currentData = []byte{}
					inMime = true
				}
			} else {
				if len(currentData) == 0 {
					inQuote = true //We were not already in a quote, so enter a quote
				} else {
					return nil //Fail out as we are trying to start a quote in the middle of value
				}
			}
			continue //Skip processing for this character
		}

		//Add data to current item
		currentData = append(currentData, c)
	}
	//Verify we have ended cleanly by checking the end state
	if !inMime || len(currentData) != 0 {
		return nil
	}

	//Verify our embed has data
	if len(embedData) == 0 {
		return nil
	}
	//Advance reader
	block.Advance(i + 1)

	//Return new ast node
	return geast.NewEmbed(embedData)
}

type embedHTMLRenderer struct {
}

// NewHTMLRenderer returns a new embedHTMLRenderer.
func NewHTMLRenderer() renderer.NodeRenderer {
	return &embedHTMLRenderer{}
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs. These provide goldmark with the necessary function hook for rendering our custom AST node
func (r *embedHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(geast.KindEmbed, r.renderEmbed)
}

//renderEmbed is called by gomark to actually render our custom AST node. Goldmark is aware of this function through RegisterFuncs above
func (r *embedHTMLRenderer) renderEmbed(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {

	if !entering {
		return ast.WalkContinue, nil
	}
	node := n.(*geast.Embed)
	//Generte html output
	fmt.Fprintf(w, "<video controls>")
	for _, v := range node.EmbedData {
		fmt.Fprintf(w, "<source src=\"%s\" type=\"%s\">", html.EscapeString(string(v.URL)), html.EscapeString(string(v.MIMEType)))
	}
	fmt.Fprintf(w, "</video>")
	return ast.WalkContinue, nil
}

//embed is a generic embed goldmark extender, this just acts as a single entry point for adding our renderor, AST node, and parser to goldmark
type embed struct {
}

// DefaultEmbed is the default embed extender
var DefaultEmbed = &embed{}

// New should be used to return a new extension with given options, but right now, we just return the default as we currently do not support any additional options
func New() goldmark.Extender {
	return DefaultEmbed
}

// Extend implements goldmark.Extender, this actually adds our extension to goldmark, both the parser and the renderer
func (e *embed) Extend(m goldmark.Markdown) {

	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewParser(), 200),
	))

	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHTMLRenderer(), 200),
	))
}
