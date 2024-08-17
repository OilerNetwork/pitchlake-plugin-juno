package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

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

	l, err := net.Listen("tcp", "50555")
	if err != nil {
		return err
	}
	log.Printf("listening on ws://%v", l.Addr())
	ws := NewWebsocket()
	s := &http.Server{
		Handler:      ws,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	// sigs := make(chan os.Signal,2a 1)
	// signal.Notify(sigs, os.Interrupt)
	// select {
	// case err := <-errc:
	// 	log.Printf("failed to serve: %v", err)
	// case sig := <-sigs:
	// 	log.Printf("terminating: %v", sig)
	// }

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
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

					//If the event is a state transition, update locked unlocked balances

					//If the event is deposit/withdraw update lp balance
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
