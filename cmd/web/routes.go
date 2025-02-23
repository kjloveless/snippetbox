package main

import "net/http"

// The routes() method returns a servemux containing our application routes.
func (app *application) routes() http.Handler {
  mux := http.NewServeMux()

  fileServer := http.FileServer(http.Dir("./ui/static"))
  mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

  mux.HandleFunc("GET /{$}", app.home)
  mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
  mux.HandleFunc("GET /snippet/create", app.snippetCreate)
  mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)

  // Pass the servemux as the 'next' parameter to the commonHeaders middleware.
  // Because commonHeaders is just a function, and the function returns a
  // http.Handler we don't need to do anything else.
  return app.recoverPanic(app.logRequest(commonHeaders(mux)))
}
