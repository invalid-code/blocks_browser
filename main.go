package main

import (
	"fmt"
	"iter"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/html"
)

func convSeqToSlice[T any](seq iter.Seq[T]) []T {
	sliceToRet := []T{}
	for val := range seq {
		sliceToRet = append(sliceToRet, val)
	}
	return sliceToRet
}

type Queue[T any] struct {
	items []T
}

func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

func (q *Queue[T]) Dequeue() T {
	item := q.items[0]
	if len(q.items) == 1 {
		q.items = []T{}
	} else {
		q.items = q.items[1:]
	}
	return item
}

func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

func main() {
	a := app.New()
	w := a.NewWindow("Blocks Browser")
	w.Resize(fyne.NewSize(800, 600))

	content := container.NewVBox()
	searchBar := widget.NewEntry()
	searchBar.PlaceHolder = "Search the web"
	searchBar.OnSubmitted = func(s string) {
		if s != "" {
			resp, err := http.Get(fmt.Sprintf("https://%v/", s))
			if err != nil {
				fmt.Println(err)
				panic("todo")
			}
			defer resp.Body.Close()
			doc, err := html.Parse(resp.Body)
			if err != nil {
				fmt.Println(err)
				panic("todo")
			}
			domQueue := Queue[*html.Node]{items: []*html.Node{}}
			domQueue.Enqueue(doc)
			for !domQueue.IsEmpty() {
				curDomNode := domQueue.Dequeue()
				switch curDomNode.Type {
				case html.ErrorNode:
					fmt.Println("Error Node")
				case html.TextNode:
					fmt.Println(curDomNode.Data)
					continue
				case html.DocumentNode:
					fmt.Println("Document Node")
				case html.ElementNode:
					fmt.Println(curDomNode)
				case html.CommentNode:
					fmt.Println("Comment Node")
				case html.DoctypeNode:
					fmt.Println("DocType Node")
				case html.RawNode:
					fmt.Println("Raw Node")
				}
				curDomNodeChildren := convSeqToSlice(curDomNode.ChildNodes())
				if len(curDomNodeChildren) != 0 {
					for _, childNode := range curDomNodeChildren {
						domQueue.Enqueue(childNode)
					}
				}
			}
			fmt.Println("hi")
		}
	}
	refreshBtn := widget.NewButtonWithIcon("", resourceRefreshPng, func() {
		searchEntry := searchBar.Text
		if !(searchEntry == "") {
			resp, err := http.Get(fmt.Sprintf("https://%v/", searchEntry))
			if err != nil {
				fmt.Println(err)
				panic("todo")
			}
			defer resp.Body.Close()
		}
	})
	toolbar := container.NewHBox(refreshBtn, searchBar)
	w.SetContent(container.NewVBox(toolbar, content))
	w.ShowAndRun()
}
