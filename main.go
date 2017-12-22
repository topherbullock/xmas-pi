package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type updateRequest struct {
	Blink *time.Duration `json:"blink"`
	On    *bool          `json:"on"`
}

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
		"update_light": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			// Unmarshal
			var req updateRequest
			err = json.Unmarshal(b, &req)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			if req.Blink != nil {
				blink := *req.Blink
				light.Blink(blink, sigs)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "blinking on %s duration", blink.String())
				return
			}

			if req.On != nil {
				light.StopBlinking()

				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "turning light %v", *req.On)
				if *req.On {
					light.On()
				}
				light.Off()
				return
			}
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

	go func() {
		fmt.Println("listening on 8080")
		log.Fatal(http.ListenAndServe(":8080", router))
		exit <- true
	}()

	fmt.Println("awaiting signal")
	<-exit
	fmt.Println("exiting")
}
