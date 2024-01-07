package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	// r is router
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	r.GET("/posts", func(c *gin.Context) {
		posts := makeAPIRequest()
		if posts == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error making the request"})
			return
		}
		modifyPost := posts[0:5]

		c.JSON(http.StatusOK, modifyPost)
	})

	r.POST("test", func(ctx *gin.Context) {
		type User struct {
			Name     string `json:"name" binding:"required"`
			Password int    `json:"password" binding:"required"`
		}

		var form User

		err := ctx.ShouldBind(&form)
		if err == nil {
			fmt.Println(form.Name)
			fmt.Println(form.Password)
			ctx.JSON(http.StatusOK, form)
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	// Group is like router.route()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		fmt.Println("User: ", user)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "Unautorized"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

func makeAPIRequest() []Post {
	apiURL := "https://jsonplaceholder.typicode.com/posts"
	response, err := http.Get(apiURL)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": "Error making the request"})
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unexpected status code: %s", response.Status)})
		return nil
	}

	var posts []Post
	err = json.NewDecoder(response.Body).Decode(&posts)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding JSON"})
		return nil
	}

	return posts
}
