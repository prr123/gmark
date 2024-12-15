// original package: github.com/mdigger/goldmark-attributes
// modified for js output
// add img attributes
// also used as model for other extensions
// Package attributes is a extension for the goldmark
// (http://github.com/yuin/goldmark).
//
// This extension adds support for block attributes in markdowns.
//  paragraph text with attributes

package imgAttributes

import (
	"fmt"
	"goDemo/gmark/goldmark"
	"goDemo/gmark/goldmark/ast"
	"goDemo/gmark/goldmark/parser"
	"goDemo/gmark/goldmark/renderer"
	"goDemo/gmark/goldmark/renderer/html"
	"goDemo/gmark/goldmark/text"
	"goDemo/gmark/goldmark/util"
)

// A Strikethrough struct represents a strikethrough of GFM text.
type ImgAttr struct {
    ast.BaseInline
}

// Dump implements Node.Dump.
func (a *ImgAttr) Dump(source []byte, level int) {
fmt.Printf("dbg dump -- entering dump\n")
	attrs := a.Attributes()
	list := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		name := util.BytesToReadOnlyString(attr.Name)
		value := util.BytesToReadOnlyString(util.EscapeHTML(attr.Value.([]byte)))
		list[name] = value
	}

	ast.DumpHelper(a, source, level, list, nil)
}


// KindStrikethrough is a NodeKind of the Strikethrough node.
var KindImgAttr = ast.NewNodeKind("ImgAttr")

// Kind implements Node.Kind.
func (n *ImgAttr) Kind() ast.NodeKind {
    return KindImgAttr
}

// NewStrikethrough returns a new Strikethrough node.
func NewImgAttr() *ImgAttr {
    return &ImgAttr{}
}

type imgAttrDelimiterProcessor struct {
}

func (p *imgAttrDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '{'
}

func (p *imgAttrDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return false
}

func (p *imgAttrDelimiterProcessor) OnMatch(consumes int) ast.Node {
	return NewImgAttr()
}

var defaultImgAttrDelimiterProcessor = &imgAttrDelimiterProcessor{}

type imgAttrParser struct {}

var defaultImgAttrParser = &imgAttrParser{}

// NewStrikethroughParser return a new InlineParser that parses
// imgAttr expressions.
func NewImgAttrParser() parser.InlineParser {
	return defaultImgAttrParser
}

func (s *imgAttrParser) Trigger() []byte {
	return []byte{'{'}
}

func (s *imgAttrParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
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


// transformer combines imgAttr node with img node
/*
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
*/


type ImgAttrHTMLRenderer struct {
    html.Config
}

// NewStrikethroughHTMLRenderer returns a new StrikethroughHTMLRenderer.
func NewImgAttrHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
    r := &ImgAttrHTMLRenderer{
        Config: html.NewConfig(),
    }
    for _, opt := range opts {
        opt.SetHTMLOption(&r.Config)
    }
    return r
}

// RegisterFuncs implement renderer.NodeRenderer interface.
func (a *ImgAttrHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// not render
	reg.Register(KindImgAttr,
		func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
			return ast.WalkSkipChildren, nil
		})
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
//func (r *StrikethroughHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
//    reg.Register(ast.KindStrikethrough, r.renderStrikethrough)
//}

// extension defines a goldmark.Extender for markdown block attributes.
type imgAttrExt struct{}


func (e *imgAttrExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewImgAttrParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewImgAttrHTMLRenderer(), 500),
	))
}

// Extension is a goldmark.Extender with markdown block attributes support.
var ImgAttrExt goldmark.Extender = new(imgAttrExt)

// Enable is a goldmark.Option with block attributes support.
var Enable = goldmark.WithExtensions(ImgAttrExt)
