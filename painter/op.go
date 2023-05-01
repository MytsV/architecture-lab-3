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

// OperationList групує список операції в одну.
type OperationList []Operation

func (ol OperationList) Do(t screen.Texture) (ready bool) {
	for _, o := range ol {
		ready = o.Do(t) || ready
	}
	return
}

type OperationState struct {
	backgroundOperation Operation
}

func (os OperationState) Do(t screen.Texture) (ready bool) {
	if os.backgroundOperation != nil {
		os.backgroundOperation.Do(t)
	}
	return false
}

func (os *OperationState) Add(o Operation) {
	switch o.(type) {
	case OperationFunc:
		os.backgroundOperation = o
	default:
		panic("Unknown operation")
	}
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

// WhiteFill зафарбовує тестуру у білий колір. Може бути викоистана як Operation через OperationFunc(WhiteFill).
func WhiteFill(t screen.Texture) {
	t.Fill(t.Bounds(), color.White, screen.Src)
}

// GreenFill зафарбовує тестуру у зелений колір. Може бути викоистана як Operation через OperationFunc(GreenFill).
func GreenFill(t screen.Texture) {
	t.Fill(t.Bounds(), color.RGBA{G: 0xff, A: 0xff}, screen.Src)
}

var OperationRect = operationRect{}

type operationRect struct{}

func (o operationRect) Do(t screen.Texture) bool {
	rect := image.Rect(50, 50, 100, 100)
	t.Fill(rect, color.RGBA{R: 0xff, A: 0xff}, draw.Src)
	return false
}
