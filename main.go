package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	WIDTH, HEIGHT int = 800, 600
)

type RenderTree struct {
	rootNode renderTreeNode
	degree   int
}

func (renderTree *RenderTree) printTree(degree int) {
	treeNodeStack := Stack[*renderTreeNode]{
		items: []*renderTreeNode{},
	}
	levelStack := Stack[int]{items: []int{}}
	treeNodeStack.Push(&renderTree.rootNode)
	levelStack.Push(0)
	if degree < 0 {
		degree = renderTree.degree
	}
	for !treeNodeStack.IsEmpty() {
		curTreeNode := treeNodeStack.Pop()
		curLevel := levelStack.Pop()
		curTreeNodeChildren := curTreeNode.children
		for i := 0; i < len(curTreeNodeChildren); i++ {
			levelStack.Push(curLevel + 1)
		}
		treeNodeStack.Push(revSlice(curTreeNodeChildren)...)
		if curLevel <= degree {
			for i := 0; i < curLevel; i++ {
				fmt.Printf("-")
			}
			if curTreeNode.isText {
				fmt.Printf("\"%v\"\n", curTreeNode.text)
			} else {
				fmt.Println(curTreeNode.elemType)
			}
		}
	}
}

func createRenderTree(doc *html.Node) (RenderTree, string, string) {
	var renderTree RenderTree
	renderTreeNodeQueue := Queue[*renderTreeNode]{items: []*renderTreeNode{}}
	domQueue := Queue[*html.Node]{items: []*html.Node{}}
	domQueue.Enqueue(doc)
	styles, scripts := "", ""
	for !domQueue.IsEmpty() {
		curDomNode := domQueue.Dequeue()
		curDomNodeChildren := convSeqToSlice(curDomNode.ChildNodes())
		switch curDomNode.Type {
		case html.ElementNode:
			switch curDomNode.DataAtom {
			case atom.Link:
				href := ""
				for _, attr := range curDomNode.Attr {
					if attr.Key == "href" {
						href = attr.Val
					}
				}
				fileContents, err := os.ReadFile(href)
				if err != nil {
					panic(err)
				}
				styles += string(fileContents)
				continue
			case atom.Script:
				src := ""
				for _, attr := range curDomNode.Attr {
					if attr.Key == "src" {
						src = attr.Val
					}
				}
				if src != "" {
					fileContents, err := os.ReadFile(src)
					if err != nil {
						panic(err)
					}
					scripts += string(fileContents)
					continue
				}
			case atom.Meta, atom.Style, atom.Title:
			case atom.Head:
				renderTreeNodeQueue.Dequeue()
			case atom.Html:
				renderTree.rootNode = renderTreeNode{
					children:     []*renderTreeNode{},
					isRoot:       true,
					elemTypeAtom: curDomNode.DataAtom,
					elemType:     "html",
					isText:       false,
					style:        map[string]string{},
					parent:       nil,
				}
				for i := 0; i < len(curDomNodeChildren); i++ {
					renderTreeNodeQueue.Enqueue(&renderTree.rootNode)
				}
			default:
				neededAttrName := ""
				neededAttr := ""
				for _, attr := range curDomNode.Attr {
					if attr.Key == "style" {
						styles += fmt.Sprintf("\n%v {\n%v\n}", curDomNode.Data, attr.Val)
					}
					if attr.Key == "src" || attr.Key == "href" {
						neededAttrName = attr.Key
						neededAttr = attr.Val
					}
				}
				parentRenderTreeNode := renderTreeNodeQueue.Dequeue()
				renderTreeNode := renderTreeNode{
					attr:         map[string]string{neededAttrName: neededAttr},
					children:     []*renderTreeNode{},
					elemTypeAtom: curDomNode.DataAtom,
					elemType:     curDomNode.Data,
					isRoot:       false,
					isText:       false,
					text:         "",
					parent:       parentRenderTreeNode,
					style:        map[string]string{},
				}
				parentRenderTreeNode.children = append(parentRenderTreeNode.children, &renderTreeNode)
				for i := 0; i < len(curDomNodeChildren); i++ {
					renderTreeNodeQueue.Enqueue(&renderTreeNode)
					if i == len(curDomNodeChildren)-1 {
						renderTree.degree += 1
					}
				}
			}
		case html.TextNode:
			text := strings.Trim(curDomNode.Data, "\n ")
			if text == "" {
				if curDomNode.Parent.DataAtom != atom.Head {
					renderTreeNodeQueue.Dequeue()
				}
				continue
			}
			switch curDomNode.Parent.DataAtom {
			case atom.Script:
				scripts += text
			case atom.Style:
				styles += text
			case atom.Title:
			default:
				parentRenderTreeNode := renderTreeNodeQueue.Dequeue()
				renderTreeNode := renderTreeNode{
					attr:         map[string]string{},
					children:     []*renderTreeNode{},
					isRoot:       false,
					isText:       true,
					elemTypeAtom: 0,
					elemType:     "",
					text:         text,
					style:        map[string]string{},
					parent:       parentRenderTreeNode,
				}
				parentRenderTreeNode.children = append(parentRenderTreeNode.children, &renderTreeNode)
				continue
			}
		}
		domQueue.Enqueue(curDomNodeChildren...)
	}
	return renderTree, styles, scripts
}

