package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// register template helper functions
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

	// ------------------------
	// ROUTES
	// ------------------------

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
		// TODO: real check
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
		// TODO: process registration
		c.Redirect(http.StatusFound, "/login")
	})

	// Problems list
	r.GET("/problems", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":      "All Problems",
			"Problems":   []any{},      // replace with real slice
			"Page":       1,            // current page
			"TotalPages": 1,            // total pages
			"Query":      c.Query("q"), // search query
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
				"InputSpec":   "input description",
				"OutputSpec":  "output description",
				"Difficulty":  "Easy",
				"TimeLimit":   1,
				"MemoryLimit": 256,
			},
			"SupportedLangs":     []string{"go", "cpp", "java", "python"},
			"CSRFToken":          "", // if using CSRF
			"LastSubmissionCode": "",
		})
	})

	// Submission history
	r.GET("/submissions", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":       "My Submissions",
			"Submissions": []any{}, // replace with real submissions slice
		})
	})

	// Profile
	r.GET("/profile", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Profile",
			"User": gin.H{
				"Username": "you",
				"Email":    "you@example.com",
			},
			"Stats": gin.H{
				"Solved": 0,
				"Total":  0,
				"Rank":   0,
			},
		})
	})

	// Start server (only ONE call)
	r.Run(":8080")
}
