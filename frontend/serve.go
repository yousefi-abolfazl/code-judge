package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
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
		registered := c.Query("registered") == "true"
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":      "Login",
			"Path":       c.Request.URL.Path,
			"ApiURL":     apiURL,
			"Registered": registered,
		})
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":  "Register",
			"Path":   c.Request.URL.Path,
			"ApiURL": apiURL,
		})
	})

	r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		email := c.PostForm("email")
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")

		if password != confirmPassword {
			c.HTML(http.StatusBadRequest, "base.html", gin.H{
				"Title":  "Register",
				"Path":   c.Request.URL.Path,
				"Error":  "Passwords do not match",
				"ApiURL": apiURL,
			})
			return
		}

		requestBody, err := json.Marshal(map[string]string{
			"username": username,
			"email":    email,
			"password": password,
		})
		if err != nil {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Register",
				"Path":   c.Request.URL.Path,
				"Error":  "Error processing your request",
				"ApiURL": apiURL,
			})
			return
		}

		apiEndpoint := fmt.Sprintf("%s/api/auth/register", apiURL)
		resp, err := http.Post(
			apiEndpoint,
			"application/json",
			bytes.NewBuffer(requestBody),
		)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Register",
				"Path":   c.Request.URL.Path,
				"Error":  "Cannot connect to authentication service",
				"ApiURL": apiURL,
			})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Register",
				"Path":   c.Request.URL.Path,
				"Error":  "Error reading response from server",
				"ApiURL": apiURL,
			})
			return
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			var errorResponse map[string]interface{}
			json.Unmarshal(body, &errorResponse)

			errorMsg := "Registration failed"
			if errMsg, ok := errorResponse["error"]; ok {
				errorMsg = fmt.Sprintf("%v", errMsg)
			}

			c.HTML(resp.StatusCode, "base.html", gin.H{
				"Title":  "Register",
				"Path":   c.Request.URL.Path,
				"Error":  errorMsg,
				"ApiURL": apiURL,
			})
			return
		}

		c.Redirect(http.StatusFound, "/login?registered=true")
	})

	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "" || password == "" {
			c.HTML(http.StatusBadRequest, "base.html", gin.H{
				"Title":  "Login",
				"Path":   c.Request.URL.Path,
				"Error":  "Username and password are required",
				"ApiURL": apiURL,
			})
			return
		}

		requestBody, err := json.Marshal(map[string]string{
			"username": username,
			"password": password,
		})
		if err != nil {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Login",
				"Path":   c.Request.URL.Path,
				"Error":  "Error processing your request",
				"ApiURL": apiURL,
			})
			return
		}

		apiEndpoint := fmt.Sprintf("%s/api/auth/login", apiURL)
		resp, err := http.Post(
			apiEndpoint,
			"application/json",
			bytes.NewBuffer(requestBody),
		)

		if err != nil {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Login",
				"Path":   c.Request.URL.Path,
				"Error":  "Cannot connect to authentication service",
				"ApiURL": apiURL,
			})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Login",
				"Path":   c.Request.URL.Path,
				"Error":  "Error reading response from server",
				"ApiURL": apiURL,
			})
			return
		}

		if resp.StatusCode != http.StatusOK {
			var errorResponse map[string]interface{}
			json.Unmarshal(body, &errorResponse)

			errorMsg := "Login failed. Invalid username or password."
			if errMsg, ok := errorResponse["error"]; ok {
				errorMsg = fmt.Sprintf("%v", errMsg)
			}

			c.HTML(resp.StatusCode, "base.html", gin.H{
				"Title":  "Login",
				"Path":   c.Request.URL.Path,
				"Error":  errorMsg,
				"ApiURL": apiURL,
			})
			return
		}

		var loginResponse map[string]interface{}
		if err := json.Unmarshal(body, &loginResponse); err != nil {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Login",
				"Path":   c.Request.URL.Path,
				"Error":  "Error processing server response",
				"ApiURL": apiURL,
			})
			return
		}

		token, ok := loginResponse["token"].(string)
		if !ok {
			c.HTML(http.StatusInternalServerError, "base.html", gin.H{
				"Title":  "Login",
				"Path":   c.Request.URL.Path,
				"Error":  "Invalid authentication token received",
				"ApiURL": apiURL,
			})
			return
		}

		c.SetCookie(
			"auth_token",
			token,
			3600*24,
			"/",
			"",
			false,
			true,
		)

		c.Redirect(http.StatusFound, "/problems")
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
