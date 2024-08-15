package main

import (
	"fmt"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	junoplugin "github.com/NethermindEth/juno/plugin"
)

//go:generate go build -buildmode=plugin -o ../../build/plugin.so ./example.go
type pitchlakePlugin struct {
	vaultAddress felt.Felt
}

// Important: "JunoPluginInstance" needs to be exported for Juno to load the plugin correctly
var JunoPluginInstance = pitchlakePlugin{
	//Initialize with vaultAddress
}
var _ junoplugin.JunoPlugin = (*pitchlakePlugin)(nil)

func (p *pitchlakePlugin) Init() error {
	fmt.Println("ExamplePlugin initialised")
	return nil
}

func (p *pitchlakePlugin) Shutdown() error {
	fmt.Println("ExamplePlugin shutdown")
	return nil
}

func (p pitchlakePlugin) NewBlock(block *core.Block, stateUpdate *core.StateUpdate, newClasses map[felt.Felt]core.Class) error {
	fmt.Println("ExamplePlugin NewBlock called")
	for i := 0; i < len(block.Receipts); i++ {
		if block.Receipts[i].Events != nil {

			for j := 0; j < len(block.Receipts[i].Events); j++ {
				if block.Receipts[i].Events[j].From == &p.vaultAddress {
					//Event is from the contract, perform actions here
				}
			}
		}
	}
	return nil
}

func (p *pitchlakePlugin) RevertBlock(from, to *junoplugin.BlockAndStateUpdate, reverseStateDiff *core.StateDiff) error {
	fmt.Println("ExamplePlugin RevertBlock called")
	return nil
}
