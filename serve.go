package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Static files
	r.Static("/static", "./static")

	//Template helpers
	r.SetFuncMap(template.FuncMap{
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"upper": func(s string) string { return strings.ToUpper(s) },
		"lower": func(s string) string { return strings.ToLower(s) },
		"eq":    func(a, b any) bool { return a == b },
	})

	// Load every .html in templates/
	r.LoadHTMLGlob("templates/*.html")
	log.Println("✔️  Templates Loaded:", r.HTMLRender.Instance("", nil) != nil)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Home",
			"Path":  c.Request.URL.Path,
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Login",
			"Path":  c.Request.URL.Path,
		})
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Register",
			"Path":  c.Request.URL.Path,
		})
	})

	r.GET("/problems", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":      "All Problems",
			"Path":       c.Request.URL.Path,
			"Problems":   []any{},
			"Page":       1,
			"TotalPages": 1,
			"Query":      c.Query("q"),
		})
	})

	r.GET("/submissions", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":       "My Submissions",
			"Path":        c.Request.URL.Path,
			"Submissions": []any{},
		})
	})

	if err := r.Run(":2020"); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
