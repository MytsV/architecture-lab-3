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

// StatefulOperation наслідує Operation, але крім цього ще й вміє змінювати стан малюнку.
type StatefulOperation interface {
	Operation
	// SetState виконує зміну переданої операції зі станом.
	SetState(sol *StatefulOperationList)
}

// StatefulOperationList групує операції, що впливають на стан, в одну.
type StatefulOperationList struct {
	backgroundOperation Operation
}

// Виконує операції відносно до збереженого стану.
func (sol StatefulOperationList) Do(t screen.Texture) (ready bool) {
	if sol.backgroundOperation != nil {
		sol.backgroundOperation.Do(t)
	}
	return false
}

func (sol *StatefulOperationList) Update(o StatefulOperation) {
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
	sol.backgroundOperation = op
}

var OperationRect = operationRect{}

type operationRect struct{}

func (o operationRect) Do(t screen.Texture) bool {
	rect := image.Rect(50, 50, 100, 100)
	t.Fill(rect, color.RGBA{R: 0xff, A: 0xff}, draw.Src)
	return false
}
