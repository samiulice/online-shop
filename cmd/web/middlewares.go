package main

import (
	"net/http"
	"strings"
)

// SessionLoad loads and saves the session in every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

//Auth check for authenticated user
func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "user_id") || !app.Session.Exists(r.Context(), "account_type") {
			http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
			return
		} else {
			acc := app.Session.Get(r.Context(), "account_type").(string)
			if acc == "" || !strings.Contains(r.RequestURI, acc[:len(acc)-1]){
				http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
