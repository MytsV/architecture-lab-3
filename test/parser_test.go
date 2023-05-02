package test

import (
	"strings"
	"testing"

	"github.com/MytsV/architecture-lab-3/painter"
	"github.com/MytsV/architecture-lab-3/painter/lang"
	"github.com/stretchr/testify/assert"
)

func TestParser_Input(t *testing.T) {
	type testCase struct {
		name string
		cmd  string
		op   painter.Operation
		err  string
	}

	var testTable = []testCase{
		{
			name: "duplicate command",
			cmd:  "white white",
			op:   nil,
			err:  "Invalid argument count",
		},
		{
			name: "invalid argument amount",
			cmd:  "reset white",
			op:   nil,
			err:  "Invalid argument count",
		},
		{
			name: "uknown command",
			cmd:  "hello",
			op:   nil,
			err:  "Unknown command",
		},
		{
			name: "multiple command: uknown command",
			cmd:  "green\n bgrect 0.1 0.1 0.1 0.1\n hello",
			op:   nil,
			err:  "Unknown command",
		},
		{
			name: "out of the range",
			cmd:  "move 3 3",
			op:   nil,
			err:  "Value at pos 0 is not in [-1,1] range",
		},
		{
			name: "out of the range",
			cmd:  "bgrect 0.3 -8 0.5 0.3",
			op:   nil,
			err:  "Value at pos 1 is not in [-1,1] range",
		},
		{
			name: "invalid argument",
			cmd:  "figure j -0.9",
			op:   nil,
			err:  "Invalid argument at pos 0",
		},
		{
			name: "invalid amount of argument",
			cmd:  "bgrect 0.3 a",
			op:   nil,
			err:  "Invalid argument count",
		},
		{
			name: "move: invalid amount of argument",
			cmd:  "move 0.3 3 3 3",
			op:   nil,
			err:  "Invalid argument count",
		},
		{
			name: "miltiple cmd: duplicate cmd",
			cmd:  "move 0.3 0.1\n white white",
			op:   nil,
			err:  "Invalid argument count",
		},
		{
			name: "figure: out of range",
			cmd:  "move 0.3 0.1\n white\n figure 1.2 0.1",
			op:   nil,
			err:  "Value at pos 0 is not in [-1,1] range",
		},
		{
			name: "invalid argument",
			cmd:  "reset\n figure j -0.9\n green\n update",
			op:   nil,
			err:  "Invalid argument at pos 0",
		},
	}

	for _, test := range testTable {
		p := &lang.Parser{}
		_, err := p.Parse(strings.NewReader(test.cmd))

		assert.Equal(t, test.err, err.Error(), test.name)
	}
}
