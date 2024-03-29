package ui

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/imageutil"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

const windowSize = 800

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	tx   chan screen.Texture
	done chan struct{}

	sz size.Event
	mp image.Point
}

func (pw *Visualizer) Main() {
	pw.tx = make(chan screen.Texture, 1024)
	pw.done = make(chan struct{})
	pw.mp.X = windowSize / 2
	pw.mp.Y = windowSize / 2
	driver.Main(pw.run)
}

func (pw *Visualizer) Update(t screen.Texture) {
	pw.tx <- t
}

func (pw *Visualizer) run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  pw.Title,
		Width:  windowSize,
		Height: windowSize,
	})

	//Виклик OnScreenReady був перенесений, оскільки пізня ініціалізація вікна викликала помилку сегментації
	if pw.OnScreenReady != nil {
		pw.OnScreenReady(s)
	}

	if err != nil {
		log.Fatal("Failed to initialize the app window:", err)
	}
	defer func() {
		w.Release()
		close(pw.done)
	}()

	pw.w = w

	events := make(chan any)
	go func() {
		for {
			e := w.NextEvent()
			if pw.Debug {
				log.Printf("new event: %v", e)
			}
			if detectTerminate(e) {
				close(events)
				break
			}
			events <- e
		}
	}()

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)

		case t = <-pw.tx:
			w.Send(paint.Event{})
		}
	}
}

func detectTerminate(e any) bool {
	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {
			return true // Window destroy initiated.
		}
	case key.Event:
		if e.Code == key.CodeEscape {
			return true // Esc pressed.
		}
	}
	return false
}

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	switch e := e.(type) {

	case size.Event: // Оновлення даних про розмір вікна.
		pw.sz = e

	case error:
		log.Printf("ERROR: %s", e)

	case mouse.Event:
		if t == nil {
			if e.Button == mouse.ButtonLeft && e.Direction == mouse.DirPress {
				pw.mp.X = int(e.X)
				pw.mp.Y = int(e.Y)
				pw.w.Send(paint.Event{})
			}
		}

	case paint.Event:
		// Малювання контенту вікна.
		if t == nil {
			pw.drawDefaultUI()
		} else {
			// Використання текстури отриманої через виклик Update.
			pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)
		}
		pw.w.Publish()
	}
}

func (pw *Visualizer) drawDefaultUI() {
	pw.w.Fill(pw.sz.Bounds(), color.Black, draw.Src) // Фон.

	//Малювання фігури - жовта буква "Т"
	pw.DrawFigure(pw.mp.X, pw.mp.Y)

	// Малювання білої рамки.
	for _, br := range imageutil.Border(pw.sz.Bounds(), 10) {
		pw.w.Fill(br, color.White, draw.Src)
	}
}

func (pw *Visualizer) DrawFigure(x, y int) {
	//Виміри фігури
	hlen := 115
	hwidth := 35
	yellow := color.RGBA{R: 0xff, G: 0xff, A: 0xff}

	horizontal := image.Rect(x-hlen, y-hlen, x+hlen, y-hlen+hwidth*2)
	pw.w.Fill(horizontal, yellow, draw.Src)
	vertical := image.Rect(x-hwidth, y-hlen, x+hwidth, y+hlen)
	pw.w.Fill(vertical, yellow, draw.Src)
}
