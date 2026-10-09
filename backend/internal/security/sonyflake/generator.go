package sonyflake

import (
	"fmt"

	sony "github.com/sony/sonyflake/v2"
)

type Generator struct {
	sonyflake *sony.Sonyflake
}

func NewGenerator(machineID int) (*Generator, error) {
	generator, err := sony.New(sony.Settings{
		MachineID: func() (int, error) {
			return machineID, nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Sonyflake generator: %w", err)
	}
	return &Generator{sonyflake: generator}, nil
}

func (g *Generator) NextID() (int64, error) {
	id, err := g.sonyflake.NextID()
	if err != nil {
		return 0, fmt.Errorf("generate Sonyflake ID: %w", err)
	}
	return id, nil
}