func (renderTree *RenderTree) parseCss(styles string) map[string]map[string]string {
	parser := css.NewParser(parse.NewInput(strings.NewReader(styles)), false)
	finished := false
	curRule := ""
	rules := map[string]map[string]string{}
	for !finished {
		ruleSet, _, data := parser.Next()
		switch ruleSet {
		case css.ErrorGrammar:
			if parser.Err().Error() == "EOF" {
				finished = true
			}
		case css.BeginRulesetGrammar:
			curRule = string(parser.Values()[0].Data)
			rules[curRule] = map[string]string{}
		case css.DeclarationGrammar:
			ruleVal := ""
			for _, ruleValPart := range parser.Values() {
				ruleVal += string(ruleValPart.Data)
			}
			rules[curRule][string(data)] = ruleVal
		}
	}
	return rules
}

func (renderTree *RenderTree) applyRules(rules map[string]map[string]string) {
	renderTreeNodeQueue := Queue[*renderTreeNode]{items: []*renderTreeNode{}}
	renderTreeNodeQueue.Enqueue(&renderTree.rootNode)
	for !renderTreeNodeQueue.IsEmpty() {
		curRenderTreeNode := renderTreeNodeQueue.Dequeue()
		if renderTreeNodeStyle, ok := rules[curRenderTreeNode.elemType]; ok {
			for key, val := range renderTreeNodeStyle {
				curRenderTreeNode.style[key] = val
			}
		}
		switch curRenderTreeNode.elemTypeAtom {
		case atom.Html:
			curRenderTreeNode.style["display"] = "block"
		case atom.Body:
			curRenderTreeNode.style["display"] = "block"
		case atom.Input:
			curRenderTreeNode.style["display"] = "inline"
		case atom.Div:
			curRenderTreeNode.style["display"] = "block"
		case atom.Button:
			curRenderTreeNode.style["display"] = "inline"
		case atom.A:
			curRenderTreeNode.style["display"] = "inline"
		case atom.Ul:
			curRenderTreeNode.style["display"] = "block"
		case atom.Ol:
			curRenderTreeNode.style["display"] = "block"
		case atom.Li:
			curRenderTreeNode.style["display"] = "block"
		case atom.Table:
			curRenderTreeNode.style["display"] = "block"
		case atom.Tbody:
			curRenderTreeNode.style["display"] = "block"
		case atom.Tr:
			curRenderTreeNode.style["display"] = "block"
		case atom.Td:
			curRenderTreeNode.style["display"] = "block"
		case atom.Span:
			curRenderTreeNode.style["display"] = "inline"
		case atom.Img:
			curRenderTreeNode.style["display"] = "inline"
		}
		renderTreeNodeQueue.Enqueue(curRenderTreeNode.children...)
	}
}

