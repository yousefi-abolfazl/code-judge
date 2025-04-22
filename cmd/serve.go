package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 1) register template helpers
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"eq":    func(a, b any) bool { return a == b },
		"lt":    func(a, b any) bool { return a.(int) < b.(int) },
		"gt":    func(a, b any) bool { return a.(int) > b.(int) },
		"upper": func(s string) string { return strings.ToUpper(s) },
		"lower": func(s string) string { return strings.ToLower(s) },
	})
	r.LoadHTMLGlob("templates/*")

	// 2) define *all* routes *before* starting the server

	// Home
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Home",
		})
	})

	// Login (GET + POST)
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Login",
		})
	})
	r.POST("/login", func(c *gin.Context) {
		user := c.PostForm("username")
		pass := c.PostForm("password")
		if user == "admin" && pass == "password" {
			c.Redirect(http.StatusFound, "/profile")
		} else {
			c.HTML(http.StatusUnauthorized, "base.html", gin.H{
				"Title":   "Login",
				"Message": "Invalid credentials",
			})
		}
	})

	// Register (GET + POST)
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Register",
		})
	})
	r.POST("/register", func(c *gin.Context) {
		// TODO: actually create user
		c.Redirect(http.StatusFound, "/login")
	})

	// Problems list
	r.GET("/problems", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":      "All Problems",
			"Problems":   []any{}, // replace with your slice
			"Page":       1,
			"TotalPages": 1,
			"Query":      c.Query("q"),
		})
	})

	// Problem details
	r.GET("/problems/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Problem #" + id,
			"Problem": gin.H{
				"ID":          id,
				"Title":       "Sample Problem",
				"Statement":   "// problem text here",
				"InputSpec":   "…",
				"OutputSpec":  "…",
				"Difficulty":  "Easy",
				"TimeLimit":   1,
				"MemoryLimit": 256,
			},
			"SupportedLangs":     []string{"go", "cpp", "java", "python"},
			"CSRFToken":          "",
			"LastSubmissionCode": "",
		})
	})

	// My submissions
	r.GET("/submissions", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":       "My Submissions",
			"Submissions": []any{}, // replace with real data
		})
	})

	// Profile
	r.GET("/profile", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Profile",
			"User":  gin.H{"Username": "you", "Email": "you@example.com"},
			"Stats": gin.H{"Solved": 0, "Total": 0, "Rank": 0},
		})
	})

	// 3) now *finally* start the server once
	r.Run(":8080")
}
