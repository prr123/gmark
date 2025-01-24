package extDom

import (
	"bytes"
	"fmt"
	"strconv"

	"goDemo/gmark/goldmark"
	gast "goDemo/gmark/goldmark/ast"
	"goDemo/gmark/goldmark/extension/ast"
	"goDemo/gmark/goldmark/parser"
	"goDemo/gmark/goldmark/renderer"
	jsdom "goDemo/gmark/goldmark/renderer/jsdom"
	"goDemo/gmark/goldmark/text"
	"goDemo/gmark/goldmark/util"
)

var footnoteListKey = parser.NewContextKey()
var footnoteLinkListKey = parser.NewContextKey()

type footnoteBlockParser struct {
}

var defaultFootnoteBlockParser = &footnoteBlockParser{}

// NewFootnoteBlockParser returns a new parser.BlockParser that can parse
// footnotes of the Markdown(PHP Markdown Extra) text.
func NewFootnoteBlockParser() parser.BlockParser {
	return defaultFootnoteBlockParser
}

func (b *footnoteBlockParser) Trigger() []byte {
	return []byte{'['}
}

func (b *footnoteBlockParser) Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || line[pos] != '[' {
		return nil, parser.NoChildren
	}
	pos++
	if pos > len(line)-1 || line[pos] != '^' {
		return nil, parser.NoChildren
	}
	open := pos + 1
	var closes int
	closure := util.FindClosure(line[pos+1:], '[', ']', false, false) //nolint:staticcheck
	closes = pos + 1 + closure
	next := closes + 1
	if closure > -1 {
		if next >= len(line) || line[next] != ':' {
			return nil, parser.NoChildren
		}
	} else {
		return nil, parser.NoChildren
	}
	padding := segment.Padding
	label := reader.Value(text.NewSegment(segment.Start+open-padding, segment.Start+closes-padding))
	if util.IsBlank(label) {
		return nil, parser.NoChildren
	}
	item := ast.NewFootnote(label)

	pos = next + 1 - padding
	if pos >= len(line) {
		reader.Advance(pos)
		return item, parser.NoChildren
	}
	reader.AdvanceAndSetPadding(pos, padding)
	return item, parser.HasChildren
}

func (b *footnoteBlockParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if util.IsBlank(line) {
		return parser.Continue | parser.HasChildren
	}
	childpos, padding := util.IndentPosition(line, reader.LineOffset(), 4)
	if childpos < 0 {
		return parser.Close
	}
	reader.AdvanceAndSetPadding(childpos, padding)
	return parser.Continue | parser.HasChildren
}

func (b *footnoteBlockParser) Close(node gast.Node, reader text.Reader, pc parser.Context) {
	var list *ast.FootnoteList
	if tlist := pc.Get(footnoteListKey); tlist != nil {
		list = tlist.(*ast.FootnoteList)
	} else {
		list = ast.NewFootnoteList()
		pc.Set(footnoteListKey, list)
		node.Parent().InsertBefore(node.Parent(), node, list)
	}
	node.Parent().RemoveChild(node.Parent(), node)
	list.AppendChild(list, node)
}

func (b *footnoteBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *footnoteBlockParser) CanAcceptIndentedLine() bool {
	return false
}

type footnoteParser struct {
}

var defaultFootnoteParser = &footnoteParser{}

// NewFootnoteParser returns a new parser.InlineParser that can parse
// footnote links of the Markdown(PHP Markdown Extra) text.
func NewFootnoteParser() parser.InlineParser {
	return defaultFootnoteParser
}

func (s *footnoteParser) Trigger() []byte {
	// footnote syntax probably conflict with the image syntax.
	// So we need trigger this parser with '!'.
	return []byte{'!', '['}
}

