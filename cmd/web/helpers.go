package main

import (
  "bytes"
  "errors"
  "fmt"
  "net/http"
  "runtime/debug"
  "time"

  "github.com/go-playground/form/v4"
  "github.com/justinas/nosurf"
)

// Create a newTemplate() helper, which returns a templateData struct
// initialized with the current year. Note that we're not using the
// *http.Request parameter here at the moment, but we will do that later in the
// book.
func (app *application) newTemplateData(r *http.Request) templateData {
  return templateData{
    CurrentYear:  time.Now().Year(),
    // Add the flash message to the tempalte data, if one exists.
    Flash:        app.sessionManager.PopString(r.Context(), "flash"),
    // Add the authentication status to the template data.
    IsAuthenticated:  app.isAuthenticated(r),
    CSRFToken:        nosurf.Token(r),
  }
}

// Create a new decodePostForm() helper method. The second parameter here, dst,
// is the target destination that we want to decode the form data into.
func (app *application) decodePostForm(r *http.Request, dst any) error {
  // Call ParseForm() on the request, in the same way that we did in our
  // snippetCreatePost handler.
  err := r.ParseForm()
  if err != nil {
    return err
  }

  // Call Decode() on our decoder instance, passing the target destination as
  // the fist parameter.
  err = app.formDecoder.Decode(dst, r.PostForm)
  if err != nil {
    // If we try to use an invalid target destination, the Decode() method will
    // return an error with the type *form.InvalidDecodeError. We use
    // errors.As() tp check for this and raise a panic rather than returning
    // the error.
    var invalidDecoderError *form.InvalidDecoderError

    if errors.As(err, &invalidDecoderError) {
      panic(err)
    }

    // For all other errors, we return them as normal.
    return err
  }

  return nil
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
  // Retrieve the appropriate template set from the cache based on the page
  // name (like 'home.tmpl'). If no entry exists in the cache with provided
  // name, then create a new error and call the serverError() helper method
  // that we made earlier and return.
  ts, ok := app.templateCache[page]
  if !ok {
    err := fmt.Errorf("the template %s does not exist", page)
    app.serverError(w, r, err)
    return
  }

  // Initialize a new buffer.
  buf := new(bytes.Buffer)

  // Write the template to the buffer, instead of straight to the
  // http.ResponseWriter. If there's an error, call our serviceError() helper
  // and then return.
  err := ts.ExecuteTemplate(buf, "base", data)
  if err != nil {
    app.serverError(w, r, err)
    return
  }

  // If the template is written to the buffer without any errors, we are safe
  // to go ahead and write the HTTP status code to http.ResponseWriter.
  // Write out the provided HTTP status code ('200 OK', '400 Bad Request' etc).
  w.WriteHeader(status)

  // Write the contents of the buffer to the http.ResponseWriter. Note: this
  // is another time where we pass our http.ResponseWriter to a function that
  // takes an io.Writer.
  buf.WriteTo(w)
}

// The serverError helper writes a log entry at Error level (including the
// request method and URI as attributes), then sends a generic 500 Internal
// Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, 
  err error) {
  var (
    method  = r.Method
    uri     = r.URL.RequestURI()
    // Use debug.Stack() to get the stack trace. This returns a byte slice,
    // which we need to convert to a string so that it's readable in the log
    // entry.
    trace = string(debug.Stack())
  )

  app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
  http.Error(w, http.StatusText(http.StatusInternalServerError), 
    http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding
// description to the user. We'll use this later in the book to send responses
// like 400 "Bad Request" when there's a problem with the request that the user
// sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
  http.Error(w, http.StatusText(status), status)
}

// Return true if the current request is from an authenticated user, otherwise
// return false.
func (app *application) isAuthenticated(r *http.Request) bool {
  isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
  if !ok {
    return false
  }
  return isAuthenticated
}
