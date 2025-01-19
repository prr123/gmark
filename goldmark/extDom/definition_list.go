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

type definitionListParser struct {
}

var defaultDefinitionListParser = &definitionListParser{}

// NewDefinitionListParser return a new parser.BlockParser that
// can parse PHP Markdown Extra Definition lists.
func NewDefinitionListParser() parser.BlockParser {
	return defaultDefinitionListParser
}

func (b *definitionListParser) Trigger() []byte {
	return []byte{':'}
}

func (b *definitionListParser) Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	if _, ok := parent.(*ast.DefinitionList); ok {
		return nil, parser.NoChildren
	}
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	indent := pc.BlockIndent()
	if pos < 0 || line[pos] != ':' || indent != 0 {
		return nil, parser.NoChildren
	}

	last := parent.LastChild()
	// need 1 or more spaces after ':'
	w, _ := util.IndentWidth(line[pos+1:], pos+1)
	if w < 1 {
		return nil, parser.NoChildren
	}
	if w >= 8 { // starts with indented code
		w = 5
	}
	w += pos + 1 /* 1 = ':' */

	para, lastIsParagraph := last.(*gast.Paragraph)
	var list *ast.DefinitionList
	status := parser.HasChildren
	var ok bool
	if lastIsParagraph {
		list, ok = last.PreviousSibling().(*ast.DefinitionList)
		if ok { // is not first item
			list.Offset = w
			list.TemporaryParagraph = para
		} else { // is first item
			list = ast.NewDefinitionList(w, para)
			status |= parser.RequireParagraph
		}
	} else if list, ok = last.(*ast.DefinitionList); ok { // multiple description
		list.Offset = w
		list.TemporaryParagraph = nil
	} else {
		return nil, parser.NoChildren
	}

	return list, status
}

func (b *definitionListParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if util.IsBlank(line) {
		return parser.Continue | parser.HasChildren
	}
	list, _ := node.(*ast.DefinitionList)
	w, _ := util.IndentWidth(line, reader.LineOffset())
	if w < list.Offset {
		return parser.Close
	}
	pos, padding := util.IndentPosition(line, reader.LineOffset(), list.Offset)
	reader.AdvanceAndSetPadding(pos, padding)
	return parser.Continue | parser.HasChildren
}

func (b *definitionListParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

func (b *definitionListParser) CanInterruptParagraph() bool {
	return true
}

func (b *definitionListParser) CanAcceptIndentedLine() bool {
	return false
}

type definitionDescriptionParser struct {
}

var defaultDefinitionDescriptionParser = &definitionDescriptionParser{}

// NewDefinitionDescriptionParser return a new parser.BlockParser that
// can parse definition description starts with ':'.
func NewDefinitionDescriptionParser() parser.BlockParser {
	return defaultDefinitionDescriptionParser
}

func (b *definitionDescriptionParser) Trigger() []byte {
	return []byte{':'}
}

func (b *definitionDescriptionParser) Open(
	parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	indent := pc.BlockIndent()
	if pos < 0 || line[pos] != ':' || indent != 0 {
		return nil, parser.NoChildren
	}
	list, _ := parent.(*ast.DefinitionList)
	if list == nil {
		return nil, parser.NoChildren
	}
	para := list.TemporaryParagraph
	list.TemporaryParagraph = nil
	if para != nil {
		lines := para.Lines()
		l := lines.Len()
		for i := 0; i < l; i++ {
			term := ast.NewDefinitionTerm()
			segment := lines.At(i)
			term.Lines().Append(segment.TrimRightSpace(reader.Source()))
			list.AppendChild(list, term)
		}
		para.Parent().RemoveChild(para.Parent(), para)
	}
	cpos, padding := util.IndentPosition(line[pos+1:], pos+1, list.Offset-pos-1)
	reader.AdvanceAndSetPadding(cpos+1, padding)

	return ast.NewDefinitionDescription(), parser.HasChildren
}

func (b *definitionDescriptionParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	// definitionListParser detects end of the description.
	// so this method will never be called.
	return parser.Continue | parser.HasChildren
}

func (b *definitionDescriptionParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {
	desc := node.(*ast.DefinitionDescription)
	desc.IsTight = !desc.HasBlankPreviousLines()
	if desc.IsTight {
		for gc := desc.FirstChild(); gc != nil; gc = gc.NextSibling() {
			paragraph, ok := gc.(*gast.Paragraph)
			if ok {
				textBlock := gast.NewTextBlock()
				textBlock.SetLines(paragraph.Lines())
				desc.ReplaceChild(desc, paragraph, textBlock)
			}
		}
	}
}

func (b *definitionDescriptionParser) CanInterruptParagraph() bool {
	return true
}

func (b *definitionDescriptionParser) CanAcceptIndentedLine() bool {
	return false
}

// DefinitionListHTMLRenderer is a renderer.NodeRenderer implementation that
// renders DefinitionList nodes.
type DefinitionListJsDOMRenderer struct {
	jsdom.Config
	dlcount int
	dtcount int
	dbg bool
}

// NewDefinitionListHTMLRenderer returns a new DefinitionListHTMLRenderer.
func NewDefinitionListJsDOMRenderer(opts ...jsdom.Option) renderer.NodeRenderer {
	r := &DefinitionListJsDOMRenderer{
		Config: jsdom.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetJsDOMOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *DefinitionListJsDOMRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindDefinitionList, r.renderDefinitionList)
	reg.Register(ast.KindDefinitionTerm, r.renderDefinitionTerm)
	reg.Register(ast.KindDefinitionDescription, r.renderDefinitionDescription)
}

// DefinitionListAttributeFilter defines attribute names which dl elements can have.
var DefinitionListAttributeFilter = jsdom.GlobalAttributeFilter

func (r *DefinitionListJsDOMRenderer) renderDefinitionList(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
fmt.Println("dbg -- entering def list!")
        pnode := n.Parent()
        if pnode == nil {return gast.WalkStop, fmt.Errorf("def list: no pnode")}
        elNam, res := pnode.AttributeString("el")
        if !res {return gast.WalkStop, fmt.Errorf("def list no par el name! %s", pnode.Kind().String())}

        if r.dbg {
            dbgStr := fmt.Sprintf("// dbg -- dl par: %s kind:%s\n", elNam.(string), pnode.Kind().String())
            _, _ = w.WriteString(dbgStr)
        }

		r.dlcount++
		dlNam := fmt.Sprintf("dl%d", r.dlcount)
        n.SetAttributeString("dl",dlNam)
        dlStr := "let " + dlNam + "= document.createElement('dl');\n"
        _, _ = w.WriteString(dlStr)
		//todo check that myStyle.do exists
		dlStyl := "if (mdStyle.hasOwnProperty('dlStyl')) {Object.assign(" + dlNam + ".style, mdStyle.dl);}\n"
        _, _ = w.WriteString(dlStyl)

		if n.Attributes() != nil {
//			_, _ = w.WriteString("<dl")
			jsdom.RenderElAttributes(w, n, DefinitionListAttributeFilter, dlNam)
		}

		dlaStr:=elNam.(string) + ".appendChild(" + dlNam + ");\n"
    	_, _ = w.WriteString(dlaStr)
	}

	return gast.WalkContinue, nil
}

// DefinitionTermAttributeFilter defines attribute names which dd elements can have.
var DefinitionTermAttributeFilter = jsdom.GlobalAttributeFilter

func (r *DefinitionListJsDOMRenderer) renderDefinitionTerm(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
fmt.Println("dbg -- entering def term!")
        pnode := n.Parent()
        if pnode == nil {return gast.WalkStop, fmt.Errorf("def term: no pnode")}
        dlNam, res := pnode.AttributeString("dl")
        if !res {return gast.WalkStop, fmt.Errorf("def term no dl name! %s", pnode.Kind().String())}

        if r.dbg {
            dbgStr := fmt.Sprintf("// dbg -- dl par: %s kind:%s\n", dlNam.(string), pnode.Kind().String())
            _, _ = w.WriteString(dbgStr)
        }

		r.dtcount++
		dtNam := fmt.Sprintf("dt%d", r.dtcount)
        n.SetAttributeString("dt",dtNam)
        dtStr := "let " + dtNam + "= document.createElement('dt');\n"
        _, _ = w.WriteString(dtStr)
		//todo check that myStyle.do exists
		dtStyl := "if (mdStyle.hasOwnProperty('dtStyl')) {Object.assign(" + dtNam + ".style, mdStyle.dt);}\n"
        _, _ = w.WriteString(dtStyl)

		if n.Attributes() != nil {
//			_, _ = w.WriteString("<dl")
			jsdom.RenderElAttributes(w, n, DefinitionListAttributeFilter, dtNam)
		}

		dtaStr:=dlNam.(string) + ".appendChild(" + dtNam + ");\n"
        _, _ = w.WriteString(dtaStr)
	}
	return gast.WalkContinue, nil
}