func (s *footnoteParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, segment := block.PeekLine()
	pos := 1
	if len(line) > 0 && line[0] == '!' {
		pos++
	}
	if pos >= len(line) || line[pos] != '^' {
		return nil
	}
	pos++
	if pos >= len(line) {
		return nil
	}
	open := pos
	closure := util.FindClosure(line[pos:], '[', ']', false, false) //nolint:staticcheck
	if closure < 0 {
		return nil
	}
	closes := pos + closure
	value := block.Value(text.NewSegment(segment.Start+open, segment.Start+closes))
	block.Advance(closes + 1)

	var list *ast.FootnoteList
	if tlist := pc.Get(footnoteListKey); tlist != nil {
		list = tlist.(*ast.FootnoteList)
	}
	if list == nil {
		return nil
	}
	index := 0
	for def := list.FirstChild(); def != nil; def = def.NextSibling() {
		d := def.(*ast.Footnote)
		if bytes.Equal(d.Ref, value) {
			if d.Index < 0 {
				list.Count++
				d.Index = list.Count
			}
			index = d.Index
			break
		}
	}
	if index == 0 {
		return nil
	}

	fnlink := ast.NewFootnoteLink(index)
	var fnlist []*ast.FootnoteLink
	if tmp := pc.Get(footnoteLinkListKey); tmp != nil {
		fnlist = tmp.([]*ast.FootnoteLink)
	} else {
		fnlist = []*ast.FootnoteLink{}
		pc.Set(footnoteLinkListKey, fnlist)
	}
	pc.Set(footnoteLinkListKey, append(fnlist, fnlink))
	if line[0] == '!' {
		parent.AppendChild(parent, gast.NewTextSegment(text.NewSegment(segment.Start, segment.Start+1)))
	}

	return fnlink
}

type footnoteASTTransformer struct {
}

var defaultFootnoteASTTransformer = &footnoteASTTransformer{}

// NewFootnoteASTTransformer returns a new parser.ASTTransformer that
// insert a footnote list to the last of the document.
func NewFootnoteASTTransformer() parser.ASTTransformer {
	return defaultFootnoteASTTransformer
}

func (a *footnoteASTTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	var list *ast.FootnoteList
	var fnlist []*ast.FootnoteLink
	if tmp := pc.Get(footnoteListKey); tmp != nil {
		list = tmp.(*ast.FootnoteList)
	}
	if tmp := pc.Get(footnoteLinkListKey); tmp != nil {
		fnlist = tmp.([]*ast.FootnoteLink)
	}

	pc.Set(footnoteListKey, nil)
	pc.Set(footnoteLinkListKey, nil)

	if list == nil {
		return
	}

	counter := map[int]int{}
	if fnlist != nil {
		for _, fnlink := range fnlist {
			if fnlink.Index >= 0 {
				counter[fnlink.Index]++
			}
		}
		refCounter := map[int]int{}
		for _, fnlink := range fnlist {
			fnlink.RefCount = counter[fnlink.Index]
			if _, ok := refCounter[fnlink.Index]; !ok {
				refCounter[fnlink.Index] = 0
			}
			fnlink.RefIndex = refCounter[fnlink.Index]
			refCounter[fnlink.Index]++
		}
	}
	for footnote := list.FirstChild(); footnote != nil; {
		var container gast.Node = footnote
		next := footnote.NextSibling()
		if fc := container.LastChild(); fc != nil && gast.IsParagraph(fc) {
			container = fc
		}
		fn := footnote.(*ast.Footnote)
		index := fn.Index
		if index < 0 {
			list.RemoveChild(list, footnote)
		} else {
			refCount := counter[index]
			backLink := ast.NewFootnoteBacklink(index)
			backLink.RefCount = refCount
			backLink.RefIndex = 0
			container.AppendChild(container, backLink)
			if refCount > 1 {
				for i := 1; i < refCount; i++ {
					backLink := ast.NewFootnoteBacklink(index)
					backLink.RefCount = refCount
					backLink.RefIndex = i
					container.AppendChild(container, backLink)
				}
			}
		}
		footnote = next
	}
	list.SortChildren(func(n1, n2 gast.Node) int {
		if n1.(*ast.Footnote).Index < n2.(*ast.Footnote).Index {
			return -1
		}
		return 1
	})
	if list.Count <= 0 {
		list.Parent().RemoveChild(list.Parent(), list)
		return
	}

	node.AppendChild(node, list)
}


