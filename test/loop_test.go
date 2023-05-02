package painter

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/MytsV/architecture-lab-3/painter"
	"golang.org/x/exp/shiny/screen"
)

type testReceiver struct {
	LastTexture screen.Texture
}

func (tr *testReceiver) Update(t screen.Texture) {
	tr.LastTexture = t
}

type mockScreen struct{}

func (m mockScreen) NewBuffer(size image.Point) (screen.Buffer, error) {
	return nil, nil
}
func (m mockScreen) NewTexture(size image.Point) (screen.Texture, error) {
	return new(mockTexture), nil
}
func (m mockScreen) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	panic("mockScreen: NewWindow")
}

type mockTexture struct {
	FillCnt   int
	UploadCnt int
}

func (m *mockTexture) Release()          {}
func (m *mockTexture) Size() image.Point { return image.Point{0, 0} }
func (m *mockTexture) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{0, 0}}
}
func (m *mockTexture) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {
	m.UploadCnt++
}
func (m *mockTexture) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	m.FillCnt++
}

type mockOperation interface {
	Do(t screen.Texture) (ready bool)
}

type mockFillOperation struct{}

func (o mockFillOperation) Do(t screen.Texture) bool {
	t.Fill(t.Bounds(), color.White, screen.Src)
	return false
}

type mockUpdateOperation struct{}

func (op mockUpdateOperation) Do(t screen.Texture) bool {
	b, _ := mockScreen{}.NewBuffer(image.Point{0, 0})
	t.Upload(image.Point{0, 0}, b, t.Bounds())
	return true
}

type mockOperationList []mockOperation

func (ol mockOperationList) Do(t screen.Texture) (ready bool) {
	for _, o := range ol {
		ready = o.Do(t) || ready
	}
	return
}

type mockOperationFunc func(t screen.Texture)

func (f mockOperationFunc) Do(t screen.Texture) bool {
	f(t)
	return false
}

// Start запускає цикл подій. Цей метод потрібно запустити до того, як викликати на ньому будь-які інші методи.
func TestLoop_Start(t *testing.T) {
	t.Run("Start method starts event loop", func(t *testing.T) {
		var l painter.Loop

		l.Start(mockScreen{})
		err := l.Post(mockFillOperation{})
		if err != nil {
			t.Errorf("expexted nil, got: %v", err)
		}
	})

	t.Run("Without first calling Start method, event loop won't work", func(t *testing.T) {
		var l painter.Loop
		var count int

		err := l.Post(mockOperationFunc(func(t screen.Texture) { count++ }))
		if err == nil {
			t.Errorf("expexted err, got: %v", err)
		}
		if count != 0 {
			t.Errorf("expexted 0, got: %v", count)
		}

		err = l.StopAndWait()
		if err == nil {
			t.Errorf("expexted err, got:  %v", err)
		}
	})

	t.Run("If Start method called second time, it should start a different event loop (new message queue, textures) ", func(t *testing.T) {
		var (
			l  painter.Loop
			tr testReceiver
		)
		l.Receiver = &tr

		l.Start(mockScreen{})
		l.Post(mockFillOperation{})
		l.Post(mockUpdateOperation{})

		l.Start(mockScreen{})
		l.Post(mockUpdateOperation{})

		if tr.LastTexture != nil {
			t.Fatal("Receiver got the texture too early")
		}
		l.StopAndWait()
		tx, _ := tr.LastTexture.(*mockTexture)
		if tx.FillCnt != 0 || tx.UploadCnt != 1 {
			t.Errorf("Unexpected number of calls. Expected 0 fill, got %v; expected 1 upload, got %v", tx.FillCnt, tx.UploadCnt)
		}
	})

	t.Run("If Start method was called after the loop was stopped, it should start the event loop again", func(t *testing.T) {
		var l painter.Loop
		l.Start(mockScreen{})
		err := l.StopAndWait()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		l.Start(mockScreen{})
		err = l.Post(mockFillOperation{})
		if err != nil {
			t.Errorf("expexted nil, got: %v", err)
		}
	})
}