func (renderTree *RenderTree) layoutElements() *fyne.Container {
	isInline := false
	layoutTree, inlineContainer := container.NewVBox(), container.NewHBox()
	olIndex := 1
	renderTreeNodeQueue := Queue[*renderTreeNode]{items: []*renderTreeNode{}}
	contentQueue := Queue[*fyne.Container]{items: []*fyne.Container{}}
	renderTreeNodeQueue.Enqueue(&renderTree.rootNode)
	contentQueue.Enqueue(layoutTree)
	for !renderTreeNodeQueue.IsEmpty() {
		curRenderTreeNode := renderTreeNodeQueue.Dequeue()
		curContent := contentQueue.Dequeue()
		if curRenderTreeNode.isText {
			switch curRenderTreeNode.parent.elemTypeAtom {
			case atom.A:
				href, ok := curRenderTreeNode.parent.attr["href"]
				if !ok {
					panic("href attribute missing from a element")
				}
				link, err := url.Parse(href)
				if err != nil {
					panic(err)
				}
				curContent.Add(widget.NewHyperlink(curRenderTreeNode.text, link))
			case atom.Button:
				curContent.Add(widget.NewButton(curRenderTreeNode.text, func() {}))
			case atom.Li:
				switch curRenderTreeNode.parent.parent.elemTypeAtom {
				case atom.Ul:
					curContent.Add(
						widget.NewRichText(
							&widget.ListSegment{
								Items: []widget.RichTextSegment{
									&widget.TextSegment{
										Text:  curRenderTreeNode.text,
										Style: widget.RichTextStyle{},
									},
								},
								Ordered: false,
							},
						),
					)
				case atom.Ol:
					curContent.Add(
						widget.NewLabel(
							fmt.Sprintf("%v. %v", olIndex, curRenderTreeNode.text),
						),
					)
					olIndex += 1
				}
			case atom.Div, atom.Span, atom.Td:
				curContent.Add(widget.NewLabel(curRenderTreeNode.text))
			case atom.Body:
				textWidget := widget.NewLabel(curRenderTreeNode.text)
				if isInline {
					inlineContainer.Add(textWidget)
				} else {
					curContent.Add(textWidget)
				}
			}
			continue
		}
		switch curRenderTreeNode.style["display"] {
		case "block":
			if isInline {
				isInline = false
				inlineContainer = container.NewHBox()
			}
			switch curRenderTreeNode.elemTypeAtom {
			case atom.Html, atom.Body:
				for i := 0; i < len(curRenderTreeNode.children); i++ {
					contentQueue.Enqueue(curContent)
				}
			case atom.Tr:
				newContainer := container.NewHBox()
				for i := 0; i < len(curRenderTreeNode.children); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			case atom.Table:
				var newContainer *fyne.Container
				if _, ok := curRenderTreeNode.style["border"]; ok {
					newContainer = container.NewVBox()
				} else {
					newContainer = container.NewVBox()
				}
				for i := 0; i < len(curRenderTreeNode.children); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			default:
				newContainer := container.NewVBox()
				for i := 0; i < len(curRenderTreeNode.children); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			}
		case "inline":
			switch curRenderTreeNode.elemTypeAtom {
			case atom.Input:
				inlineContainer.Add(widget.NewEntry())
			case atom.Button, atom.A, atom.Span:
				newContainer := container.NewVBox()
				inlineContainer.Add(newContainer)
				contentQueue.Enqueue(newContainer)
			case atom.Img:
				src, ok := curRenderTreeNode.attr["src"]
				if !ok {
					panic("src attribute missing from img element")
				}
				img := canvas.NewImageFromFile(src)
				img.FillMode = canvas.ImageFillOriginal
				inlineContainer.Add(img)
			}
			if !isInline {
				isInline = true
				curContent.Add(inlineContainer)
			}
		}
		renderTreeNodeQueue.Enqueue(curRenderTreeNode.children...)
	}
	return layoutTree
}

type renderTreeNode struct {
	attr         map[string]string
	children     []*renderTreeNode
	elemTypeAtom atom.Atom
	elemType     string
	isRoot       bool
	isText       bool
	text         string
	parent       *renderTreeNode
	style        map[string]string
}

func main() {
	a := app.New()
	w := a.NewWindow("Blocks Browser")
	w.Resize(fyne.NewSize(float32(WIDTH), float32(HEIGHT)))

	file, err := os.Open("test.html")
	if err != nil {
		fmt.Println(err)
		panic("todo")
	}
	defer file.Close()
	doc, err := html.Parse(file)
	if err != nil {
		fmt.Println(err)
		panic("todo")
	}
	renderTree, styles, _ := createRenderTree(doc)
	rules := renderTree.parseCss(styles)
	renderTree.applyRules(rules)
	layoutTree := renderTree.layoutElements()

	height := 0
	for _, child := range layoutTree.Objects {
		height += int(child.MinSize().Height)
	}
	if height > HEIGHT {
		scrollableContent := container.NewVScroll(layoutTree)
		w.SetContent(scrollableContent)
	} else {
		w.SetContent(layoutTree)
	}
	w.ShowAndRun()
}