// FootnoteConfig holds configuration values for the footnote extension.
//
// Link* and Backlink* configurations have some variables:
// Occurrences of “^^” in the string will be replaced by the
// corresponding footnote number in the HTML output.
// Occurrences of “%%” will be replaced by a number for the
// reference (footnotes can have multiple references).
type FootnoteConfig struct {
	jsdom.Config

	// IDPrefix is a prefix for the id attributes generated by footnotes.
	IDPrefix []byte

	// IDPrefix is a function that determines the id attribute for given Node.
	IDPrefixFunction func(gast.Node) []byte

	// LinkTitle is an optional title attribute for footnote links.
	LinkTitle []byte

	// BacklinkTitle is an optional title attribute for footnote backlinks.
	BacklinkTitle []byte

	// LinkClass is a class for footnote links.
	LinkClass []byte

	// BacklinkClass is a class for footnote backlinks.
	BacklinkClass []byte

	// BacklinkHTML is an HTML content for footnote backlinks.
	BacklinkJsDOM []byte
}

// FootnoteOption interface is a functional option interface for the extension.
type FootnoteOption interface {
	renderer.Option
	// SetFootnoteOption sets given option to the extension.
	SetFootnoteOption(*FootnoteConfig)
}

// NewFootnoteConfig returns a new Config with defaults.
func NewFootnoteConfig() FootnoteConfig {
	return FootnoteConfig{
		Config:        jsdom.NewConfig(),
		LinkTitle:     []byte(""),
		BacklinkTitle: []byte(""),
		LinkClass:     []byte("footnote-ref"),
		BacklinkClass: []byte("footnote-backref"),
//		BacklinkHTML:  []byte("&#x21a9;&#xfe0e;"),
	}
}

// SetOption implements renderer.SetOptioner.
func (c *FootnoteConfig) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case optFootnoteIDPrefixFunction:
		c.IDPrefixFunction = value.(func(gast.Node) []byte)
	case optFootnoteIDPrefix:
		c.IDPrefix = value.([]byte)
	case optFootnoteLinkTitle:
		c.LinkTitle = value.([]byte)
	case optFootnoteBacklinkTitle:
		c.BacklinkTitle = value.([]byte)
	case optFootnoteLinkClass:
		c.LinkClass = value.([]byte)
	case optFootnoteBacklinkClass:
		c.BacklinkClass = value.([]byte)
//	case optFootnoteBacklinkHTML:
//		c.BacklinkHTML = value.([]byte)
	default:
		c.Config.SetOption(name, value)
	}
}

type withFootnoteJsDOMOptions struct {
	value []jsdom.Option
}

func (o *withFootnoteJsDOMOptions) SetConfig(c *renderer.Config) {
	if o.value != nil {
		for _, v := range o.value {
			v.(renderer.Option).SetConfig(c)
		}
	}
}

func (o *withFootnoteJsDOMOptions) SetFootnoteOption(c *FootnoteConfig) {
	if o.value != nil {
		for _, v := range o.value {
			v.SetJsDOMOption(&c.Config)
		}
	}
}

// WithFootnoteHTMLOptions is functional option that wraps goldmark HTMLRenderer options.
func WithFootnoteJsDOMOptions(opts ...jsdom.Option) FootnoteOption {
	return &withFootnoteJsDOMOptions{opts}
}

const optFootnoteIDPrefix renderer.OptionName = "FootnoteIDPrefix"

type withFootnoteIDPrefix struct {
	value []byte
}

func (o *withFootnoteIDPrefix) SetConfig(c *renderer.Config) {
	c.Options[optFootnoteIDPrefix] = o.value
}

func (o *withFootnoteIDPrefix) SetFootnoteOption(c *FootnoteConfig) {
	c.IDPrefix = o.value
}

// WithFootnoteIDPrefix is a functional option that is a prefix for the id attributes generated by footnotes.
func WithFootnoteIDPrefix[T []byte | string](a T) FootnoteOption {
	return &withFootnoteIDPrefix{[]byte(a)}
}

