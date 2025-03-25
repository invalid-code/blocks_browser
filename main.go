package main

import (
	"fmt"
	"iter"
	"net/url"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/html"
)

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

func revSlice[T any](slice []T) []T {
	revedSlice := []T{}
	for i := len(slice) - 1; i >= 0; i-- {
		revedSlice = append(revedSlice, slice[i])
	}
	return revedSlice
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
	w.Resize(fyne.NewSize(800, 600))

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
	domStack := Stack[*html.Node]{items: []*html.Node{}}
	contentStack := Stack[*fyne.Container]{items: []*fyne.Container{}}
	domStack.Push(doc)
	contentStack.Push(content)
	isInline := false
	for !domStack.IsEmpty() {
		curDomNode := domStack.Pop()
		curContent := contentStack.Pop()
		switch curDomNode.Type {
		case html.DocumentNode:
			domStack.Push(revSlice(convSeqToSlice(curDomNode.ChildNodes()))...)
			contentStack.Push(curContent)
		case html.ElementNode:
			if curDomNode.Data == "head" {
				contentStack.Push(curContent)
				continue
			}
			if curDomNode.Data == "html" {
				contentStack.Push(curContent)
			} else if curDomNode.Data == "input" || curDomNode.Data == "button" || curDomNode.Data == "a" {
				var elemToIns fyne.Widget
				if curDomNode.Data == "input" {
					elemToIns = widget.NewEntry()
				} else if curDomNode.Data == "button" {
					buttonTxt := "Placeholder"
					if curDomNode.FirstChild != nil {
						buttonTxt = curDomNode.FirstChild.Data
					}
					elemToIns = widget.NewButton(buttonTxt, func() {})
				} else if curDomNode.Data == "a" {
					linkTxt := curDomNode.FirstChild.Data
					linkUrl, _ := url.Parse("#")
					if curDomNode.FirstChild != nil {
						linkUrlVal := ""
						for _, linkAttr := range curDomNode.FirstChild.Attr {
							if linkAttr.Key == "href" {
								linkUrlVal = linkAttr.Val
								break
							}
						}
						linkUrl, err = url.Parse(linkUrlVal)
						if err != nil {
							fmt.Println(err)
							panic("todo")
						}
					}
					elemToIns = widget.NewHyperlink(linkTxt, linkUrl)
				}
				if isInline {
					curContent.Add(elemToIns)
					contentStack.Push(curContent)
				} else {
					newContainer := container.NewHBox()
					newContainer.Add(elemToIns)
					curContent.Add(newContainer)
					contentStack.Push(curContent)
					contentStack.Push(newContainer)
				}
				isInline = true
				if curDomNode.Data == "button" || curDomNode.Data == "a" {
					continue
				}
			} else if curDomNode.Data == "ol" || curDomNode.Data == "ul" {
				if isInline {
					isInline = false
					curContent = contentStack.Pop()
				}
				curDomNodeChildren := convSeqToSlice(curDomNode.ChildNodes()) // all li elem
				listItems := []string{}
				for _, curDomNodeChild := range curDomNodeChildren {
					if strings.Trim(curDomNodeChild.Data, "\n ") == "" {
						continue
					}
					listItems = append(listItems, curDomNodeChild.FirstChild.Data)
				}
				newContainer := container.NewVBox()
				for i, listItem := range listItems {
					if curDomNode.Data == "ul" {
						newContainer.Add(widget.NewLabel(fmt.Sprintf("- %v", listItem)))
					} else {
						newContainer.Add(widget.NewLabel(fmt.Sprintf("%v. %v", i+1, listItem)))
					}
				}
				curContent.Add(newContainer)
				contentStack.Push(curContent)
				continue
			} else if curDomNode.Data == "table" {
				if isInline {
					isInline = false
					curContent = contentStack.Pop()
				}
				tableChildren := convSeqToSlice(curDomNode.ChildNodes())
				var rowChildren []*html.Node
				for _, tableChild := range tableChildren {
					if strings.Trim(tableChild.Data, "\n ") == "" {
						continue
					}
					rowChildren = convSeqToSlice(tableChild.ChildNodes())
				}
				tableData := [][]string{}
				i := 0
				for _, rowChild := range rowChildren {
					if strings.Trim(rowChild.Data, "\n ") == "" {
						continue
					}
					tableData = append(tableData, []string{})
					columnChildren := convSeqToSlice(rowChild.ChildNodes())
					for _, columnChild := range columnChildren {
						if strings.Trim(columnChild.Data, "\n ") == "" {
							continue
						}
						tableData[i] = append(tableData[i], columnChild.FirstChild.Data)
					}
					i++
				}
				tableWidget := container.NewVBox()
				for _, row := range tableData {
					rowCollection := container.NewHBox()
					for _, col := range row {
						rowCollection.Add(widget.NewLabel(col))
					}
					tableWidget.Add(rowCollection)
				}
				curContent.Add(tableWidget)
				contentStack.Push(curContent)
				continue
			} else {
				if isInline {
					isInline = false
					curContent = contentStack.Pop()
				}
				newContainer := container.NewVBox()
				curContent.Add(newContainer)
				contentStack.Push(newContainer)
			}
			domStack.Push(revSlice(convSeqToSlice(curDomNode.ChildNodes()))...)
		case html.TextNode:
			strText := strings.Trim(curDomNode.Data, "\n ")
			contentStack.Push(curContent)
			if strText == "" {
				continue
			}
			curContent.Add(widget.NewLabel(strText))
		default:
			contentStack.Push(curContent)
		}
	}
	content.Refresh()

	w.SetContent(content)
	w.ShowAndRun()
}