// StopAndWait сигналізує про необхідність завершення циклу подій після виконання всіх операцій з черги і чекає на завершення.
func TestLoop_StopAndWait(t *testing.T) {
	t.Run("StopAndWait method stops event loop", func(t *testing.T) {
		var l painter.Loop

		l.Start(mockScreen{})
		err := l.StopAndWait()
		if err != nil {
			t.Errorf("expexted nil, got: %v", err)
		}

		err = l.Post(mockFillOperation{})
		if err == nil {
			t.Errorf("expexted err, got: %v", err)
		}
	})

	t.Run("StopAndWait method waits until message queue is empty and only after this stops event loop", func(t *testing.T) {
		var (
			l  painter.Loop
			tr testReceiver
		)
		l.Receiver = &tr
		l.Start(mockScreen{})

		opL := mockOperationList{
			mockFillOperation{},
			mockFillOperation{},
			mockFillOperation{},
			mockFillOperation{},
			mockFillOperation{},
			mockFillOperation{},
			mockUpdateOperation{},
		}
		l.Post(opL)
		if tr.LastTexture != nil {
			t.Fatal("Receiver got the texture too early")
		}

		l.StopAndWait()
		tx, _ := tr.LastTexture.(*mockTexture)
		if tx.FillCnt != 6 || tx.UploadCnt != 1 {
			t.Errorf("Unexpected number of calls. Expected 0 fill, got %v; expected 1 upload, got %v", tx.FillCnt, tx.UploadCnt)
		}
	})

	t.Run("StopAndWait must return error, if event loop wasn't started", func(t *testing.T) {
		var l painter.Loop
		err := l.StopAndWait()
		if err == nil {
			t.Errorf("expexted nil, got: %v", err)
		}
	})

	t.Run("StopAndWait must return error, if event loop was already stopped", func(t *testing.T) {
		var l painter.Loop
		l.Start(mockScreen{})
		err := l.StopAndWait()
		if err != nil {
			t.Fatalf("unexpexted err: %v", err)
		}

		err = l.StopAndWait()
		if err == nil {
			t.Errorf("expexted nil, got: %v", err)
		}
	})
}

func TestLoop_Post(t *testing.T) {
	t.Run("The number of the correct operations, that were posted to the loop, should be equal to the number of the executed", func(t *testing.T) {
		var (
			l  painter.Loop
			tr testReceiver
		)
		l.Receiver = &tr
		l.Start(mockScreen{})

		opL := mockOperationList{
			mockFillOperation{},
			mockFillOperation{},
			mockFillOperation{},
			mockFillOperation{},
			mockFillOperation{},
			mockUpdateOperation{},
		}

		l.Post(opL)
		if tr.LastTexture != nil {
			t.Fatal("Receiver got the texture too early")
		}

		l.StopAndWait()
		tx, ok := tr.LastTexture.(*mockTexture)
		if !ok {
			t.Fatal("Receiver still has no texture")
		}
		if tx.FillCnt != 5 || tx.UploadCnt != 1 {
			t.Error("Unexpected number of fill calls:", tx.FillCnt)
		}

	})

	t.Run("Post method must return an err if event loop wasn't started", func(t *testing.T) {
		var l painter.Loop
		l.StopAndWait()
		err := l.Post(nil)
		if err == nil {
			t.Errorf("expected err, got %v", err)
		}
	})

	t.Run("Post method must return an err if event loop was stopped", func(t *testing.T) {
		var l painter.Loop
		l.Start(mockScreen{})
		l.StopAndWait()
		err := l.Post(nil)
		if err == nil {
			t.Errorf("expected err, got %v", err)
		}
	})

	t.Run("Post method must return an err if operation was nil", func(t *testing.T) {
		var l painter.Loop
		err := l.Post(nil)
		if err == nil {
			t.Errorf("expected err, got %v", err)
		}
	})
}
