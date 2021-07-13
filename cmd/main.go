package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	t := &Template{
		templates: template.Must(template.New("Main").Funcs(template.FuncMap{
			"Capitalize": func(name string) string {
				return strings.Title(name)
			},
			"formatTime": func(t time.Time) string {
				return t.Format("2 January, 17:00")
			},
		}).ParseGlob("views/*.html")),
	}

	e := echo.New()

	e.Renderer = t
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} ${status} ${latency_human}\n",
	}))
	e.Use(middleware.Recover())

	e.Static("/static", "views/static")

	// routes
	// route to get responses
	e.GET("/", func(c echo.Context) error {
		ctx := context.Background()
		sa := option.WithCredentialsFile("./cmd/creds.json")
		app, err := firebase.NewApp(ctx, nil, sa)
		if err != nil {
			fmt.Println(err)
			// return err
		}
		client, err := app.Firestore(ctx)
		if err != nil {
			return err
		}
		defer client.Close()

		responses := []map[string]interface{}{}

		iter := client.Collection("responses").Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Failed to iterate: %v", err)
			}
			responses = append(responses, doc.Data())
		}

		for _, response := range responses {
			code := "{\n"
			for k, v := range response {
				code += fmt.Sprintf("\t\"%v\":\"%v\",\n", k, v)
			}
			code += "}"
			response["Display"] = code
		}

		temp := template.Must(template.New("Secondary").Funcs(template.FuncMap{
			"Capitalize": func(name string) string {
				return strings.Title(name)
			},
			"formatTime": func(t time.Time) string {
				loc, _ := time.LoadLocation("Asia/Calcutta")
				return t.In(loc).Format("2 January, 15:04")
			},
		}).ParseGlob("views/*.html"))
		return temp.ExecuteTemplate(c.Response().Writer, "index", responses)
		// return c.Render(http.StatusOK, "base", nil)
	})

	// route to post responses
	e.POST("/:username", func(c echo.Context) error {
		req := map[string]interface{}{}
		req["created"] = time.Now()
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		ctx := context.Background()
		sa := option.WithCredentialsFile("./cmd/creds.json")
		app, err := firebase.NewApp(ctx, nil, sa)
		if err != nil {
			fmt.Println(err)
			// return err
		}
		client, err := app.Firestore(ctx)
		if err != nil {
			return err
		}
		defer client.Close()
		_, _, err = client.Collection("responses").Add(ctx, req)
		if err != nil {
			return err
		}
		return c.HTML(http.StatusOK, "<h1>Your response has been recorded</h1>")
	})

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "1323"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