// DefinitionDescriptionAttributeFilter defines attribute names which dd elements can have.
var DefinitionDescriptionAttributeFilter = jsdom.GlobalAttributeFilter

func (r *DefinitionListJsDOMRenderer) renderDefinitionDescription(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {

	if entering {
fmt.Println("dbg -- entering def desc!")
        pnode := node.Parent()
        if pnode == nil {return gast.WalkStop, fmt.Errorf("def term: no pnode")}
        dtNam, res := pnode.AttributeString("dt")
        if !res {return gast.WalkStop, fmt.Errorf("def desc no par dt name! %s", pnode.Kind().String())}

        if r.dbg {
            dbgStr := fmt.Sprintf("// dbg -- dt par: %s kind:%s\n", dtNam.(string), pnode.Kind().String())
            _, _ = w.WriteString(dbgStr)
        }

//		_, _ = w.WriteString("<dd")
		n := node.(*ast.DefinitionDescription)
		r.dtcount++
		ddNam := fmt.Sprintf("dd%d", r.dtcount)
        n.SetAttributeString("dd",ddNam)
        ddStr := "let " + ddNam + "= document.createElement('dd');\n"
        _, _ = w.WriteString(ddStr)
		//todo check that myStyle.do exists
		ddStyl := "if (mdStyle.hasOwnProperty('ddStyl')) {Object.assign(" + ddNam + ".style, mdStyle.dd);}\n"
        _, _ = w.WriteString(ddStyl)

		if n.Attributes() != nil {
			jsdom.RenderElAttributes(w, n, DefinitionDescriptionAttributeFilter, ddNam)
		}

/*
		if n.IsTight {
			_, _ = w.WriteString(">")
		} else {
			_, _ = w.WriteString(">\n")
		}
*/
		ddaStr:=dtNam.(string) + ".appendChild(" + ddNam + ");\n"
        _, _ = w.WriteString(ddaStr)

	}
	return gast.WalkContinue, nil
}

type definitionList struct {
}

// DefinitionList is an extension that allow you to use PHP Markdown Extra Definition lists.
var DefinitionList = &definitionList{}

func (e *definitionList) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewDefinitionListParser(), 101),
		util.Prioritized(NewDefinitionDescriptionParser(), 102),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewDefinitionListJsDOMRenderer(), 500),
	))
}
