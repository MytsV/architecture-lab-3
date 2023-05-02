package painter

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/MytsV/architecture-lab-3/painter"
	"golang.org/x/exp/shiny/screen"
)

func TestLoop_(t *testing.T) {
	var (
		l  painter.Loop
		tr testReceiver
	)

	l.Receiver = &tr

	l.Start(mockScreen{})
	l.Post(painter.OperationFunc(painter.WhiteFill))
	l.Post(painter.UpdateOp)
	l.StopAndWait()
	l.Post(painter.UpdateOp)

	l.Post(painter.OperationFunc(func(screen.Texture) {
		l.StopAndWait()
	}))
	fmt.Println("kek")

	tx, _ := tr.LastTexture.(*mockTexture)
	if tx.FillCnt != 1 {
		t.Error("Unexpected number of fill calls:", tx.FillCnt)
	}

}

func TestLoop_Post(t *testing.T) {
	var (
		l  painter.Loop
		tr testReceiver
	)

	l.Receiver = &tr

	l.Start(mockScreen{})

	l.Post(painter.OperationFunc(painter.WhiteFill))
	l.Post(painter.OperationFunc(painter.GreenFill))
	l.Post(painter.UpdateOp)
	if tr.LastTexture != nil {
		t.Fatal("Receiver got the texture too early")
	}

	l.StopAndWait()
	tx, ok := tr.LastTexture.(*mockTexture)
	if !ok {
		t.Fatal("Receiver still has no texture")
	}
	if tx.FillCnt != 2 {
		t.Error("Unexpected number of fill calls:", tx.FillCnt)
	}
}

// тест, де викликали СтопендВейт з повною чергою. Перевірити чи вона довиконається
func TestLoop_FullQ(t *testing.T) {
	var (
		l  painter.Loop
		tr testReceiver
	)

	l.Receiver = &tr

	l.Start(mockScreen{})
	l.Post(painter.OperationFunc(painter.WhiteFill))
	l.Post(painter.OperationFunc(painter.GreenFill))
	l.Post(painter.OperationFunc(painter.WhiteFill))
	l.Post(painter.OperationFunc(painter.GreenFill))
	l.Post(painter.OperationFunc(painter.WhiteFill))
	l.Post(painter.OperationFunc(painter.GreenFill))
	l.Post(painter.OperationFunc(painter.WhiteFill))
	l.Post(painter.OperationFunc(painter.GreenFill))
	l.Post(painter.UpdateOp)
	if tr.LastTexture != nil {
		t.Fatal("Receiver got the texture too early")
	}

	l.StopAndWait()
	tx, ok := tr.LastTexture.(*mockTexture)
	if !ok {
		t.Fatal("Receiver still has no texture")
	}
	if tx.FillCnt != 8 {
		t.Error("Unexpected number of fill calls:", tx.FillCnt)
	}

}

type testReceiver struct {
	LastTexture screen.Texture
}

func (tr *testReceiver) Update(t screen.Texture) {
	tr.LastTexture = t
}

type mockScreen struct{}

func (m mockScreen) NewBuffer(size image.Point) (screen.Buffer, error) {
	panic("mockScreen: NewBuffer")
}
func (m mockScreen) NewTexture(size image.Point) (screen.Texture, error) {
	return new(mockTexture), nil
}
func (m mockScreen) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	panic("mockScreen: NewWindow")
}

type mockTexture struct {
	FillCnt int
}

func (m *mockTexture) Release()          {}
func (m *mockTexture) Size() image.Point { return image.Point{400, 400} }
func (m *mockTexture) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{400, 400}}
}
func (m *mockTexture) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {
	panic("mockTexture: Upload")
}
func (m *mockTexture) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	m.FillCnt++
}
