package painter

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/exp/shiny/screen"
)

// Operation змінює вхідну текстуру.
type Operation interface {
	// Do виконує зміну операції, повертаючи true, якщо текстура вважається готовою для відображення.
	Do(t screen.Texture) (ready bool)
}

// StateTweaker вміє змінювати стан малюнку.
type StateTweaker interface {
	// SetState виконує зміну переданої операції зі станом.
	SetState(sol *StatefulOperationList)
}

// StatefulOperationList групує операції, що впливають на стан, в одну.
type StatefulOperationList struct {
	BgOperation      Operation
	BgRectOperation  Operation
	FigureOperations []*OperationFigure
}

// Виконує операції відносно до збереженого стану.
func (sol StatefulOperationList) Do(t screen.Texture) (ready bool) {
	if sol.BgOperation != nil {
		sol.BgOperation.Do(t)
	} else {
		t.Fill(t.Bounds(), color.Black, screen.Src)
	}
	if sol.BgRectOperation != nil {
		sol.BgRectOperation.Do(t)
	}
	for _, op := range sol.FigureOperations {
		op.Do(t)
	}
	return false
}

func (sol *StatefulOperationList) Update(o StateTweaker) {
	o.SetState(sol)
}

// UpdateOp операція, яка не змінює текстуру, але сигналізує, що текстуру потрібно розглядати як готову.
var UpdateOp = updateOp{}

type updateOp struct{}

func (op updateOp) Do(t screen.Texture) bool { return true }

// OperationFunc використовується для перетворення функції оновлення текстури в Operation.
type OperationFunc func(t screen.Texture)

func (f OperationFunc) Do(t screen.Texture) bool {
	f(t)
	return false
}

// OperationFill зафарбовує текстуру у будь-який колір.
type OperationFill struct {
	Color color.Color
}

func (op OperationFill) Do(t screen.Texture) bool {
	t.Fill(t.Bounds(), op.Color, screen.Src)
	return false
}

func (op OperationFill) SetState(sol *StatefulOperationList) {
	sol.BgOperation = op
}

type RelativePoint struct {
	X float64
	Y float64
}

func (p RelativePoint) ToAbs(size image.Point) image.Point {
	return image.Point{
		X: int(p.X * float64(size.X)),
		Y: int(p.Y * float64(size.Y)),
	}
}

type OperationBGRect struct {
	Min RelativePoint
	Max RelativePoint
}

func (op OperationBGRect) Do(t screen.Texture) bool {
	minAbs := op.Min.ToAbs(t.Size())
	maxAbs := op.Max.ToAbs(t.Size())

	rect := image.Rect(minAbs.X, minAbs.Y, maxAbs.X, maxAbs.Y)
	t.Fill(rect, color.Black, draw.Src)
	return false
}

func (op OperationBGRect) SetState(sol *StatefulOperationList) {
	sol.BgRectOperation = op
}

type OperationFigure struct {
	Center RelativePoint
}

func (op OperationFigure) Do(t screen.Texture) bool {
	centerAbs := op.Center.ToAbs(t.Size())
	x := centerAbs.X
	y := centerAbs.Y

	//Виміри фігури
	hlen := 115
	hwidth := 35
	yellow := color.RGBA{R: 0xff, G: 0xff, A: 0xff}

	horizontal := image.Rect(x-hlen, y+hlen, x+hlen, y+hlen-hwidth*2)
	t.Fill(horizontal, yellow, draw.Src)
	vertical := image.Rect(x-hwidth, y-hlen, x+hwidth, y+hlen)
	t.Fill(vertical, yellow, draw.Src)

	return false
}

func (op OperationFigure) SetState(sol *StatefulOperationList) {
	//sol.FigureOperations = append(sol.FigureOperations, &op)
}

type MoveTweaker struct {
	Offset RelativePoint
}

func (t MoveTweaker) SetState(sol *StatefulOperationList) {
	for _, op := range sol.FigureOperations {
		op.Center.X += t.Offset.X
		op.Center.Y += t.Offset.Y
	}
}

type ResetTweaker struct{}

func (op ResetTweaker) SetState(sol *StatefulOperationList) {
	sol.BgOperation = nil
	sol.BgRectOperation = nil
	sol.FigureOperations = []*OperationFigure{}
}
