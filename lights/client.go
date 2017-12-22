package lights

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/luismesas/goPi/MCP23S17"
)

type Light interface {
	IsOn() bool
	On()
	Off()
	Blink(interval time.Duration, done chan os.Signal)
	ToJSON() ([]byte, error)
}

type light struct {
	register *MCP23S17.MCP23S17RegisterBit
	statusMu sync.Mutex
	status   bool
}

func New(register *MCP23S17.MCP23S17RegisterBit) Light {
	register.AllOff()
	return &light{
		register: register,
	}
}

func (l *light) IsOn() bool {
	l.statusMu.Lock()
	defer l.statusMu.Unlock()
	return l.status
}

func (l *light) On() {
	l.statusMu.Lock()
	defer l.statusMu.Unlock()

	l.status = true
	l.register.AllOn()
}

func (l *light) Off() {
	l.statusMu.Lock()
	defer l.statusMu.Unlock()

	l.status = false
	l.register.AllOff()
}

func (l *light) Toggle() {
	if l.IsOn() {
		l.Off()
	} else {
		l.On()
	}
}

func (l *light) Blink(interval time.Duration, done chan os.Signal) {
	var exit bool
	timer := time.NewTicker(interval)
	for !exit {
		select {
		case _ = <-done:
			fmt.Println("received exit signal")
			l.Off()
			exit = true
		case _ = <-timer.C:
			l.Toggle()
		}
	}
}

func (l *light) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{"status": l.status})
}