const optFootnoteIDPrefixFunction renderer.OptionName = "FootnoteIDPrefixFunction"

type withFootnoteIDPrefixFunction struct {
	value func(gast.Node) []byte
}

func (o *withFootnoteIDPrefixFunction) SetConfig(c *renderer.Config) {
	c.Options[optFootnoteIDPrefixFunction] = o.value
}

func (o *withFootnoteIDPrefixFunction) SetFootnoteOption(c *FootnoteConfig) {
	c.IDPrefixFunction = o.value
}

// WithFootnoteIDPrefixFunction is a functional option that is a prefix for the id attributes generated by footnotes.
func WithFootnoteIDPrefixFunction(a func(gast.Node) []byte) FootnoteOption {
	return &withFootnoteIDPrefixFunction{a}
}

const optFootnoteLinkTitle renderer.OptionName = "FootnoteLinkTitle"

type withFootnoteLinkTitle struct {
	value []byte
}

func (o *withFootnoteLinkTitle) SetConfig(c *renderer.Config) {
	c.Options[optFootnoteLinkTitle] = o.value
}

func (o *withFootnoteLinkTitle) SetFootnoteOption(c *FootnoteConfig) {
	c.LinkTitle = o.value
}

// WithFootnoteLinkTitle is a functional option that is an optional title attribute for footnote links.
func WithFootnoteLinkTitle[T []byte | string](a T) FootnoteOption {
	return &withFootnoteLinkTitle{[]byte(a)}
}

const optFootnoteBacklinkTitle renderer.OptionName = "FootnoteBacklinkTitle"

type withFootnoteBacklinkTitle struct {
	value []byte
}

func (o *withFootnoteBacklinkTitle) SetConfig(c *renderer.Config) {
	c.Options[optFootnoteBacklinkTitle] = o.value
}

func (o *withFootnoteBacklinkTitle) SetFootnoteOption(c *FootnoteConfig) {
	c.BacklinkTitle = o.value
}

// WithFootnoteBacklinkTitle is a functional option that is an optional title attribute for footnote backlinks.
func WithFootnoteBacklinkTitle[T []byte | string](a T) FootnoteOption {
	return &withFootnoteBacklinkTitle{[]byte(a)}
}

const optFootnoteLinkClass renderer.OptionName = "FootnoteLinkClass"

type withFootnoteLinkClass struct {
	value []byte
}

func (o *withFootnoteLinkClass) SetConfig(c *renderer.Config) {
	c.Options[optFootnoteLinkClass] = o.value
}

func (o *withFootnoteLinkClass) SetFootnoteOption(c *FootnoteConfig) {
	c.LinkClass = o.value
}

// WithFootnoteLinkClass is a functional option that is a class for footnote links.
func WithFootnoteLinkClass[T []byte | string](a T) FootnoteOption {
	return &withFootnoteLinkClass{[]byte(a)}
}

const optFootnoteBacklinkClass renderer.OptionName = "FootnoteBacklinkClass"

type withFootnoteBacklinkClass struct {
	value []byte
}

func (o *withFootnoteBacklinkClass) SetConfig(c *renderer.Config) {
	c.Options[optFootnoteBacklinkClass] = o.value
}

func (o *withFootnoteBacklinkClass) SetFootnoteOption(c *FootnoteConfig) {
	c.BacklinkClass = o.value
}

// WithFootnoteBacklinkClass is a functional option that is a class for footnote backlinks.
func WithFootnoteBacklinkClass[T []byte | string](a T) FootnoteOption {
	return &withFootnoteBacklinkClass{[]byte(a)}
}


const optFootnoteBacklinkJsDOM renderer.OptionName = "FootnoteBacklinkJsDOM"

type withFootnoteBacklinkJsDOM struct {
	value []byte
}

func (o *withFootnoteBacklinkJsDOM) SetConfig(c *renderer.Config) {
	c.Options[optFootnoteBacklinkJsDOM] = o.value
}

