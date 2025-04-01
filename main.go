package main

import (
	"fmt"
	"iter"

	// "net/url"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	// "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/widget"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	WIDTH, HEIGHT int = 800, 600
)

type CssOM struct {
	rootNode CssOMNode
	degree   int
}

func (cssOM *CssOM) printTree(degree int) {
	treeNodeStack := Stack[*CssOMNode]{items: []*CssOMNode{}}
	levelStack := Stack[int]{items: []int{}}
	treeNodeStack.Push(&cssOM.rootNode)
	levelStack.Push(0)
	if degree < 0 {
		degree = cssOM.degree
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
			fmt.Println(curTreeNode.elemType)
		}
	}
}

type CssOMNode struct {
	children []*CssOMNode
	isRoot   bool
	elemType string
	style    map[string]string
	parent   *CssOMNode
}

type Stack[T any] struct {
	items []T
}

func (stack *Stack[T]) Push(items ...T) {
	for _, item := range items {
		stack.items = append(stack.items, item)
	}
}

func (stack *Stack[T]) Pop() T {
	item := stack.items[len(stack.items)-1]
	stack.items = stack.items[:len(stack.items)-1]
	return item
}

func (stack *Stack[T]) IsEmpty() bool {
	return len(stack.items) == 0
}

type Queue[T any] struct {
	items []T
}

func (queue *Queue[T]) Enqueue(items ...T) {
	for _, item := range items {
		queue.items = append(queue.items, item)
	}
}

func (queue *Queue[T]) Dequeue() T {
	item := queue.items[0]
	if len(queue.items) > 1 {
		queue.items = queue.items[1:]
	} else {
		queue.items = []T{}
	}
	return item
}

func (queue *Queue[T]) IsEmpty() bool {
	return len(queue.items) == 0
}

func convSeqToSlice[T any](seq iter.Seq[T]) []T {
	sliceToRet := []T{}
	for item := range seq {
		sliceToRet = append(sliceToRet, item)
	}
	return sliceToRet
}

func revSlice[T any](sliceToRev []T) []T {
	revedSlice := []T{}
	for i := len(sliceToRev) - 1; i > -1; i-- {
		revedSlice = append(revedSlice, sliceToRev[i])
	}
	return revedSlice
}

