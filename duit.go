package duit

import (
	"image"
	"log"

	"9fans.net/go/draw"
)

const (
	Margin  = 4
	Padding = 2
	Border  = 1

	ScrollbarSize = 8
)
const (
	Button1 = 1<<iota
	Button2
	Button3
	Button4
	Button5
)
const (
	Fn = 0xf000 // use like Fn + <number>

	ArrowUp    = 0xf00e
	ArrowDown  = 0x80
	ArrowLeft  = 0xf011
	ArrowRight = 0xf012
	PageUp     = 0xf00f
	PageDown   = 0xf013
)

type DUI struct {
	Display  *draw.Display
	Mousectl *draw.Mousectl
	Kbdctl   *draw.Keyboardctl
	Top      UI

	mouse       draw.Mouse
	lastMouseUI UI
	logEvents   bool
}

type sizes struct {
	margin  int
	padding int
	border  int
	space   int
}

func check(err error, msg string) {
	if err != nil {
		log.Printf(msg)
		panic(err)
	}
}

func NewDUI(name, dim string) (*DUI, error) {
	dui := &DUI{}
	display, err := draw.Init(nil, "", name, dim)
	if err != nil {
		return nil, err
	}
	dui.Display = display

	dui.Mousectl = display.InitMouse()
	dui.Kbdctl = display.InitKeyboard()

	return dui, nil
}

func (d *DUI) Render() {
	d.Top.Layout(d.Display, d.Display.ScreenImage.R, image.ZP)
	d.Display.ScreenImage.Draw(d.Display.ScreenImage.R, d.Display.White, nil, image.ZP)
	d.Redraw()
}

func (d *DUI) Redraw() {
	d.Top.Draw(d.Display, d.Display.ScreenImage, image.ZP, d.mouse)
	d.Display.Flush()
}

func (d *DUI) Mouse(m draw.Mouse) {
	d.mouse = m
	if d.logEvents {
		log.Printf("mouse %v, %b\n", m, m.Buttons)
	}
	r := d.Top.Mouse(m)
	if r.Layout {
		d.Render()
	} else if r.Hit != d.lastMouseUI || r.Redraw {
		d.Redraw()
	}
	d.lastMouseUI = r.Hit
}

func (d *DUI) Resize() {
	if d.logEvents {
		log.Printf("resize")
	}
	check(d.Display.Attach(draw.Refmesg), "attach after resize")
	d.Render()
}

func (d *DUI) Key(r rune) {
	if d.logEvents {
		log.Printf("kdb %c, %x\n", r, r)
	}
	if r == Fn+1 {
		d.logEvents = !d.logEvents
	}
	result := d.Top.Key(image.ZP, d.mouse, r)
	if !result.Consumed && r == '\t' {
		first := d.Top.FirstFocus()
		if first != nil {
			result.Warp = first
			result.Consumed = true
		}
	}
	if result.Warp != nil {
		err := d.Display.MoveTo(*result.Warp)
		if err != nil {
			log.Printf("move mouse to %v: %v\n", result.Warp, err)
		}
		m := d.mouse
		m.Point = *result.Warp
		result2 := d.Top.Mouse(m)
		result.Redraw = result.Redraw || result2.Redraw || true
		result.Layout = result.Layout || result2.Layout
		d.mouse = m
		d.lastMouseUI = result2.Hit
	}
	if result.Layout {
		d.Render()
	} else if result.Redraw {
		d.Redraw()
	}
}

func (d *DUI) Focus(ui UI) {
	p := d.Top.Focus(ui)
	if p == nil {
		return
	}
	err := d.Display.MoveTo(*p)
	if err != nil {
		log.Printf("move mouse to %v: %v\n", *p, err)
		return
	}
	d.mouse.Point = *p
	r := d.Top.Mouse(d.mouse)
	if r.Layout {
		d.Render()
	} else {
		d.Redraw()
	}
}

func scale(d *draw.Display, n int) int {
	return (d.DPI / 100) * n
}

func (d *DUI) Scale(n int) int {
	return (d.Display.DPI / 100) * n
}

func setSizes(d *draw.Display, sizes *sizes) {
	sizes.padding = d.Scale(Padding)
	sizes.margin = d.Scale(Margin)
	sizes.border = Border // slim border is nicer
	sizes.space = sizes.margin + sizes.border + sizes.padding
}
