package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"goodies/goodies"
	"net/http"
	"os"
	"time"
)

type Command struct {
	Name   string
	Params []string
}

func handler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var cmd Command
	err := decoder.Decode(&cmd)
	if err != nil {
		fmt.Fprintf(w, formatError("REQUEST-PARSING-FAILED cannot parse request body"))
		return
	}
	fmt.Fprintf(w, handleCommand(cmd))
}

func handleCommand(cmd Command) string {
	return cmd.Params[0]
}

func formatError(err string) string {
	return fmt.Sprintf("error %v", err)
}

func main() {
	g := goodies.NewGoodies(1*time.Minute, "./goodies.dat", 30*time.Second)
	http.HandleFunc("/goodies", handler)
	http.ListenAndServe(":9006", nil) //9006 as for good

	fmt.Println("Enter any text to exit")
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Exiting...")
	g.Stop()
	<-time.After(5 * time.Second)
	fmt.Println("Bye")
}

// func createEcho(g *goodies.Goodies) *echo.Echo {
// 	e := echo.New()
// 	e.SetDebug(true)

// 	e.GET("/ping", func(c echo.Context) error {
// 		return c.String(http.StatusOK, "PONG")
// 	})

// 	e.GET("/values/:key", func(c echo.Context) error {
// 		value, found := g.Get(c.Param("key"))
// 		if found {
// 			return c.String(http.StatusOK, fmt.Sprintf("%#v", value))
// 		}
// 		return c.String(http.StatusNotFound, "Requested key not found")
// 	})

// 	e.POST("/values/", func(c echo.Context) error {
// 		key := c.FormValue("key")
// 		value := c.FormValue("value")
// 		ttlSec := c.FormValue("ttl_sec")
// 		var ttl time.Duration
// 		if ttlSec != "" {
// 			seconds, err := strconv.Atoi(ttlSec)
// 			if err == nil {
// 				ttl = time.Duration(seconds) * time.Second
// 			} else {
// 				return err
// 			}
// 		} else {
// 			ttl = goodies.ExpireDefault
// 		}

// 		g.Set(key, value, ttl)
// 		return c.String(http.StatusCreated, "Key created successfully")
// 	})

// 	e.DELETE("/values/:key", func(c echo.Context) error {
// 		key := c.Param("key")
// 		g.Remove(key)
// 		return c.String(http.StatusOK, "Success")
// 	})

// 	e.GET("/keys/", func(c echo.Context) error {
// 		res := "[" + strings.Join(g.Keys(), ";") + "]"
// 		return c.String(http.StatusOK, res)
// 	})
// 	return e
// }
