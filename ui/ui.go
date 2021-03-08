package ui

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"sync"

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
	"github.com/psanford/android-media-backup/db"
	"github.com/psanford/android-media-backup/jgo"
	"github.com/psanford/android-media-backup/ui/plog"
	"github.com/psanford/android-media-backup/upload"
)

type UI struct {
	db *db.DB
}

func New() *UI {
	store, err := db.Open()
	if err != nil {
		panic(err)
	}
	return &UI{
		db: store,
	}
}

func (ui *UI) Run() error {
	w := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(700)))
	dataDir, err := app.DataDir()
	if err != nil {
		plog.Printf("DataDir err: %s", err)
	} else {
		plog.Printf("DataDir: %s", dataDir)
	}

	if err := jgo.StartBGWorker(); err != nil {
		log.Fatal(err)
	}

	if err := ui.loop(w); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (ui *UI) loop(w *app.Window) error {
	enabledConf, err := ui.db.Enabled()
	if err != nil {
		plog.Printf("get enabled err: %s", err)
	}
	allowMobileUpload, err := ui.db.AllowMobileUpload()
	if err != nil {
		plog.Printf("get allowMobile err: %s", err)
	}

	url, err := ui.db.URL()
	if err != nil {
		plog.Printf("get url err: %s", err)
	}
	username, err := ui.db.Username()
	if err != nil {
		plog.Printf("get username err: %s", err)
	}
	password, err := ui.db.Password()
	if err != nil {
		plog.Printf("get password err: %s", err)
	}

	urlEditor.SetText(url)
	usernameEditor.SetText(username)
	if password != "" {
		passwordEditor.SetText(password)
	}
	enabledToggle.Value = enabledConf
	wifiOnlyToggle.Value = !allowMobileUpload

	th := material.NewTheme(gofont.Collection())

	var (
		permResult <-chan jgo.PermResult
		viewEvent  app.ViewEvent

		manualUpload = make(chan chan struct{})
	)

	go func() {
		for result := range manualUpload {
			upload.Upload()
			select {
			case result <- struct{}{}:
			default:
			}
			w.Invalidate()
		}
	}()

	var ops op.Ops
	for {
		select {
		case result := <-permResult:
			permResult = nil
			plog.Printf("Perm result: %t %s", result.Authorized, result.Err)

			files, err := ioutil.ReadDir("/sdcard/DCIM/Camera")
			if err != nil {
				plog.Printf("read sdcard err: %s", err)
			} else {
				var names []string
				for _, f := range files {
					names = append(names, f.Name())
				}
				plog.Printf("sdcard pictures: %+v", names)
			}
			w.Invalidate()

		case logMsg := <-plog.MsgChan():
			logText.Insert(logMsg)

		case e := <-w.Events():
			switch e := e.(type) {
			case app.ViewEvent:
				viewEvent = e
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				var testUploadClicked bool
				for uploadBtn.Clicked() {
					testUploadClicked = true
				}

				var resetFilesOnce sync.Once
				for resetBtn.Clicked() {
					resetFilesOnce.Do(func() {
						ui.db.ResetFiles()
					})
				}

				if urlEditor.Text() != url {
					url = urlEditor.Text()
					ui.db.SetURL(url)
				}

				if usernameEditor.Text() != username {
					username = usernameEditor.Text()
					ui.db.SetUsername(username)
				}

				if passwordEditor.Text() != password {
					password = passwordEditor.Text()
					ui.db.SetPassword(password)
				}

				if testUploadClicked {
					plog.Printf("start test upload")
					result := make(chan struct{}, 1)
					select {
					case manualUpload <- result:
						uploadInProgress = true
						go func() {
							<-result
							uploadInProgress = false
						}()
					default:
						plog.Printf("upload already in progress")
					}
				}

				if wifiOnlyToggle.Changed() {
					allowMobile := !wifiOnlyToggle.Value
					ui.db.SetAllowMobileUpload(allowMobile)
				}

				if enabledToggle.Changed() {
					enabled := enabledToggle.Value
					ui.db.SetEnabled(enabled)

					if enabled {
						permResult = jgo.RequestPermission(viewEvent)
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
	logText   = new(widget.Editor)
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
	uploadInProgress = false
	uploadBtn        = new(widget.Clickable)
	resetBtn         = new(widget.Clickable)

	settingsList = &layout.List{
		Axis: layout.Vertical,
	}

	topLabel       = "Android Media Backup"
	enabledToggle  = new(widget.Bool)
	wifiOnlyToggle = new(widget.Bool)

	tabs = Tabs{
		tabs: []Tab{
			{
				Title: "Settings",
			},
			{
				Title: "Files",
			},
			{
				Title: "Debug",
			},
		},
	}
)

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
				selected := tabs.tabs[tabs.selected].Title
				switch selected {
				case "Settings":
					return drawSettings(gtx, th)
				case "Debug":
					return drawDebug(gtx, th)
				default:
					return layout.Center.Layout(gtx,
						material.H1(th, fmt.Sprintf("Tab content %s", selected)).Layout,
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
				layout.Flexed(0.8, func(gtx C) D {
					return material.H6(th, "Enable Automatic Backups").Layout(gtx)
				}),
				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx,
						material.CheckBox(th, enabledToggle, "").Layout,
					)
				}),
			)
		},

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(0.8, func(gtx C) D {
					return material.H6(th, "On Wifi Only").Layout(gtx)
				}),

				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx,
						material.CheckBox(th, wifiOnlyToggle, "").Layout,
					)
				}),
			)
		},

		func(gtx layout.Context) layout.Dimensions {
			if uploadInProgress || !enabledToggle.Value {
				gtx = gtx.Disabled()
			}
			btn := material.Button(th, uploadBtn, "Test Upload")
			return btn.Layout(gtx)
		},
		material.Button(th, resetBtn, "Reset Files").Layout,
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func drawDebug(gtx layout.Context, th *material.Theme) layout.Dimensions {
	border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Px(2)}

	widgets := []layout.Widget{
		material.H5(th, "Event Log").Layout,
		func(gtx C) D {
			return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = gtx.Px(unit.Dp(500))
				gtx.Constraints.Max.Y = gtx.Px(unit.Dp(500))
				return material.Editor(th, logText, "").Layout(gtx)
			})
		},
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}
