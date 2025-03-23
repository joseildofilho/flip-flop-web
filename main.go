package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type LoginFormValue struct {
	Username string
	Password string
}

func main() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.Router.GET("/", apis.Static(os.DirFS("./pb_public"), false))
		e.Router.GET("/auth/", apis.Static(os.DirFS("./pb_public/auth"), false))

		group := e.Router.Group("/auth/login")
		group.POST("/", func(e *core.RequestEvent) error {
			e.App.Logger().Info("Login handler")
			form := &LoginFormValue{
				Username: e.Request.FormValue("username"),
				Password: e.Request.FormValue("password"),
			}
			e.App.Logger().Info(fmt.Sprintf("Login form: %s", form))
			user, err := e.App.FindAuthRecordByEmail("users", form.Username)
			if err != nil {
				e.App.Logger().Error(fmt.Sprintf("Error: %s", err))
				e.Next()
			}

			valid := user.ValidatePassword(form.Password)
			e.App.Logger().Info(fmt.Sprintf("Valid: %t", valid))
			if !valid {
				e.App.Logger().Error("Invalid password")
				return fmt.Errorf("Invalid password")
			}
			e.App.Logger().Info(fmt.Sprintf("User: %v", user))
			token, err := user.NewAuthToken()
			if err != nil {
				e.App.Logger().Error(fmt.Sprintf("Error: %s", err))
				return e.Next()
			}

			e.SetCookie(&http.Cookie{
				Name:     "session",
				Value:    token,
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
			})

			e.Response.Header().Add("HX-Location", "/home")
			return e.NoContent(204)
		})

		recordsGroups := e.Router.Group("/records")
		recordsGroups.GET("/", func(e *core.RequestEvent) error {
			tokenCookie, err := e.Request.Cookie("session")
			e.App.Logger().Info(fmt.Sprintf("Token: %s", tokenCookie))
			if err != nil {
				return e.Next()
			}

			token := tokenCookie.Value
			_, err = e.App.FindAuthRecordByToken(token)
			if err != nil {
				e.App.Logger().Error(fmt.Sprintf("Error: %s", err))
				return e.Next()
			}

			records, err := e.App.FindAllRecords("records")
			e.App.Logger().Info(fmt.Sprintf("Records: %v", records))
			if err != nil {
				e.App.Logger().Error(fmt.Sprintf("Error: %s", err))
				return e.Next()
			}

			var value string
			for _, record := range records {
				value += fmt.Sprintf("<tr><td>%s</td></tr>\n", record.GetString("started_at"))
			}
			e.App.Logger().Info(fmt.Sprintf("Value: %s", value))

			return e.HTML(200, value)
		})

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}
