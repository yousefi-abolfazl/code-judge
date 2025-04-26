package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Get API URL from environment or use default
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://backend:8080"
	}

	log.Printf("Using API URL: %s", apiURL)

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
			"Title":  "Home",
			"Path":   c.Request.URL.Path,
			"ApiURL": apiURL,
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":  "Login",
			"Path":   c.Request.URL.Path,
			"ApiURL": apiURL,
		})
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":  "Register",
			"Path":   c.Request.URL.Path,
			"ApiURL": apiURL,
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
			"ApiURL":     apiURL,
		})
	})

	r.GET("/submissions", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":       "My Submissions",
			"Path":        c.Request.URL.Path,
			"Submissions": []any{},
			"ApiURL":      apiURL,
		})
	})

	if err := r.Run(":2020"); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