func main() {
	a := app.New()
	w := a.NewWindow("Blocks Browser")
	w.Resize(fyne.NewSize(float32(WIDTH), float32(HEIGHT)))

	content := container.NewVBox()
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
	var cssOM CssOM
	cssOMNodeQueue := Queue[*CssOMNode]{items: []*CssOMNode{}}
	domQueue := Queue[*html.Node]{items: []*html.Node{}}
	domQueue.Enqueue(doc)
	// cssom creation
	styles := ""
	scripts := ""
	for !domQueue.IsEmpty() {
		curDomNode := domQueue.Dequeue()
		// fmt.Println(cssOMNodeQueue)
		// switch curDomNode.Type {
		// case html.ErrorNode:
		// 	fmt.Println("Error Node")
		// case html.TextNode:
		// 	fmt.Println("Text Node")
		// 	fmt.Printf("'%v'\n", curDomNode.Data)
		// case html.DocumentNode:
		// 	fmt.Println("Document Node")
		// case html.ElementNode:
		// 	fmt.Println("Element Node")
		// 	fmt.Println(curDomNode.Data)
		// case html.CommentNode:
		// 	fmt.Println("Comment Node")
		// case html.DoctypeNode:
		// 	fmt.Println("Doctype Node")
		// case html.RawNode:
		// 	fmt.Println("Raw Node")
		// }
		// fmt.Println()
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
				cssOMNodeQueue.Dequeue()
			case atom.Html:
				cssOM.rootNode = CssOMNode{children: []*CssOMNode{}, isRoot: true, elemType: curDomNode.Data, style: map[string]string{}, parent: nil}
				for i := 0; i < len(curDomNodeChildren); i++ {
					cssOMNodeQueue.Enqueue(&cssOM.rootNode)
				}
			default:
				for _, attr := range curDomNode.Attr {
					if attr.Key == "style" {
						styles += fmt.Sprintf("\n%v {\n%v\n}", curDomNode.Data, attr.Val)
					}
				}
				parentCssOMNode := cssOMNodeQueue.Dequeue()
				cssOMNode := CssOMNode{children: []*CssOMNode{}, isRoot: false, elemType: curDomNode.Data, style: map[string]string{}, parent: parentCssOMNode}
				parentCssOMNode.children = append(parentCssOMNode.children, &cssOMNode)
				for i := 0; i < len(curDomNodeChildren); i++ {
					cssOM.degree += 1
					cssOMNodeQueue.Enqueue(&cssOMNode)
				}
			}
		case html.TextNode:
			text := strings.Trim(curDomNode.Data, "\n ")
			if text == "" {
				if curDomNode.Parent.DataAtom != atom.Head {
					cssOMNodeQueue.Dequeue()
				}
				continue
			}
			switch curDomNode.Parent.DataAtom {
			case atom.Script:
				scripts += curDomNode.Data
				continue
			case atom.Style:
				styles += curDomNode.Data
				continue
			case atom.Title:
				continue
			}
			cssOMNodeQueue.Dequeue()
		}
		domQueue.Enqueue(curDomNodeChildren...)
	}
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
	cssOM.printTree(-1)

	// rendering
	// isInline := false
	// inlineContainer := container.NewHBox()
	// olIndex := 1
	// for !domQueue.IsEmpty() {
	// 	curDomNode := domQueue.Dequeue()
	// curContent := contentQueue.Dequeue()
	// curDomNodeChildren := convSeqToSlice(curDomNode.ChildNodes())
	// switch curDomNode.Type {
	// case html.ElementNode:
	// 	switch curDomNode.Data {
	// 	case "head":
	// 		continue
	// 	case "html":
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(curContent)
	// 		}
	// 	case "body":
	// 		newContainer := container.NewVBox()
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(newContainer)
	// 		}
	// 		curContent.Add(newContainer)
	// 	case "div":
	// 		if isInline {
	// 			isInline = false
	// 			inlineContainer = container.NewHBox()
	// 		}
	// 		newContainer := container.NewVBox()
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(newContainer)
	// 		}
	// 		curContent.Add(newContainer)
	// 	case "input":
	// 		inlineContainer.Add(widget.NewEntry())
	// 		if !isInline {
	// 			isInline = true
	// 			curContent.Add(inlineContainer)
	// 		}
	// 	case "button":
	// 		newContainer := container.NewVBox()
	// 		inlineContainer.Add(newContainer)
	// 		contentQueue.Enqueue(newContainer)
	// 		if !isInline {
	// 			isInline = true
	// 			curContent.Add(inlineContainer)
	// 		}
	// 	case "a":
	// 		newContainer := container.NewVBox()
	// 		inlineContainer.Add(newContainer)
	// 		contentQueue.Enqueue(newContainer)
	// 		if !isInline {
	// 			isInline = true
	// 			curContent.Add(inlineContainer)
	// 		}
	// 	case "ul":
	// 		if isInline {
	// 			isInline = false
	// 			inlineContainer = container.NewHBox()
	// 		}
	// 		newContainer := container.NewVBox()
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(newContainer)
	// 		}
	// 		curContent.Add(newContainer)
	// 	case "ol":
	// 		if isInline {
	// 			isInline = false
	// 			inlineContainer = container.NewHBox()
	// 		}
	// 		newContainer := container.NewVBox()
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(newContainer)
	// 		}
	// 		curContent.Add(newContainer)
	// 	case "li":
	// 		newContainer := container.NewVBox()
	// 		curContent.Add(newContainer)
	// 		contentQueue.Enqueue(newContainer)
	// 	case "table":
	// 		if isInline {
	// 			isInline = false
	// 			inlineContainer = container.NewHBox()
	// 		}
	// 		newContainer := container.NewVBox()
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(newContainer)
	// 		}
	// 		curContent.Add(newContainer)
	// 	case "tr":
	// 		newContainer := container.NewHBox()
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(newContainer)
	// 		}
	// 		curContent.Add(newContainer)
	// 	case "td":
	// 		newContainer := container.NewVBox()
	// 		curContent.Add(newContainer)
	// 		contentQueue.Enqueue(newContainer)
	// 	case "img":
	// 		imgSrc := ""
	// 		for _, attr := range curDomNode.Attr {
	// 			if attr.Key == "src" {
	// 				imgSrc = attr.Val
	// 			}
	// 		}
	// 		inlineContainer.Add(canvas.NewImageFromFile(imgSrc))
	// 		if !isInline {
	// 			isInline = true
	// 			curContent.Add(inlineContainer)
	// 		}
	// 	case "span":
	// 		newContainer := container.NewVBox()
	// 		inlineContainer.Add(newContainer)
	// 		contentQueue.Enqueue(newContainer)
	// 		if !isInline {
	// 			isInline = true
	// 			curContent.Add(inlineContainer)
	// 		}
	// 	case "tbody":
	// 		newContainer := container.NewVBox()
	// 		for i := 0; i < len(curDomNodeChildren); i++ {
	// 			contentQueue.Enqueue(newContainer)
	// 		}
	// 		curContent.Add(newContainer)
	// 	}
	// 	domQueue.Enqueue(curDomNodeChildren...)
	// case html.TextNode:
	// 	text := strings.Trim(curDomNode.Data, "\n ")
	// 	if text == "" {
	// 		continue
	// 	}
	// 	textWidget := widget.NewLabel(text)
	// 	switch curDomNode.Parent.Data {
	// 	case "a":
	// 		linkAttr := ""
	// 		for _, attr := range curDomNode.Parent.Attr {
	// 			if attr.Key == "href" {
	// 				linkAttr = attr.Val
	// 			}
	// 		}
	// 		link, err := url.Parse(linkAttr)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		curContent.Add(widget.NewHyperlink(text, link))
	// 	case "button":
	// 		curContent.Add(widget.NewButton(text, func() {}))
	// 	case "li":
	// 		switch curDomNode.Parent.Parent.Data {
	// 		case "ul":
	// 			curContent.Add(widget.NewLabel(fmt.Sprintf("- %v", text)))
	// 		case "ol":
	// 			curContent.Add(widget.NewLabel(fmt.Sprintf("%v. %v", olIndex, text)))
	// 			olIndex += 1
	// 		}
	// 	case "div":
	// 		curContent.Add(textWidget)
	// 	case "span":
	// 		curContent.Add(textWidget)
	// 	case "td":
	// 		curContent.Add(textWidget)
	// 	case "body":
	// 		if isInline {
	// 			inlineContainer.Add(textWidget)
	// 		} else {
	// 			curContent.Add(textWidget)
	// 		}
	// 	}
	// default:
	// 	for i := 0; i < len(curDomNodeChildren); i++ {
	// 		contentQueue.Enqueue(curContent)
	// 	}
	// 	domQueue.Enqueue(curDomNodeChildren...)
	// }
	// }
	height := 0
	for _, child := range content.Objects {
		height += int(child.MinSize().Height)
	}
	if height > HEIGHT {
		scrollableContent := container.NewVScroll(content)
		w.SetContent(scrollableContent)
	} else {
		w.SetContent(content)
	}
	w.ShowAndRun()
}
