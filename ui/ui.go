package ui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"strconv"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dustin/go-humanize"
	"github.com/psanford/android-media-backup/db"
	"github.com/psanford/android-media-backup/jgo"
	"github.com/psanford/android-media-backup/ui/plog"
	"github.com/psanford/android-media-backup/upload"
	"github.com/psanford/android-media-backup/version"
)

type UI struct {
	db *db.DB
}

func New() *UI {
	store, err := db.Open()
	if err != nil {
		panic(fmt.Sprintf("Open db err: %s", err))
	}
	return &UI{
		db: store,
	}
}

func (ui *UI) Run() error {
	w := new(app.Window)
	w.Option(app.Size(unit.Dp(800), unit.Dp(700)))
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

	var (
		permResult <-chan jgo.PermResult
		viewEvent  app.ViewEvent

		th           = material.NewTheme()
		manualUpload = make(chan chan struct{})
		minuteTicker = time.NewTicker(1 * time.Minute)
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

	recheckStats := func() {
		lastSyncTime, _ = ui.db.LastCheckTime()
		lastFileUpload, _ = ui.db.LastFileUpload()
		pendingUploads, _ = ui.db.PendingUploads()
		recentUploads, _ = ui.db.UploadsSince(time.Now().Add(-30*24*time.Hour), db.UploadSuccess)
		recentFailedUploads, _ = ui.db.UploadsSince(time.Now().Add(-30*24*time.Hour), db.UploadFailed)

		files, _ = ui.db.GetFiles()

	}
	recheckStats()

	events := make(chan event.Event)
	acks := make(chan struct{})

	go func() {
		for {
			ev := w.Event()
			events <- ev
			<-acks
			if _, ok := ev.(app.DestroyEvent); ok {
				return
			}
		}
	}()

	var ops op.Ops
	for {
		select {
		case <-minuteTicker.C:
			recheckStats()
		case result := <-permResult:
			permResult = nil
			plog.Printf("Perm result: %t %s", result.Authorized, result.Err)
			if result.Authorized {
				plog.Printf("authorized: recheck files")
				upload.ScanFiles(ui.db)
				recheckStats()
			}
			w.Invalidate()

		case logMsg := <-plog.MsgChan():
			logText.Insert(logMsg)

		case e := <-events:
			switch e := e.(type) {
			case app.ViewEvent:
				viewEvent = e
				acks <- struct{}{}
			case app.DestroyEvent:
				acks <- struct{}{}
				return e.Err
			case app.ConfigEvent:
				plog.Printf("config event: recheck files")
				upload.ScanFiles(ui.db)
				recheckStats()
				plog.Printf("config event: recheck files done")
				acks <- struct{}{}
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				var testUploadClicked bool
				if uploadBtn.Clicked(gtx) {
					testUploadClicked = true
				}

				var resetFilesOnce sync.Once
				if resetBtn.Clicked(gtx) {
					resetFilesOnce.Do(func() {
						ui.db.ResetFiles()
					})
				}

				var resetFailedFilesOnce sync.Once
				for resetFailedBtn.Clicked(gtx) {
					resetFailedFilesOnce.Do(func() {
						ui.db.ResetFailedUploads()
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

				if wifiOnlyToggle.Update(gtx) {
					allowMobile := !wifiOnlyToggle.Value
					ui.db.SetAllowMobileUpload(allowMobile)
				}

				if enabledToggle.Update(gtx) {
					enabled := enabledToggle.Value
					ui.db.SetEnabled(enabled)

					if enabled {
						permResult = jgo.RequestPermission(viewEvent)
					}
				}

				ui.drawTabs(gtx, th)
				e.Frame(gtx.Ops)
				acks <- struct{}{}

			default:
				plog.Printf("unhandled event: %+v %T", e, e)
				acks <- struct{}{}
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
	resetFailedBtn   = new(widget.Clickable)

	lastSyncTime        time.Time
	lastFileUpload      time.Time
	pendingUploads      int
	recentUploads       int
	recentFailedUploads int

	settingsList = &layout.List{
		Axis: layout.Vertical,
	}

	debugList = &layout.List{
		Axis: layout.Vertical,
	}

	filesList = &layout.List{
		Axis: layout.Vertical,
	}

	files []db.File

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

func (ui *UI) drawTabs(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.H4(th, topLabel).Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return tabs.list.Layout(gtx, len(tabs.tabs), func(gtx C, tabIdx int) D {
				t := &tabs.tabs[tabIdx]
				if t.btn.Clicked(gtx) {
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
							return layout.UniformInset(unit.Dp(12)).Layout(gtx,
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
						tabHeight := 4
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
				case "Files":
					return ui.drawFiles(gtx, th)
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
					border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Dp(2)}
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
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(0.6, func(gtx C) D {
					return material.H6(th, "Last Sync Check:").Layout(gtx)
				}),

				layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
					str := "never"
					if !lastSyncTime.IsZero() {
						str = humanize.Time(lastSyncTime)
					}
					return material.H6(th, str).Layout(gtx)
				}),
			)
		},

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(0.6, func(gtx C) D {
					return material.H6(th, "Last Upload:").Layout(gtx)
				}),

				layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
					str := "never"
					if !lastFileUpload.IsZero() {
						str = humanize.Time(lastFileUpload)
					}
					return material.H6(th, str).Layout(gtx)
				}),
			)
		},

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(0.8, func(gtx C) D {
					return material.H6(th, "Pending Uploads:").Layout(gtx)
				}),

				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					return material.H6(th, strconv.Itoa(pendingUploads)).Layout(gtx)
				}),
			)
		},

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(0.8, func(gtx C) D {
					return material.H6(th, "Uploads in the last month:").Layout(gtx)
				}),

				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					return material.H6(th, strconv.Itoa(recentUploads)).Layout(gtx)
				}),
			)
		},

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(0.8, func(gtx C) D {
					return material.H6(th, "Failures in the last month:").Layout(gtx)
				}),

				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					return material.H6(th, strconv.Itoa(recentFailedUploads)).Layout(gtx)
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
		material.Button(th, resetFailedBtn, "Reset Failed Uploads").Layout,
		material.Button(th, resetBtn, "Reset Full DB State").Layout,
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func (ui *UI) drawFiles(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return filesList.Layout(gtx, len(files), func(gtx layout.Context, i int) layout.Dimensions {
		file := files[i]

		border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Dp(2)}

		borderA := widget.Border{Color: color.NRGBA{A: 0xff, R: 0xFF}, Width: unit.Dp(0)}
		borderB := widget.Border{Color: color.NRGBA{A: 0xff, G: 0xFF}, Width: unit.Dp(0)}
		borderC := widget.Border{Color: color.NRGBA{A: 0xff, B: 0xFF}, Width: unit.Dp(0)}

		return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			sz := gtx.Dp(unit.Dp(300))
			gtx.Constraints = layout.Exact(gtx.Constraints.Constrain(image.Point{X: sz, Y: sz}))
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Flexed(0.1, func(gtx C) D {
					return borderB.Layout(gtx, material.H6(th, file.Name).Layout)
				}),
				layout.Flexed(0.5, func(gtx C) D {

					img, err := ui.db.Thumbnail(file)
					if err != nil {
						img = image.NewRGBA(image.Rectangle{Max: image.Point{X: 256, Y: 256}})
					}

					wimg := widget.Image{
						Src: paint.NewImageOp(img),
						Fit: widget.Contain,
					}
					return borderA.Layout(gtx, wimg.Layout)
				}),
				layout.Flexed(0.1, func(gtx C) D {
					return borderC.Layout(gtx, material.H6(th, file.Created.In(time.Local).Format("01/02 15:04")).Layout)
				}),
				layout.Flexed(0.1, func(gtx C) D {
					return borderC.Layout(gtx, material.H6(th, file.State.String()).Layout)
				}),
				layout.Flexed(0.1, func(gtx C) D {
					ts := file.UploadStarted
					if !file.UploadEnd.IsZero() {
						ts = file.UploadEnd
					}
					return borderC.Layout(gtx, material.H6(th, ts.In(time.Local).Format("01/02 15:04")).Layout)
				}),
			)
		})
	})
}

func drawDebug(gtx layout.Context, th *material.Theme) layout.Dimensions {
	border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Dp(2)}

	widgets := []layout.Widget{
		material.H5(th, "Version:").Layout,
		material.H6(th, version.Version).Layout,
		material.H5(th, "Event Log").Layout,
		func(gtx C) D {
			return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = gtx.Dp(500)
				gtx.Constraints.Max.Y = gtx.Dp(500)
				return material.Editor(th, logText, "").Layout(gtx)
			})
		},
	}

	return debugList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}
