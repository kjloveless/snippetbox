package main

import (
  "errors"
  "fmt"
  "html/template"
  "net/http"
  "strconv"

  "github.com/kjloveless/snippetbox/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Server", "Go")

  snippets, err := app.snippets.Latest()
  if err != nil {
    app.serverError(w, r, err)
    return
  }

  // Initialize a slice containing the paths to the two files. It's important
  // to note that the file containing our base template must be the *first*
  // file in the slice.
  files := []string{
    "./ui/html/base.tmpl",
    "./ui/html/partials/nav.tmpl",
    "./ui/html/pages/home.tmpl",
  }

  // Use the template.ParseFiles() function to read the template file into a
  // template set. If there's an error, we log the detailed error message, use
  // the http.Error() function to send an Internal Server Error response to the
  // user, and then return from the handler so no subsequent code is executed.
  ts, err := template.ParseFiles(files...)
  if err != nil {
    app.serverError(w, r, err) // Use the serverError() helper.
    return
  }

  // Create an instance of a templateData struct holding the slice of snippets.
  data := templateData{
    Snippets: snippets,
  }

  // Then we use the Execute() method on the template set to write the
  // template content as the response body. The last parameters to Execute()
  // represents any dynamic data that we want to pass in.
  err = ts.ExecuteTemplate(w, "base", data)
  if err != nil {
    app.serverError(w, r, err) // Use the serverError() helper.
  }
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
  id, err := strconv.Atoi(r.PathValue("id"))
  if err != nil || id < 1 {
    http.NotFound(w, r)
    return
  }

  // Use the SnippetModel's Get() method to retrieve the data for a specific
  // record based on its ID. If no matching record is found, return a 404 Not
  // Found response.
  snippet, err := app.snippets.Get(id)
  if err != nil {
    if errors.Is(err, models.ErrNoRecord) {
      http.NotFound(w, r)
    } else {
      app.serverError(w, r, err)
    }
    return
  }

  // Initialize a slice containing the paths to the view.tmpl file,
  // plus the base layout and navigation partial that we made earlier.
  files := []string{
    "./ui/html/base.tmpl",
    "./ui/html/partials/nav.tmpl",
    "./ui/html/pages/view.tmpl",
  }

  // Parse the template files...
  ts, err := template.ParseFiles(files...)
  if err != nil {
    app.serverError(w, r, err)
    return
  }

  // Create an instance of a templateData struct holding the snippet data.
  data := templateData{
    Snippet: snippet,
  }

  // Pass in the templateData struct when executing the template.
  err = ts.ExecuteTemplate(w, "base", data)
  if err != nil {
    app.serverError(w, r, err)
  }
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Display a form for creating a new snippet..."))
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
  // Create some variables holding dummy data. We'll remove these later on 
  // during the build.
  title := "0 snail"
  content := "0 snail\nclimb mt fuji,\nbut slowly!\n\n- kobayashi issa"
  expires := 7

  // Pass the data to the SnippetModel.Insert() method, receiving the
  // ID of the new record back.
  id, err := app.snippets.Insert(title, content, expires)
  if err != nil {
    app.serverError(w, r, err)
    return
  }

  // Redirect the user to the relevant page for the snippet.
  http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
