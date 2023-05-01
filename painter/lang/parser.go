package lang

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/MytsV/architecture-lab-3/painter"
)

// Parser уміє прочитати дані з вхідного io.Reader та повернути список операцій представлені вхідним скриптом.
type Parser struct {
	// Зберігає стан малюнку у спеціальній операції.
	state painter.StatefulOperationList
}

func (p *Parser) Parse(in io.Reader) ([]painter.Operation, error) {
	var res []painter.Operation

	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		commandLine := scanner.Text()
		// Отримуємо відповідну до команди операцію ії тип.
		op, status := parse(commandLine)
		if op == nil || status == absent {
			// Якщо операція не імплементована, видаємо помилку.
			return nil, fmt.Errorf("Unknown command")
		} else if status == regular {
			// Якщо операція звичайна, просто додаємо її у список до передачі в цикл.
			res = append(res, op)
		} else {
			// Якщо операція впливає на стан, пробуємо перевести її під інтерфейс StatefulOperation.
			stateOp, ok := op.(painter.StatefulOperation)
			if !ok {
				// Якщо парсер хоче оновлення стану за допомогою звичайної операції, закінчуємо програму з помилкою.
				panic("Tried to use the state of a regular operation")
			}
			// Інакше оновлюємо стан.
			p.state.Update(stateOp)
		}
	}
	// Завжди надсилаємо операцію зі станом у цикл подій, на першому місці.
	res = append([]painter.Operation{p.state}, res...)

	return res, nil
}

type cmdType int32

const (
	regular  cmdType = 0
	stateful cmdType = 1
	absent   cmdType = 2
)

func parse(cmd string) (painter.Operation, cmdType) {
	//Розділяємо строку на окремі текстові строки команди за пропусками.
	fields := strings.Fields(cmd)
	switch fields[0] {
	case "white":
		return painter.OperationFill{Color: color.White}, stateful
	case "green":
		return painter.OperationFill{Color: color.RGBA{G: 0xff, A: 0xff}}, stateful
	case "rect":
		return painter.OperationRect, regular
	case "update":
		return painter.UpdateOp, regular
	default:
		return nil, absent
	}
}
