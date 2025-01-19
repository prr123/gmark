package extDom

import (
	"fmt"
	"goDemo/gmark/goldmark"
	gast "goDemo/gmark/goldmark/ast"
	"goDemo/gmark/goldmark/extension/ast"
	"goDemo/gmark/goldmark/parser"
	"goDemo/gmark/goldmark/renderer"
	jsdom "goDemo/gmark/goldmark/renderer/jsdom"
	"goDemo/gmark/goldmark/text"
	"goDemo/gmark/goldmark/util"
)

type strikethroughDelimiterProcessor struct {
}

func (p *strikethroughDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '~'
}

func (p *strikethroughDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *strikethroughDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return ast.NewStrikethrough()
}

var defaultStrikethroughDelimiterProcessor = &strikethroughDelimiterProcessor{}

type strikethroughParser struct {
}

var defaultStrikethroughParser = &strikethroughParser{}

// NewStrikethroughParser return a new InlineParser that parses
// strikethrough expressions.
func NewStrikethroughParser() parser.InlineParser {
	return defaultStrikethroughParser
}

func (s *strikethroughParser) Trigger() []byte {
	return []byte{'~'}
}

func (s *strikethroughParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultStrikethroughDelimiterProcessor)
	if node == nil || node.OriginalLength > 2 || before == '~' {
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

// StrikethroughJsDOMRenderer is a renderer.NodeRenderer implementation that
// renders Strikethrough nodes.
type StrikethroughJsDOMRenderer struct {
	jsdom.Config
	scount int
	dbg bool
}

// NewStrikethroughHTMLRenderer returns a new StrikethroughHTMLRenderer.
func NewStrikethroughJsDOMRenderer(opts ...jsdom.Option) renderer.NodeRenderer {
	r := &StrikethroughJsDOMRenderer{
		Config: jsdom.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetJsDOMOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *StrikethroughJsDOMRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindStrikethrough, r.renderStrikethrough)
}

// StrikethroughAttributeFilter defines attribute names which dd elements can have.
var StrikethroughAttributeFilter = jsdom.GlobalAttributeFilter

func (r *StrikethroughJsDOMRenderer) renderStrikethrough(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
        r.scount++
        elNam := fmt.Sprintf("st%d",r.scount)
        n.SetAttributeString("st",elNam)
        elStr:= "let " + elNam + "=document.createElement('del');\n"
        _, _ = w.WriteString(elStr)
        if n.Attributes() != nil {jsdom.RenderElAttributes(w, n, jsdom.EmphasisAttributeFilter, elNam)}
        // child
		// examine the case of emphasis as child node!!
        chn := n.FirstChild()
        if _, ok := chn.(*gast.Text); ok {
            segment := chn.(*gast.Text).Segment
            value := segment.Value(source)
            chStr := elNam + ".textContent=`" + string(value) + "`\n";
            _, _ = w.WriteString(chStr)
        }
        return gast.WalkSkipChildren, nil
 	}
    pnode := n.Parent()
    if pnode == nil {return gast.WalkStop, fmt.Errorf("Emphasis -- no pnode")}
    elNam, res := n.AttributeString("st")
    if !res {return gast.WalkStop, fmt.Errorf("del --no el name!")}
    parElNam, res := pnode.AttributeString("el")
    if !res {return gast.WalkStop, fmt.Errorf("Emphasis -- no parent el name: %s!", elNam)}
    if r.dbg {
        dbgStr := fmt.Sprintf("// dbg -- el: %s parent:%s kind:%s\n", elNam, parElNam, pnode.Kind().String())
        _, _ = w.WriteString(dbgStr)
    }
    elStr := parElNam.(string) + ".appendChild(" + elNam.(string) + ");\n"
    _, _ = w.WriteString(elStr)

	return gast.WalkContinue, nil
}

type strikethrough struct {
}

// Strikethrough is an extension that allow you to use strikethrough expression like '~~text~~' .
var Strikethrough = &strikethrough{}

func (e *strikethrough) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewStrikethroughParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewStrikethroughJsDOMRenderer(), 500),
	))
}
