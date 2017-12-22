package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/luismesas/goPi/piface"
	"github.com/luismesas/goPi/spi"
	"github.com/tedsuo/rata"
	"github.com/topherbullock/xmas-pi/lights"
)

var lightRoutes = rata.Routes{
	{Name: "get_light", Method: rata.GET, Path: "/light"},
	{Name: "update_light", Method: rata.PUT, Path: "/light"},
}

var light lights.Light

func main() {
	pfd := piface.NewPiFaceDigital(spi.DEFAULT_HARDWARE_ADDR, spi.DEFAULT_BUS, spi.DEFAULT_CHIP)

	err := pfd.InitBoard()
	if err != nil {
		fmt.Printf("Error on init board: %s", err)
		return
	}

	light = lights.New(pfd.OutputPins[4])

	sigs := make(chan os.Signal, 1)
	exit := make(chan bool, 1)

	var lightHandlers = rata.Handlers{
		"get_light": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l, err := light.ToJSON()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err.Error())
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, string(l))
		}),
		"blink_light": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			light.Blink(time.Second, sigs)
		}),
	}

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		exit <- true
	}()

	router, err := rata.NewRouter(lightRoutes, lightHandlers)
	if err != nil {
		panic(err)
	}
	http.Handle("/api", router)
	log.Fatal(http.ListenAndServe(":8080", nil))

	fmt.Println("awaiting signal")
	<-exit
	fmt.Println("exiting")
}
