package lang

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"strconv"
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
		op, err := parse(commandLine)

		if _, ok := err.(statefulError); ok {
			// Якщо операція впливає на стан, пробуємо перевести її під інтерфейс StatefulOperation.
			stateOp, ok := op.(painter.StatefulOperation)
			if !ok {
				// Якщо парсер хоче оновлення стану за допомогою звичайної операції, закінчуємо програму з помилкою.
				panic("Tried to use a regular operation as stateful")
			}
			// Інакше, оновлюємо стан.
			p.state.Update(stateOp)
		} else if err != nil {
			// Якщо виникла неопізнана помилка при обробці операції, повертаємо цю помилку.
			return nil, err
		} else {
			// Якщо операція звичайна, просто додаємо її у список до передачі в цикл.
			res = append(res, op)
		}
	}
	// Завжди надсилаємо операцію зі станом у цикл подій, на першому місці.
	res = append([]painter.Operation{p.state}, res...)

	return res, nil
}

// Помилка, що видається парсером, якщо відповідна команді операція потребує обробки стану.
type statefulError struct{}

func (e statefulError) Error() string {
	return "The operation requires state management"
}

func parse(cmd string) (painter.Operation, error) {
	//Розділяємо строку на окремі текстові строки команди за пропусками.
	fields := strings.Fields(cmd)
	switch fields[0] {
	case "white":
		return painter.OperationFill{Color: color.White}, statefulError{}
	case "green":
		return painter.OperationFill{Color: color.RGBA{G: 0xff, A: 0xff}}, statefulError{}
	case "bgrect":
		args, err := processArguments(fields[1:], 4)
		if err != nil {
			return nil, err
		}
		return painter.OperationBGRect{
			Min: painter.RelativePoint{X: args[0], Y: args[1]},
			Max: painter.RelativePoint{X: args[2], Y: args[3]},
		}, statefulError{}
	case "update":
		return painter.UpdateOp, nil
	default:
		return nil, fmt.Errorf("Unknown command")
	}
}

func processArguments(args []string, requiredLen int) ([]float64, error) {
	if len(args) != requiredLen {
		return nil, fmt.Errorf("Invalid argument count")
	}
	var processed []float64
	for idx, arg := range args {
		num, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid argument at pos %d", idx)
		}
		if num >= 0 && num <= 1 {
			processed = append(processed, num)
		} else {
			return nil, fmt.Errorf("Value at pos %d is not in [0,1] range", idx)
		}
	}
	return processed, nil
}