func (o *withFootnoteBacklinkJsDOM) SetFootnoteOption(c *FootnoteConfig) {
	c.BacklinkJsDOM = o.value
}

// WithFootnoteBacklinkHTML is an HTML content for footnote backlinks.
func WithFootnoteBacklinkJsDOM[T []byte | string](a T) FootnoteOption {
	return &withFootnoteBacklinkJsDOM{[]byte(a)}
}


// FootnoteHTMLRenderer is a renderer.NodeRenderer implementation that
// renders FootnoteLink nodes.
type FootnoteJsDOMRenderer struct {
	FootnoteConfig
	fnCount int
}

// NewFootnoteHTMLRenderer returns a new FootnoteHTMLRenderer.
func NewFootnoteJsDOMRenderer(opts ...FootnoteOption) renderer.NodeRenderer {
	r := &FootnoteJsDOMRenderer{
		FootnoteConfig: NewFootnoteConfig(),
	}
	for _, opt := range opts {
		opt.SetFootnoteOption(&r.FootnoteConfig)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *FootnoteJsDOMRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFootnoteLink, r.renderFootnoteLink)
	reg.Register(ast.KindFootnoteBacklink, r.renderFootnoteBacklink)
	reg.Register(ast.KindFootnote, r.renderFootnote)
	reg.Register(ast.KindFootnoteList, r.renderFootnoteList)
}

func (r *FootnoteJsDOMRenderer) renderFootnoteLink(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		n := node.(*ast.FootnoteLink)
fmt.Printf("dbg -- FN link Index: %d\n", n.Index)
		is := strconv.Itoa(n.Index)

//		_, _ = w.WriteString(`<sup id="`)
//		_, _ = w.Write(r.idPrefix(node))
//		_, _ = w.WriteString(`fnref`)
		if n.RefIndex > 0 {
		dbgStr := fmt.Sprintf("// dbg -- FN link RevIndex: %v\n", n.RefIndex)
_, _ = w.WriteString(dbgStr)
//			_, _ = w.WriteString(fmt.Sprintf("%v", n.RefIndex))
		}

		pnode := n.Parent()
        if pnode == nil {return gast.WalkStop, fmt.Errorf("no pnode")}
        parElNam, res := pnode.AttributeString("el")
        if !res {return gast.WalkStop, fmt.Errorf("fn: no parent el name: %s!", parElNam)}

		r.fnCount++
        fnNam := fmt.Sprintf("fn%d",r.fnCount)
        n.SetAttributeString("el",fnNam)
        fnStr := "let " + fnNam + "= document.createElement('sup');\n"
        _, _ = w.WriteString(fnStr)
		fnStr = fnNam + ".id='" + fnNam + "';\n"
        _, _ = w.WriteString(fnStr)

        fnStyl := "Object.assign(" + fnNam + ".style, mdStyle.footnote);\n"
        _, _ = w.WriteString(fnStyl)

        r.fnCount++
		refElNam := fmt.Sprintf("ref%d",r.fnCount)
		refElStr:="let " + refElNam + "= document.createElement('a');\n"
        _, _ = w.WriteString(refElStr)
		refStr := refElNam + ".href='fnref" + is + "';\n"
        _, _ = w.WriteString(refStr)

		fnStr = refElNam + ".textContent='" + is + "';\n"
        _, _ = w.WriteString(fnStr)

		apStr:= fnNam + ".appendChild(" + refElNam + ");\n"
        _, _ = w.WriteString(apStr)

		elapStr := parElNam.(string) + ".appendChild(" + fnNam + ");\n"
        _, _ = w.WriteString(elapStr)

//        if n.Attributes() != nil {jsdom.RenderElAttributes(w, n, TableAttributeFilter, tblNam)}
 
/*
		_ = w.WriteByte(':')
		_, _ = w.WriteString(is)
		_, _ = w.WriteString(`"><a href="#`)
		_, _ = w.Write(r.idPrefix(node))
		_, _ = w.WriteString(`fn:`)
		_, _ = w.WriteString(is)
		_, _ = w.WriteString(`" class="`)
		_, _ = w.Write(applyFootnoteTemplate(r.FootnoteConfig.LinkClass,
			n.Index, n.RefCount))
		if len(r.FootnoteConfig.LinkTitle) > 0 {
			_, _ = w.WriteString(`" title="`)
			_, _ = w.Write(util.EscapeHTML(applyFootnoteTemplate(r.FootnoteConfig.LinkTitle, n.Index, n.RefCount)))
		}
		_, _ = w.WriteString(`" role="doc-noteref">`)

		_, _ = w.WriteString(is)
		_, _ = w.WriteString(`</a></sup>`)
*/
	}
	return gast.WalkContinue, nil
}

