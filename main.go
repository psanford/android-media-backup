package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	flag.Parse()

	go func() {
		w := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(700)))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())

	var ops op.Ops
	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				if checkbox.Changed() {
					if checkbox.Value {
						transformTime = e.Now
					} else {
						transformTime = time.Time{}
					}
				}

				layout.Inset{
					Bottom: e.Insets.Bottom,
					Left:   e.Insets.Left,
					Right:  e.Insets.Right,
					Top:    e.Insets.Top,
				}.Layout(gtx, func(gtx C) D {
					return drawTabs(gtx, th)
				})
				e.Frame(gtx.Ops)
			}
		}
	}
}

var (
	editor    = new(widget.Editor)
	urlEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	usernameEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	passwordEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	button            = new(widget.Clickable)
	greenButton       = new(widget.Clickable)
	flatBtn           = new(widget.Clickable)
	disableBtn        = new(widget.Clickable)
	radioButtonsGroup = new(widget.Enum)
	list              = &layout.List{
		Axis: layout.Vertical,
	}

	green         = true
	topLabel      = "Android Media Backup"
	checkbox      = new(widget.Bool)
	swtch         = new(widget.Bool)
	transformTime time.Time
	float         = new(widget.Float)
)

var tabs = Tabs{
	tabs: []Tab{
		{
			Title: "Settings",
		},
		{
			Title: "Log",
		},
		{
			Title: "Debug",
		},
	},
}

var slider Slider

type Tabs struct {
	list     layout.List
	tabs     []Tab
	selected int
}

type Tab struct {
	btn   widget.Clickable
	Title string
}

func init() {
}

type (
	C = layout.Context
	D = layout.Dimensions
)

func drawTabs(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.H4(th, topLabel).Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return tabs.list.Layout(gtx, len(tabs.tabs), func(gtx C, tabIdx int) D {
				t := &tabs.tabs[tabIdx]
				if t.btn.Clicked() {
					if tabs.selected < tabIdx {
						slider.PushLeft()
					} else if tabs.selected > tabIdx {
						slider.PushRight()
					}
					tabs.selected = tabIdx
				}
				var tabWidth int
				return layout.Stack{Alignment: layout.S}.Layout(gtx,
					layout.Stacked(func(gtx C) D {
						dims := material.Clickable(gtx, &t.btn, func(gtx C) D {
							return layout.UniformInset(unit.Sp(12)).Layout(gtx,
								material.H6(th, t.Title).Layout,
							)
						})
						tabWidth = dims.Size.X
						return dims
					}),
					layout.Stacked(func(gtx C) D {
						if tabs.selected != tabIdx {
							return layout.Dimensions{}
						}
						tabHeight := gtx.Px(unit.Dp(4))
						tabRect := image.Rect(0, 0, tabWidth, tabHeight)
						paint.FillShape(gtx.Ops, th.Palette.ContrastBg, clip.Rect(tabRect).Op())
						return layout.Dimensions{
							Size: image.Point{X: tabWidth, Y: tabHeight},
						}
					}),
				)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return slider.Layout(gtx, func(gtx C) D {
				if tabs.selected == 0 {
					return drawSettings(gtx, th)
				} else {
					return layout.Center.Layout(gtx,
						material.H1(th, fmt.Sprintf("Tab content %s", tabs.tabs[tabs.selected].Title)).Layout,
					)
				}
			})
		}),
	)
}

func drawSettings(gtx layout.Context, th *material.Theme) layout.Dimensions {
	textField := func(label, hint string, editor *widget.Editor) func(layout.Context) layout.Dimensions {
		return func(gtx layout.Context) layout.Dimensions {
			flex := layout.Flex{
				Axis: layout.Vertical,
			}
			return flex.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.H5(th, label).Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					e := material.Editor(th, editor, hint)
					border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Px(2)}
					return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
					})
				}),
			)
		}
	}

	widgets := []layout.Widget{

		textField("Server URL", "URL", urlEditor),
		textField("Username", "Username", usernameEditor),
		textField("Password", "Password", passwordEditor),

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx,
						material.Switch(th, swtch).Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						text := "enabled"
						if !swtch.Value {
							text = "disabled"
							gtx = gtx.Disabled()
						}
						btn := material.Button(th, disableBtn, text)
						return btn.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if !swtch.Value {
							return layout.Dimensions{}
						}
						return material.Loader(th).Layout(gtx)
					})
				}),
			)
		},
	}

	return list.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}
