// original package: github.com/mdigger/goldmark-attributes
// modified for js output
// add img attributes
// also used as model for other extensions
// Package attributes is a extension for the goldmark
// (http://github.com/yuin/goldmark).
//
// This extension adds support for block attributes in markdowns.
//  paragraph text with attributes

package attributes

import (
	"fmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type imgAttrDelimiterProcessor struct {
}

func (p *imgAttrDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '{'
}

//func (p *strikethroughDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
//	return opener.Char == closer.Char
//}

func (p *strikethroughDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return ast.()
}

var defaultImgAttrDelimiterProcessor = &imgAttrDelimiterProcessor{}

type imgAttrParser struct {}

var defaultImgAttrParser = &imgAttrParser{}

// NewStrikethroughParser return a new InlineParser that parses
// strikethrough expressions.
func NewImgAttrParser() parser.InlineParser {
	return defaultParser
}

func (s *imgAttrParser) Trigger() []byte {
	return []byte{'{'}
}

func (s *imgAtttrParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultImgAttrDelimiterProcessor)
	if node == nil || node.OriginalLength > 2 || before == '{' {
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *strikethroughParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}



// block are parsed attributes block.
type inline struct {
	ast.BaseInLine
}

// Dump implements Node.Dump.
func (a *inline) Dump(source []byte, level int) {
fmt.Printf("dbg -- entering dump\n")
	attrs := a.Attributes()
	list := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		name := util.BytesToReadOnlyString(attr.Name)
		value := util.BytesToReadOnlyString(util.EscapeHTML(attr.Value.([]byte)))
		list[name] = value
	}

	ast.DumpHelper(a, source, level, list, nil)
}

// KindAttributes is a NodeKind of the attributes block node.
var ImgAttributes = ast.NewNodeKind("ImgAttributes")

// Kind implements Node.Kind.
func (a *inline) Kind() ast.NodeKind {
	return ImgAttributes
}

type attrParser struct{}

// Trigger implement parser.BlockParser interface.
func (a *attrParser) Trigger() []byte {
fmt.Println("dbg -- parser trigger")
	return []byte{'{'}
}

// Open implement parser.InlineParser interface.
func (a *attrParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
//fmt.Println("dbg --- found attributes")
	// add attributes if defined
	if attrs, ok := parser.ParseAttributes(reader); ok {
		node := &inline{BaseInline: ast.BaseBlock{}}
		for _, attr := range attrs {
			node.SetAttribute(attr.Name, attr.Value)
		}

		return node, parser.NoChildren
	}

	return nil, parser.RequireParagraph
}

// Continue implement parser.BlockParser interface.
func (a *attrParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

// Close implement parser.BlockParser interface.
func (a *attrParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

// CanInterruptParagraph implement parser.BlockParser interface.
func (a *attrParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine implement parser.BlockParser interface.
func (a *attrParser) CanAcceptIndentedLine() bool {
	return false
}

type transformer struct{}

// Transform implement parser.Transformer interface.
func (a *transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// collect all attributes block
	var attributes = make([]ast.Node, 0, 1000)
	_ = ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == KindAttributes {
			attributes = append(attributes, node)
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	// set attributes to next block sibling
//fmt.Printf("dbg -- attribute nodes: %d\n", len(attributes))
	for _, attr := range attributes {
		prev := attr.PreviousSibling()
		if prev != nil && prev.Type() == ast.TypeBlock &&
			!attr.HasBlankPreviousLines() {
			for _, attr := range attr.Attributes() {
				if _, exist := prev.Attribute(attr.Name); !exist {
					prev.SetAttribute(attr.Name, attr.Value)
//fmt.Printf("dbg -- attr: %s - %s\n", attr.Name, attr.Value)
				}
			}
		}

		// remove attributes node
		attr.Parent().RemoveChild(attr.Parent(), attr)
	}
}

type attrRender struct{}

// RegisterFuncs implement renderer.NodeRenderer interface.
func (a *attrRender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// not render
	reg.Register(KindAttributes,
		func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
			return ast.WalkSkipChildren, nil
		})
}

// extension defines a goldmark.Extender for markdown block attributes.
type imgAttrExt struct{}


func (e *imgAttrExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewImgAttrParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewImgAttrHtmlRenderer(), 500),
	))
}

// Extension is a goldmark.Extender with markdown block attributes support.
var ImgAttrExt goldmark.Extender = new(imgAttrExt)

// Enable is a goldmark.Option with block attributes support.
var Enable = goldmark.WithExtensions(Extension)
