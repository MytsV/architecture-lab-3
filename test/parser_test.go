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
		err  string
	}

	var testTable = []testCase{
		{
			name: "duplicate command",
			cmd:  "white white",
			err:  "Invalid argument count",
		},
		{
			name: "invalid argument amount",
			cmd:  "reset white",
			err:  "Invalid argument count",
		},
		{
			name: "uknown command",
			cmd:  "hello",
			err:  "Unknown command",
		},
		{
			name: "multiple command: uknown command",
			cmd:  "green\n bgrect 0.1 0.1 0.1 0.1\n hello",
			err:  "Unknown command",
		},
		{
			name: "out of the range",
			cmd:  "move 3 3",
			err:  "Value at pos 0 is not in [-1,1] range",
		},
		{
			name: "out of the range",
			cmd:  "bgrect 0.3 -8 0.5 0.3",
			err:  "Value at pos 1 is not in [-1,1] range",
		},
		{
			name: "invalid argument",
			cmd:  "figure j -0.9",
			err:  "Invalid argument at pos 0",
		},
		{
			name: "invalid amount of argument",
			cmd:  "bgrect 0.3 a",
			err:  "Invalid argument count",
		},
		{
			name: "move: invalid amount of argument",
			cmd:  "move 0.3 3 3 3",
			err:  "Invalid argument count",
		},
		{
			name: "miltiple cmd: duplicate cmd",
			cmd:  "move 0.3 0.1\n white white",
			err:  "Invalid argument count",
		},
		{
			name: "figure: out of range",
			cmd:  "move 0.3 0.1\n white\n figure 1.2 0.1",
			err:  "Value at pos 0 is not in [-1,1] range",
		},
		{
			name: "invalid argument",
			cmd:  "reset\n figure j -0.9\n green\n update",
			err:  "Invalid argument at pos 0",
		},
	}

	for _, test := range testTable {
		p := &lang.Parser{}
		_, err := p.Parse(strings.NewReader(test.cmd))

		assert.Equal(t, test.err, err.Error(), test.name)
	}
}

func TestParser_Operation(t *testing.T) {
	type testCase struct {
		name string
		cmd  string
		ops  []painter.Operation
	}

	testTable := []testCase{
		{
			name: "Simple consequent commands with state and without",
			cmd:  "white\ngreen\nupdate",
			ops: []painter.Operation{
				painter.StatefulOperationList{},
				painter.StatefulOperationList{},
				painter.UpdateOp,
			},
		},
		{
			name: "Figures are parsed with correct position",
			cmd:  "white\nfigure 0.5 0.5\nfigure -0.1 0.933",
			ops: []painter.Operation{
				painter.StatefulOperationList{},
				painter.StatefulOperationList{},
				painter.StatefulOperationList{},
			},
		},
	}

	for _, test := range testTable {
		p := &lang.Parser{}
		res, err := p.Parse(strings.NewReader(test.cmd))
		assert.Nil(t, err)
		assert.Equal(t, len(test.ops), len(res))
		for idx, op := range res {
			assert.IsType(t, test.ops[idx], op)
		}
	}
}
