package lang

import (
	"bufio"
	"io"
	"strings"

	"github.com/MytsV/architecture-lab-3/painter"
)

// Parser уміє прочитати дані з вхідного io.Reader та повернути список операцій представлені вхідним скриптом.
type Parser struct {
}

func (p *Parser) Parse(in io.Reader) ([]painter.Operation, error) {
	var res []painter.Operation

	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		commandLine := scanner.Text()
		op := parse(commandLine) // Отримати відповідну до команди операцію
		res = append(res, op)
	}

	return res, nil
}

func parse(line string) painter.Operation {
	fields := strings.Fields(line)
	switch fields[0] {
	case "white":
		return painter.OperationFunc(painter.WhiteFill)
	case "green":
		return painter.OperationFunc(painter.GreenFill)
	case "update":
		return painter.UpdateOp
	default:
		panic("Unknown command!")
	}
}
