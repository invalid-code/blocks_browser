package main

import (
	"fmt"
	"iter"
	"net/url"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/html"
)

const (
	WIDTH, HEIGHT int = 800, 600
)

type CssOMNode struct {
	elemType string
	style    map[string]string
	parent   *CssOMNode
}

type CssOM struct {
	nodes []CssOMNode
}

type RenderTreeNode struct{}

type RenderTree struct {
	nodes []RenderTreeNode
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
	domQueue := Queue[*html.Node]{items: []*html.Node{}}
	contentQueue := Queue[*fyne.Container]{items: []*fyne.Container{}}
	domQueue.Enqueue(doc)
	contentQueue.Enqueue(content)
	isInline := false
	inlineContainer := container.NewHBox()
	olIndex := 1
	for !domQueue.IsEmpty() {
		curDomNode := domQueue.Dequeue()
		curContent := contentQueue.Dequeue()
		curDomNodeChildren := convSeqToSlice(curDomNode.ChildNodes())
		switch curDomNode.Type {
		case html.ElementNode:
			switch curDomNode.Data {
			case "head":
				continue
			case "html":
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(curContent)
				}
			case "body":
				newContainer := container.NewVBox()
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			case "div":
				if isInline {
					isInline = false
					inlineContainer = container.NewHBox()
				}
				newContainer := container.NewVBox()
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			case "input":
				inlineContainer.Add(widget.NewEntry())
				if !isInline {
					isInline = true
					curContent.Add(inlineContainer)
				}
			case "button":
				newContainer := container.NewVBox()
				inlineContainer.Add(newContainer)
				contentQueue.Enqueue(newContainer)
				if !isInline {
					isInline = true
					curContent.Add(inlineContainer)
				}
			case "a":
				newContainer := container.NewVBox()
				inlineContainer.Add(newContainer)
				contentQueue.Enqueue(newContainer)
				if !isInline {
					isInline = true
					curContent.Add(inlineContainer)
				}
			case "ul":
				if isInline {
					isInline = false
					inlineContainer = container.NewHBox()
				}
				newContainer := container.NewVBox()
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			case "ol":
				if isInline {
					isInline = false
					inlineContainer = container.NewHBox()
				}
				newContainer := container.NewVBox()
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			case "li":
				newContainer := container.NewVBox()
				curContent.Add(newContainer)
				contentQueue.Enqueue(newContainer)
			case "table":
				if isInline {
					isInline = false
					inlineContainer = container.NewHBox()
				}
				newContainer := container.NewVBox()
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			case "tr":
				newContainer := container.NewHBox()
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			case "td":
				newContainer := container.NewVBox()
				curContent.Add(newContainer)
				contentQueue.Enqueue(newContainer)
			case "img":
				imgSrc := ""
				for _, attr := range curDomNode.Attr {
					if attr.Key == "src" {
						imgSrc = attr.Val
					}
				}
				inlineContainer.Add(canvas.NewImageFromFile(imgSrc))
				if !isInline {
					isInline = true
					curContent.Add(inlineContainer)
				}
			case "span":
				newContainer := container.NewVBox()
				inlineContainer.Add(newContainer)
				contentQueue.Enqueue(newContainer)
				if !isInline {
					isInline = true
					curContent.Add(inlineContainer)
				}
			case "tbody":
				newContainer := container.NewVBox()
				for i := 0; i < len(curDomNodeChildren); i++ {
					contentQueue.Enqueue(newContainer)
				}
				curContent.Add(newContainer)
			}
			domQueue.Enqueue(curDomNodeChildren...)
		case html.TextNode:
			text := strings.Trim(curDomNode.Data, "\n ")
			if text == "" {
				continue
			}
			textWidget := widget.NewLabel(text)
			switch curDomNode.Parent.Data {
			case "a":
				linkAttr := ""
				for _, attr := range curDomNode.Parent.Attr {
					if attr.Key == "href" {
						linkAttr = attr.Val
					}
				}
				link, err := url.Parse(linkAttr)
				if err != nil {
					panic(err)
				}
				curContent.Add(widget.NewHyperlink(text, link))
			case "button":
				curContent.Add(widget.NewButton(text, func() {}))
			case "li":
				switch curDomNode.Parent.Parent.Data {
				case "ul":
					curContent.Add(widget.NewLabel(fmt.Sprintf("- %v", text)))
				case "ol":
					curContent.Add(widget.NewLabel(fmt.Sprintf("%v. %v", olIndex, text)))
					olIndex += 1
				}
			case "div":
				curContent.Add(textWidget)
			case "span":
				curContent.Add(textWidget)
			case "td":
				curContent.Add(textWidget)
			case "body":
				if isInline {
					inlineContainer.Add(textWidget)
				} else {
					curContent.Add(textWidget)
				}
			}
		default:
			for i := 0; i < len(curDomNodeChildren); i++ {
				contentQueue.Enqueue(curContent)
			}
			domQueue.Enqueue(curDomNodeChildren...)
		}
	}
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
