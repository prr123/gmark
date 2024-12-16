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
	"goDemo/gmark/goldmark/text"
	"goDemo/gmark/goldmark/util"
)

// A Strikethrough struct represents a strikethrough of GFM text.
type ImgAttr struct {
    ast.BaseInline
}

// Dump implements Node.Dump.
func (a *ImgAttr) Dump(source []byte, level int) {
//fmt.Printf("dbg dump -- entering dump\n")
	attrs := a.Attributes()
	list := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		name := util.BytesToReadOnlyString(attr.Name)
		value := util.BytesToReadOnlyString(util.EscapeHTML(attr.Value.([]byte)))
		list[name] = value
	}

	ast.DumpHelper(a, source, level, list, nil)
}


// KindImgAttr is a NodeKind of the image attribute node.
var KindImgAttr = ast.NewNodeKind("ImgAttr")

// Kind implements Node.Kind.
func (n *ImgAttr) Kind() ast.NodeKind {
    return KindImgAttr
}

// NewStrikethrough returns a new Strikethrough node.
func NewImgAttr() *ImgAttr {
    return &ImgAttr{}
}

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

func (s *imgAttrParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) (node ast.Node) {

fmt.Printf("dbg Parse -- text: %s\n", block.Source())
	pos, Seg := block.Position()
fmt.Printf("pos: %d seg: %s\n",pos, block.Value(Seg) )

	attrs, ok := parser.ParseAttributes(block);
fmt.Printf("ok: %t\n", ok)

	if ok {
		// need to create a node
		node = &ImgAttr{BaseInline: ast.BaseInline{}}
        for _, attr := range attrs {
            node.SetAttribute(attr.Name, attr.Value)
		}
    }

	return node
}

// transformer combines imgAttr node with img node

type transformer struct{}

var defaultTransformer = &transformer{}
// Transform implement parser.Transformer interface.
func (a *transformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// collect all attributes nodes
	var attributes = make([]ast.Node, 0, 1000)
	_ = ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == KindImgAttr {
			attributes = append(attributes, node)
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	// set attributes to next block sibling
fmt.Printf("dbg -- attribute nodes: %d\n", len(attributes))
	for _, attr := range attributes {

		prev := attr.PreviousSibling()
		if prev != nil && prev.Kind() == ast.KindImage  {
			for _, attr := range attr.Attributes() {
				if _, exist := prev.Attribute(attr.Name); !exist {
					prev.SetAttribute(attr.Name, attr.Value)
fmt.Printf("dbg -- attr: %s - %v\n", attr.Name, attr.Value)
				}
			}
		}

		// remove attributes node
		attr.Parent().RemoveChild(attr.Parent(), attr)
	}
}


type ImgAttrHTMLRenderer struct {}
//    html.Config}

var imgAttrHTMLRenderer = &ImgAttrHTMLRenderer{}


// RegisterFuncs implement renderer.NodeRenderer interface.
func (a *ImgAttrHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// not render
	reg.Register(KindImgAttr,
		func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
			return ast.WalkSkipChildren, nil
		})
}

// extension defines a goldmark.Extender for markdown block attributes.
type imgAttrExt struct{}


func (e *imgAttrExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(util.Prioritized(NewImgAttrParser(), 500)),
		parser.WithASTTransformers(util.Prioritized(defaultTransformer, 500)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(imgAttrHTMLRenderer, 500)),
	)
}

// Extension is a goldmark.Extender with markdown block attributes support.
var ImgAttrExt goldmark.Extender = new(imgAttrExt)

// Enable is a goldmark.Option with block attributes support.
var Enable = goldmark.WithExtensions(ImgAttrExt)
