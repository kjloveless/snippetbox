package main

import (
  "html/template"
  "io/fs"
  "path/filepath"
  "time"

  "github.com/kjloveless/snippetbox/internal/models"
  "github.com/kjloveless/snippetbox/ui"
)

// Create a humanDate function which returns a nicely formatted string
// representation of a time.Time object.
func humanDate(t time.Time) string {
  // Return the empty string if time has the zero value.
  if t.IsZero() {
    return ""
  }

  // Convert the time to UTC before formatting it.
  return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This
// is essentially a string-keyed map which acts as a lookup between the names
// of our custom template functions and the functions themselves.
var functions = template.FuncMap{
  "humanDate": humanDate,
}

// Define a templateData type to act as the holding structure for
// any dynamic data that we want to pass to our HTML templates.
// At the moment it only contains one field, but we'll add more to it
// as the build progresses.
type templateData struct {
  CurrentYear     int
  Snippet         models.Snippet
  Snippets        []models.Snippet
  Form            any
  Flash           string
  IsAuthenticated bool
  CSRFToken       string
}

func newTemplateCache() (map[string]*template.Template, error) {
  // Initialize a new map to act as the cache.
  cache := map[string]*template.Template{}

  // Use fs.Glob() to get a slice of all filepaths in the ui.Files embedded
  // filesystem which match the pattern 'html/pages/*.tmpl'. This essentially
  // gives us a slice of al the 'page' templates for the application, just like
  // before.
  pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
  if err != nil {
    return nil, err
  }

  // Loop through the page filepaths one-by-one.
  for _, page := range pages {
    // Extract the file name (like 'home.tmpl') from the full filepath
    // and assign it to the name variable.
    name := filepath.Base(page)

    // Create a slice containing the filepath patterns for the templates we
    // want to parse.
    patterns := []string{
      "html/base.tmpl",
      "html/partials/*.tmpl",
      page,
    }

    // Parse the base template file into a template set.
    // The template.FuncMap must be registered with the template set before you
    // call the ParseFiles() method. This means we have to use template.New()
    // to create an empty template set, use the Funcs() method to register the
    // template.FuncMap, and then parse the file as normal.
    // Use ParseFS() instead of ParseFiles() to parse the template files from
    // the ui.Files embedded filesystem.
    ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
    if err != nil {
      return nil, err
    }

    // Add the template set to the map, using the name of the page
    // (like 'home.tmpl') as the key.
    cache[name] = ts
  }

  // Return the map
  return cache, nil
}
