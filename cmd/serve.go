package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Register custom template functions
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0 // Avoid division by zero
			}
			return a / b
		},
		"eq": func(a, b any) bool {
			return a == b
		},
		"lt": func(a, b any) bool {
			return a.(int) < b.(int)
		},
		"gt": func(a, b any) bool {
			return a.(int) > b.(int)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	})
	r.LoadHTMLGlob("templates/*")

	// Serve static files (optional, if you have CSS/JS/images)
	// r.Static("/static", "./static")

	// ------------------------
	// ROUTES
	// ------------------------

	// Index page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Home",
		})
	})

	// Submissions page
	r.GET("/submissions", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":       "My Submissions",
			"Submissions": []any{}, // fake for now
		})
	})

	// Login Page
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Login",
		})
	})

	// Handle login form submission
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		// Here you would typically check the username and password against a database
		// For now, just redirect to the profile page
		if username == "admin" && password == "password" { // fake check
			c.Redirect(http.StatusFound, "/profile")
		} else {
			c.HTML(http.StatusUnauthorized, "base.html", gin.H{
				"Title":   "Login",
				"Message": "Invalid username or password",
			})
		}
	})
	// Register Page
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Register",
		})
	})

	// Problems List
	r.GET("/problems", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":      "All Problems",
			"Problems":   []any{}, // placeholder
			"Page":       1,
			"TotalPages": 1,
			"Query":      "",
		})
	})

	// Problem Details (fake ID just to test)
	r.GET("/problems/1", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Two Sum",
			"Problem": gin.H{
				"ID":          1,
				"Title":       "Two Sum",
				"Statement":   "// Given an array of integers nums and a target...",
				"InputSpec":   "n\nnums...\ntarget",
				"OutputSpec":  "two indices",
				"Difficulty":  "Easy",
				"TimeLimit":   1,
				"MemoryLimit": 256,
			},
			"SupportedLangs":     []string{"go", "cpp", "java", "python"},
			"CSRFToken":          "",
			"LastSubmissionCode": "",
		})
	})

	// Profile Page
	r.GET("/profile", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": "Profile",
			"User": gin.H{
				"Username": "alireza",
				"Email":    "you@example.com",
			},
			"Stats": gin.H{
				"Solved": 12,
				"Total":  34,
				"Rank":   42,
			},
		})
	})
	r.Run(":8080")

}
