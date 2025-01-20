package supersubext

import (
	"goDemo/gmark/goldmark/supersubExt/ast"
	"goDemo/gmark/goldmark"
	gast "goDemo/gmark/goldmark/ast"
	"goDemo/gmark/goldmark/parser"
	"goDemo/gmark/goldmark/renderer"
	"goDemo/gmark/goldmark/renderer/html"
	"goDemo/gmark/goldmark/text"
	"goDemo/gmark/goldmark/util"
)

type supersubscriptDelimiterProcessor struct {
	super bool
}

func (p *supersubscriptDelimiterProcessor) IsDelimiter(b byte) bool {
	p.super =true
	if b== '~' {p.super = false}
	return b == '~' || b == '^'
}

func (p *supersubscriptDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *supersubscriptDelimiterProcessor) OnMatch(consumes int) gast.Node {
	n := ast.NewSuperSubScript()
//fmt.Printf("dbg -- onmatch: %t\n", p.super)
	n.Super = false
	if p.super {n.Super = true}
//	return ast.NewSuperSubScript()
	return n
}

var defaultSuperSubScriptDelimiterProcessor = &supersubscriptDelimiterProcessor{}

type supersubscriptParser struct {
	super bool
}

var defaultSuperSubScriptParser = &supersubscriptParser{}

// NewSubscriptParser return a new InlineParser that parses
// subscript expressions.
func NewSuperSubScriptParser() parser.InlineParser {
	return defaultSuperSubScriptParser
}

func (s *supersubscriptParser) Trigger() []byte {
	return []byte{'~','^'}
}

func (s *supersubscriptParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()

	line, segment := block.PeekLine()
//fmt.Printf("dbg -- line: %s\n", line)

	node := parser.ScanDelimiter(line, before, 1, defaultSuperSubScriptDelimiterProcessor)

	if node == nil {
		return nil
	}

	if node.CanOpen {
//fmt.Printf("dbg -- char: %q\n", line[0])
		s.super = false
		if line[0] == '^' {s.super = true}

		for i := 1; i < len(line); i++ {
			c := line[i]
			if c == line[0] {
				break
			}
			if util.IsSpace(c) {
				return nil
			}
		}
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)

	return node
}

func (s *supersubscriptParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// SubscriptHTMLRenderer is a renderer.NodeRenderer implementation that
// renders Subscript nodes.
type SuperSubScriptHTMLRenderer struct {
	html.Config
}

// NewSubscriptHTMLRenderer returns a new SubscriptHTMLRenderer.
func NewSuperSubScriptHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &SuperSubScriptHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *SuperSubScriptHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindSuperSubScript, r.renderSuperSubScript)
}

// SubscriptAttributeFilter defines attribute names which dd elements can have.
var SuperSubScriptAttributeFilter = html.GlobalAttributeFilter

func (r *SuperSubScriptHTMLRenderer) renderSuperSubScript(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	n := node.(*ast.SuperSubScript)
	if entering {
		if n.Attributes() != nil {
			if n.Super {
				_, _ = w.WriteString("<sup")
			} else {
				_, _ = w.WriteString("<sub")
			}
			html.RenderAttributes(w, n, SuperSubScriptAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			if n.Super {
				_, _ = w.WriteString("<sup>")
			} else {
				_, _ = w.WriteString("<sub>")
			}
		}
	} else {
		if n.Super {
			_, _ = w.WriteString("</sup>")
		} else {
			_, _ = w.WriteString("</sub>")
		}
	}
	return gast.WalkContinue, nil
}

type supersubscript struct {
}

// Subscript is an extension that allows you to use a subscript expression like 'x~0~'.
var SuperSubScript = &supersubscript{}

func (e *supersubscript) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewSuperSubScriptParser(), 600),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewSuperSubScriptHTMLRenderer(), 600),
	))
}
