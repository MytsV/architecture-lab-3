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
	panic("mockScreen: NewBuffer")
}
func (m mockScreen) NewTexture(size image.Point) (screen.Texture, error) {
	return new(mockTexture), nil
}
func (m mockScreen) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	panic("mockScreen: NewWindow")
}

type mockTexture struct {
	//рахує скільки операцій було застосовано до текстури
	FillCnt   int
	UploadCnt int
}

func (m *mockTexture) Release()          {}
func (m *mockTexture) Size() image.Point { return image.Point{400, 400} }
func (m *mockTexture) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{400, 400}}
}
func (m *mockTexture) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {
	m.UploadCnt++
}
func (m *mockTexture) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	m.FillCnt++
}

type mockOperation interface {
	// Do виконує зміну операції, повертаючи true, якщо текстура вважається готовою для відображення.
	Do(t screen.Texture) (ready bool)
}

type mockFillOperation struct{}

func (o mockFillOperation) Do(t screen.Texture) bool {
	t.Fill(t.Bounds(), color.White, screen.Src)
	return false
}

// UpdateOp операція, яка не змінює текстуру, але сигналізує, що текстуру потрібно розглядати як готову.
type mockUpdateOperation struct{}

func (op mockUpdateOperation) Do(t screen.Texture) bool {
	return true
}

// OperationList групує список операції в одну.
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

		err := l.Post(mockFillOperation{})
		if err == nil {
			t.Errorf("expexted err, got: %v", err)
		}

		err = l.StopAndWait()
		if err == nil {
			t.Errorf("expexted err, got:  %v", err)
		}
	})

	//!!!!перевір коректність використання лексики
	//чи має взагалі працювати така логіка?
	//я не знаю як це зробити
	t.Run("If Start method called the second time, it should start a different event loop (new message queue, textures) ", func(t *testing.T) {
		var l painter.Loop
		l.Start(mockScreen{})
		l.Start(mockScreen{})

	})

	//чи має взагалі працювати така логіка?
	//РеюЗабіліті
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
	})

	t.Run("StopAndWait method waits until message queue is empty and only after this stops event loop", func(t *testing.T) {
		var (
			l  painter.Loop
			tr testReceiver
		)
		l.Receiver = &tr
		l.Start(mockScreen{})

		//зробимо список операцій, одна з яких буде викликати StopAndWait метод і буде не останньою
		opL := mockOperationList{
			mockFillOperation{},
			mockFillOperation{},
			//mockOperationFunc(func(t screen.Texture) { l.StopAndWait() }),
			mockFillOperation{},
			mockUpdateOperation{},
		}

		l.Post(opL)
		if tr.LastTexture != nil {
			t.Fatal("Receiver got the texture too early")
		}
		l.StopAndWait()
		tx, _ := tr.LastTexture.(*mockTexture)
		if tx.FillCnt != 3 && tx.UploadCnt != 1 {
			t.Error("Unexpected number of fill calls:", tx.FillCnt)
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
		if tx.FillCnt != 5 && tx.UploadCnt != 1 {
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
