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

			e.SetCookie(&http.Cookie{
				Name:     "session",
				Value:    user.TokenKey(),
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
			})
			return e.Redirect(302, "/home")
		})

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}
