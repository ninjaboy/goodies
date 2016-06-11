package main

import (
	"fmt"
	"goodies/goodies"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/fasthttp"
)

func main() {
	g := goodies.NewGoodies(1 * time.Minute)

	e := echo.New()
	e.SetDebug(true)
	e.GET("/values/:key", func(c echo.Context) error {
		value, found := g.Get(c.Param("key"))
		if found {
			return c.String(http.StatusOK, fmt.Sprintf("%#v", value))
		}
		return c.String(http.StatusNotFound, "Requested key not found\n")
	})

	e.POST("/values/", func(c echo.Context) error {
		key := c.FormValue("key")
		value := c.FormValue("value")
		ttlSec := c.FormValue("ttl_sec")
		var ttl time.Duration
		if ttlSec != "" {
			seconds, err := strconv.Atoi(ttlSec)
			if err == nil {
				ttl = time.Duration(seconds) * time.Second
			} else {
				return err
			}
		} else {
			ttl = goodies.ExpireDefault
		}

		g.Set(key, value, ttl)
		return c.String(http.StatusCreated, "Key created successfully\n")
	})

	e.DELETE("/values/:key", func(c echo.Context) error {
		key := c.Param("key")
		g.Remove(key)
		return c.String(http.StatusOK, "Success\n")
	})

	e.GET("/keys/", func(c echo.Context) error {
		res := "[" + strings.Join(g.Keys(), ";") + "]"
		return c.String(http.StatusOK, res)
	})

	e.Run(fasthttp.New(":6009"))

	fmt.Println("Bye")
}
