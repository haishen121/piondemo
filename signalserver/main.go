package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/gin-gonic/gin"
)

func main() {

	var sdp sync.Map
	var candidates sync.Map

	var addr string
	flag.StringVar(&addr, "h", ":8080", ":8080")
	flag.Parse()

	g := gin.Default()
	g.SetTrustedProxies(nil)

	g.GET("/:user/all", func(c *gin.Context) {
		user := c.Param("user")
		s, ok := sdp.Load(user)
		if !ok {
			c.JSON(404, "")
			return
		}
		can, ok := candidates.Load(user)
		if !ok {
			c.JSON(404, "")
			return
		}
		fmt.Fprintf(c.Writer, "%s\n%s", s, can)
	})
	g.POST("/:user/all", func(c *gin.Context) {
		user := c.Param("user")

		all, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(404, "Not Found all")
			return
		}

		txt := bytes.SplitN(all, []byte("\n"), 1)
		if len(txt) != 2 {
			c.JSON(404, "Not Found all")
			return
		}
		sdp.Store(user, string(txt[0]))
		candidates.Store(user, string(txt[1]))

		c.JSON(200, "Add all")
	})

	g.GET("/:user/sdp", func(c *gin.Context) {
		user := c.Param("user")
		s, ok := sdp.Load(user)
		if !ok {
			c.JSON(404, "")
			return
		}

		fmt.Fprint(c.Writer, s)
	})

	g.POST("/:user/sdp", func(c *gin.Context) {
		user := c.Param("user")

		all, err := ioutil.ReadAll(c.Request.Body)
		if err != nil || len(all) == 0 {
			c.JSON(404, "Not Found sdp")
			return
		}
		sdp.Store(user, string(all))

		c.JSON(200, "Add sdp")
	})

	g.GET("/:user/candidates", func(c *gin.Context) {
		user := c.Param("user")
		candidate, ok := candidates.Load(user)
		if !ok {
			c.JSON(404, "")
			return
		}

		fmt.Fprint(c.Writer, candidate)
	})

	g.POST("/:user/candidates", func(c *gin.Context) {
		user := c.Param("user")

		all, err := ioutil.ReadAll(c.Request.Body)
		if err != nil || len(all) == 0 {
			c.JSON(404, "Not Found candidates")
			return
		}
		can := string(all)
		fmt.Printf("%s:%s\n<\n", user, can)
		candidates.Store(user, can)

		c.JSON(200, "Add candidates")
	})

	g.Run(addr)
}