func (r *FootnoteJsDOMRenderer) renderFootnoteBacklink(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {

	if entering {
fmt.Printf("dbg -- enter FootnoteBackLink\n")
		n := node.(*ast.FootnoteBacklink)
//		is := strconv.Itoa(n.Index)

//		_, _ = w.WriteString(`&#160;<a href="#`)
//		_, _ = w.Write(r.idPrefix(node))
//		_, _ = w.WriteString(`fnref`)
		if n.RefIndex > 0 {
//			_, _ = w.WriteString(fmt.Sprintf("%v", n.RefIndex))
		}
/*
		_ = w.WriteByte(':')
		_, _ = w.WriteString(is)
		_, _ = w.WriteString(`" class="`)
		_, _ = w.Write(applyFootnoteTemplate(r.FootnoteConfig.BacklinkClass, n.Index, n.RefCount))
		if len(r.FootnoteConfig.BacklinkTitle) > 0 {
			_, _ = w.WriteString(`" title="`)
			_, _ = w.Write(util.EscapeHTML(applyFootnoteTemplate(r.FootnoteConfig.BacklinkTitle, n.Index, n.RefCount)))
		}
		_, _ = w.WriteString(`" role="doc-backlink">`)
		_, _ = w.Write(applyFootnoteTemplate(r.FootnoteConfig.BacklinkHTML, n.Index, n.RefCount))
		_, _ = w.WriteString(`</a>`)
*/
	}
	return gast.WalkContinue, nil
}

func (r *FootnoteJsDOMRenderer) renderFootnote(
	w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {

	n := node.(*ast.Footnote)
	is := strconv.Itoa(n.Index)

fmt.Printf("dbg -- FN Nr: %d\n", n.Index)

	if entering {
		r.fnCount++
        liNam := fmt.Sprintf("fnli%d",r.fnCount)
        n.SetAttributeString("el",liNam)

		elStr := "let " + liNam + "=document.createElement('li');\n"
        _, _ = w.WriteString(elStr)
		elapStr:= liNam + ".id='fnref" + is + "';\n"
        _, _ = w.WriteString(elapStr)

        fnlStyl := "Object.assign(" + liNam + ".style, mdStyle.fnlistitem);\n"
        _, _ = w.WriteString(fnlStyl)

/*
		_, _ = w.WriteString(`<li id="`)
		_, _ = w.Write(r.idPrefix(node))
		_, _ = w.WriteString(`fn:`)
		_, _ = w.WriteString(is)
		_, _ = w.WriteString(`"`)
*/
	} else {
		pnode := n.Parent()
        if pnode == nil {return gast.WalkStop, fmt.Errorf("no pnode")}
//        parElNam, res := pnode.AttributeString("el")
//        if !res {return gast.WalkStop, fmt.Errorf("fn: no parent name: %s!", parElNam)}
        olNam, res := pnode.AttributeString("ol")
        if !res {return gast.WalkStop, fmt.Errorf("fn: no ol name: %s!", olNam)}
        liNam, res := n.AttributeString("el")
        if !res {return gast.WalkStop, fmt.Errorf("fn: li name: %s!", liNam)}
		elapStr := olNam.(string) + ".appendChild(" + liNam.(string) + ");\n"
        _, _ = w.WriteString(elapStr)
	}
	return gast.WalkContinue, nil
}

func (r *FootnoteJsDOMRenderer) renderFootnoteList(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {

	if entering {
fmt.Printf("dbg -- enter renderFNlist node: %s\n",node.Kind().String())
		n := node
		r.fnCount++
        fnlNam := fmt.Sprintf("fnl%d",r.fnCount)
        n.SetAttributeString("el",fnlNam)

        fnlStr := "let " + fnlNam + "= document.createElement('div');\n"
        _, _ = w.WriteString(fnlStr)
        fnlStyl := "Object.assign(" + fnlNam + ".style, mdStyle.fnlist);\n"
        _, _ = w.WriteString(fnlStyl)
		fnl2Str := fnlNam + ".id='FNList';\n"
        _, _ = w.WriteString(fnl2Str)

		r.fnCount++
        hrNam := fmt.Sprintf("hr%d",r.fnCount)
        hrStr := "let " + hrNam + "= document.createElement('hr');\n"
        _, _ = w.WriteString(hrStr)
		elapStr := fnlNam + ".appendChild(" +hrNam + ");\n" 
        _, _ = w.WriteString(elapStr)
		r.fnCount++
        olNam := fmt.Sprintf("ol%d",r.fnCount)
        n.SetAttributeString("ol",olNam)
        olStr := "let " + olNam + "= document.createElement('ol');\n"
        _, _ = w.WriteString(olStr)
	} else {
		pnode := node.Parent()
        if pnode == nil {return gast.WalkStop, fmt.Errorf("no pnode")}
        parElNam, res := pnode.AttributeString("el")
        if !res {return gast.WalkStop, fmt.Errorf("fn: no parent el name: %s!", parElNam.(string))}
        elNam, res := node.AttributeString("el")
        if !res {return gast.WalkStop, fmt.Errorf("fn: no el name: %s!", elNam.(string))}
        olNam, res := node.AttributeString("ol")
        if !res {return gast.WalkStop, fmt.Errorf("fn: no ol name: %s!", olNam.(string))}

		olapStr := elNam.(string) + ".appendChild(" + olNam.(string) + ");\n"
        _, _ = w.WriteString(olapStr)

		elapStr := parElNam.(string) + ".appendChild(" + elNam.(string) + ");\n"
        _, _ = w.WriteString(elapStr)
	}
	return gast.WalkContinue, nil
}

func (r *FootnoteJsDOMRenderer) idPrefix(node gast.Node) []byte {
	if r.FootnoteConfig.IDPrefix != nil {
		return r.FootnoteConfig.IDPrefix
	}
	if r.FootnoteConfig.IDPrefixFunction != nil {
		return r.FootnoteConfig.IDPrefixFunction(node)
	}
	return []byte("")
}

func applyFootnoteTemplate(b []byte, index, refCount int) []byte {
	fast := true
	for i, c := range b {
		if i != 0 {
			if b[i-1] == '^' && c == '^' {
				fast = false
				break
			}
			if b[i-1] == '%' && c == '%' {
				fast = false
				break
			}
		}
	}
	if fast {
		return b
	}
	is := []byte(strconv.Itoa(index))
	rs := []byte(strconv.Itoa(refCount))
	ret := bytes.Replace(b, []byte("^^"), is, -1)
	return bytes.Replace(ret, []byte("%%"), rs, -1)
}

type footnote struct {
	options []FootnoteOption
}

// Footnote is an extension that allow you to use PHP Markdown Extra Footnotes.
var Footnote = &footnote{
	options: []FootnoteOption{},
}

// NewFootnote returns a new extension with given options.
func NewFootnote(opts ...FootnoteOption) goldmark.Extender {
	return &footnote{
		options: opts,
	}
}

func (e *footnote) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewFootnoteBlockParser(), 999),
		),
		parser.WithInlineParsers(
			util.Prioritized(NewFootnoteParser(), 101),
		),
		parser.WithASTTransformers(
			util.Prioritized(NewFootnoteASTTransformer(), 999),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewFootnoteJsDOMRenderer(e.options...), 500),
	))
}
