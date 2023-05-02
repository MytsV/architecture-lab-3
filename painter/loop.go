package painter

import (
	"fmt"
	"image"

	"golang.org/x/exp/shiny/screen"
)

// Receiver отримує текстуру, яка була підготовлена в результаті виконання команд у циелі подій.
type Receiver interface {
	Update(t screen.Texture)
}

// Loop реалізує цикл подій для формування текстури отриманої через виконання операцій отриманих з внутрішньої черги.
type Loop struct {
	Receiver Receiver

	next screen.Texture // текстура, яка зараз формується
	prev screen.Texture // текстура, яка була відправленя останнього разу у Receiver

	mq messageQueue

	shouldStop bool
	finished   chan struct{}
}

var size = image.Pt(800, 800)

// Start запускає цикл подій. Цей метод потрібно запустити до того, як викликати на ньому будь-які інші методи.
func (l *Loop) Start(s screen.Screen) {
	l.next, _ = s.NewTexture(size)
	l.prev, _ = s.NewTexture(size)

	// Ініціалізуємо чергу операцій.
	l.mq = *newQueue()

	// на випадок, якщо на лупі повторно викликається Start, без попередньої зупинки
	if l.finished != nil {
		l.StopAndWait()
	}
	// Ініціалізуємо індентифікатор завершення циклу.
	l.finished = make(chan struct{})

	//щоб можна було знову стартувати луп після його зупинки
	l.Post(OperationFunc(func(t screen.Texture) {
		l.shouldStop = false
	}))
	// Запускаємо рутину обробки повідомлень у черзі подій.
	go beginEventLoop(l)
}

func beginEventLoop(l *Loop) {
	for !l.shouldStop || !l.mq.isEmpty() {
		op := l.mq.pull()
		update := op.Do(l.next)
		if update {
			l.Receiver.Update(l.next)
			l.next, l.prev = l.prev, l.next
		}
	}
	close(l.finished)
	l.finished = nil
}

// Post додає нову операцію у внутрішню чергу.
func (l *Loop) Post(op Operation) error {
	// Перевіримо чи цикл подій запущено
	if l.finished == nil {
		return fmt.Errorf("Loop_Post error: event loop wasn't started")
	}

	// Додаємо операцію в чергу, якщо вона ненульова.
	if op != nil {
		l.mq.push(op)
		return nil
	}

	return fmt.Errorf("Loop_Post error: operation is nil")
}

// StopAndWait сигналізує про необхідність завершення циклу подій після виконання всіх операцій з черги і чекає на завершення.
func (l *Loop) StopAndWait() error {
	// Перевіримо чи цикл подій запущено
	if l.finished == nil {
		return fmt.Errorf("Loop_StopAndWait error: event loop wasn't started")
	}

	l.Post(OperationFunc(func(t screen.Texture) {
		l.shouldStop = true
	}))
	<-l.finished
	return nil
}

// messageQueue визначає асинхронну чергу операцій.
type messageQueue struct {
	ch chan Operation
}

// newQueue створює нову чергу з максимальною місткістю у 1024 операції.
func newQueue() *messageQueue {
	return &messageQueue{ch: make(chan Operation, 1024)}
}

func (mq *messageQueue) push(op Operation) {
	mq.ch <- op
}

func (mq *messageQueue) pull() Operation {
	return <-mq.ch
}

func (mq *messageQueue) isEmpty() bool {
	return len(mq.ch) == 0
}
