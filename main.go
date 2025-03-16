package main

import (
	"fmt"
	"iter"
	"net/http"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/html"
)

type ToolbarLayout struct{}

func (toolbar *ToolbarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for _, obj := range objects {
		objMinSize := obj.MinSize()
		objSize := obj.Size()
		if objMinSize.Height > h {
			h = objMinSize.Height
		}
		if objSize.Height == 0 && objSize.Width == 0 {
			w += objMinSize.Width
		} else {
			w += objSize.Width
		}
	}
	return fyne.NewSize(w, h)
}

func (toolbar *ToolbarLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	for _, obj := range objects {
		objMinSize := obj.MinSize()
		objSize := obj.Size()
		obj.Move(pos)
		if objSize.Height == 0 && objSize.Width == 0 {
			obj.Resize(fyne.NewSize(objMinSize.Width, objMinSize.Height))
			pos.X += objMinSize.Width + 1.0
		} else {
			obj.Resize(fyne.NewSize(objSize.Width, objMinSize.Height))
			pos.X += objSize.Width + 1.0
		}
	}
}

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

func (q *Queue[T]) Enqueue(items ...T) {
	for _, item := range items {
		q.items = append(q.items, item)
	}
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
	searchBar.Resize(fyne.NewSize(800, 0))
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
			contentQueue := Queue[*fyne.Container]{items: []*fyne.Container{}}
			domQueue.Enqueue(doc)
			contentQueue.Enqueue(content)
			for !domQueue.IsEmpty() {
				curDomNode := domQueue.Dequeue()
				curContent := contentQueue.Dequeue()
				var contentToAdd *fyne.Container
				switch curDomNode.Type {
				case html.TextNode:
					curContent.Add(widget.NewLabel(curDomNode.Data))
					continue
				case html.ElementNode:
					fmt.Println(curDomNode.Data)
					switch curDomNode.Data {
					case "head":
						continue
					case "script":
						continue
					case "link":
						continue
					case "input":
						curContent.Add(widget.NewEntry())
						continue
					case "img":
						imgName := ""
						for _, attr := range curDomNode.Attr {
							if attr.Key == "src" {
								imgName = attr.Val
							}
						}
						if imgName == "" {
							fmt.Println("no image resource")
							panic("todo")
						}
						imgUrl, err := url.JoinPath(resp.Request.URL.String(), imgName)
						if err != nil {
							fmt.Println(err)
							panic("todo")
						}
						imgRes, err := fyne.LoadResourceFromURLString(imgUrl)
						if err != nil {
							fmt.Println(err)
							panic("todo")
						}
						curContent.Add(canvas.NewImageFromResource(imgRes))
						continue
					case "a":
						link := ""
						for _, attr := range curDomNode.Attr {
							if attr.Key == "href" {
								link = attr.Val
							}
						}
						if link == "" {
							fmt.Println("no link in href attribute")
							panic("todo")
						}
						parsedUrl, err := url.Parse(link)
						if err != nil {
							fmt.Println(err)
							panic("todo")
						}
						curContent.Add(widget.NewHyperlink(curDomNode.FirstChild.Data, parsedUrl))
						continue
					case "form":
						curContent.Add(container.New(layout.NewFormLayout()))
						continue
					case "div":
						contentToAdd = container.NewVBox()
					case "span":
						contentToAdd = container.NewHBox()
					case "center":
						contentToAdd = container.NewCenter()
					default:
						contentToAdd = container.NewWithoutLayout()
					}
					curContent.Add(contentToAdd)
				case html.DocumentNode:
				default:
					continue
				}
				curDomNodeChildren := convSeqToSlice(curDomNode.ChildNodes())
				if len(curDomNodeChildren) != 0 {
					domQueue.Enqueue(curDomNodeChildren...)
					contentQueue.Enqueue(contentToAdd)
				}
			}
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
	toolbar := container.New(&ToolbarLayout{}, refreshBtn, searchBar)
	w.SetContent(container.NewVBox(toolbar, content))
	w.ShowAndRun()
}
