package lang

import (
	"bufio"
	"io"
	"strings"

	"github.com/MytsV/architecture-lab-3/painter"
)

// Parser уміє прочитати дані з вхідного io.Reader та повернути список операцій представлені вхідним скриптом.
type Parser struct {
	state painter.OperationState
}

func (p *Parser) Parse(in io.Reader) ([]painter.Operation, error) {
	var res []painter.Operation

	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		commandLine := scanner.Text()
		op := parse(commandLine, &p.state) // Отримати відповідну до команди операцію
		if op != nil {
			res = append(res, op)
		}
	}
	res = append([]painter.Operation{p.state}, res...)

	return res, nil
}

func parse(line string, state *painter.OperationState) painter.Operation {
	fields := strings.Fields(line)
	switch fields[0] {
	case "white":
		state.Add(painter.OperationFunc(painter.WhiteFill))
		return nil
	case "green":
		state.Add(painter.OperationFunc(painter.GreenFill))
		return nil
	case "rect":
		return painter.OperationRect
	case "update":
		return painter.UpdateOp
	default:
		panic("Unknown command!")
	}
}
