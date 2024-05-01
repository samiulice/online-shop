package main

import "net/http"

// SessionLoad loads and saves the session in every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

//Auth check for authenticated user
func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "user_id") {
			http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
