package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// Embed represents an inline generic embed object, this should contain all the properties needed for the renderor to convert it into html
type Embed struct {
	gast.BaseInline

	EmbedData []EmbedData
}

//EmbedData contains data on a single option within an embed node
type EmbedData struct {
	URL      []byte
	MIMEType []byte
}

//String is a helper function to provide human-readable representation of this node. Used in Dump
func (ed *EmbedData) String() string {
	return "[\"" + string(ed.MIMEType) + "\", " + "\"" + string(ed.URL) + "\"]"
}

// Dump implements Node.Dump. This is purely a debugging function, when called prints human-readable output for troubleshooting
func (ge *Embed) Dump(source []byte, level int) {
	embedData := "["
	for _, v := range ge.EmbedData {
		embedData = embedData + v.String() + ", "
	}
	embedData = embedData + "]"

	m := map[string]string{
		"EmbedData": embedData,
	}
	gast.DumpHelper(ge, source, level, m, nil)
}

// KindEmbed is a NodeKind of the Embed node. This is used by goldmark to map AST nodes to renderers
var KindEmbed = gast.NewNodeKind("Embed")

// Kind implements Node.Kind, tells goldmark what kind of node it is dealing with. For this extension, it is always the Embed node
func (ge *Embed) Kind() gast.NodeKind {
	return KindEmbed
}

// NewEmbed returns a new Embed node.
func NewEmbed(EmbedData []EmbedData) *Embed {
	return &Embed{
		EmbedData: EmbedData,
	}
}
