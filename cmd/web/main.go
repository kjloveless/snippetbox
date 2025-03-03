package main

import (
  "crypto/tls"
  "database/sql"
  "flag"
  "html/template"
  "log/slog"
  "net/http"
  "os"
  "time"

  // Import the models package that we just created. You need to prefix this
  // with whatever module path you set up back in chapter 02.01 (Project Setup
  // and Creating a Module) so that the import statement looks like this:
  // "{your-module-path}/internal/models". If you can't remember what module
  // path you used, you can find it at the top of the go.mod file.
  "github.com/kjloveless/snippetbox/internal/models"

  "github.com/alexedwards/scs/mysqlstore"
  "github.com/alexedwards/scs/v2"
  "github.com/go-playground/form/v4"
  _ "github.com/go-sql-driver/mysql"
)

// Define an application struct to hold the application-wide dependencies for
// the web application. For now we'll only include the structured logger, but
// we'll add more to this as the build progresses.
// Add a snippets field to the applicaion struct. This will allow us to make
// the SnippetModel object available to our handlers.
// Add a templateCache field to the application struct.
// Add a formDecoder field to hold a pointer to a form.Decoder instance.
type application struct {
  logger          *slog.Logger
  snippets        *models.SnippetModel
  users           *models.UserModel
  templateCache   map[string]*template.Template
  formDecoder     *form.Decoder
  sessionManager  *scs.SessionManager
}

func main() {
  // Define a new command-line flag with the name 'addr', a default value of
  // ":4000" and some short help text explaining what the flag controls. The
  // value of the flag will be stored in the addr variable at runtime.
  addr := flag.String("addr", ":4000", "HTTP network address")
  pass := flag.String("pass", "pass", "Password to use in DSN")
  // Define a new command-line flag for the MySQL DSN string.
  dsn := flag.String("dsn", "web:" + *pass + "@/snippetbox?parseTime=true", "MySQL data source name")

  // Importantly, we use the flag.Parse() function to parse the command-line
  // flag. This reads in the command-line flag value and assigns it to the addr
  // variable. You need to call this *before* you use the addr variable 
  // otherwise it will always contain the default value of ":4000". If any
  // errors are encountered during parsing the application will be terminated.
  flag.Parse()

  // Use the slog.New() function to initialize a new structured logger, which
  // writes to the standard out stream and uses the default settings.
  logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

  // To keep the main() function tidy I've put the code for creating a
  // connection pool into the separate openDB() function below. We pass
  // openDB() the DSN from the command-line flag.
  db, err := openDB(*dsn)
  if err != nil {
    logger.Error(err.Error())
    os.Exit(1)
  }

  // We also defer a call to db.Close(), so that the connection pool is closed
  // before the main() function exists.
  defer db.Close()

  // Initialize a new template cache...
  templateCache, err := newTemplateCache()
  if err != nil {
    logger.Error(err.Error())
    os.Exit(1)
  }

  // Initialize a decoder instance...
  formDecoder := form.NewDecoder()

  // Use the scs.New() function to initialize a new session manager. Then we
  // configure it to use our MySQL database as the session store, and set a
  // lifetime of 12 hours (so that sessions automatically expire 12 hours after
  // first being created).
  sessionManager := scs.New()
  sessionManager.Store = mysqlstore.New(db)
  sessionManager.Lifetime = 12 * time.Hour
  // Make sure the the Secure attribute is set on our session cookies.
  // Setting this means that the cookie will only be sent by a user's web
  // browser when a HTTPS connections is being used (and won't be sent over an
  // unsecure HTTP connection).
  sessionManager.Cookie.Secure = true

  // Initializes a new instance of our application struct, containing the 
  // dependencies (for now, just the structured logger).
  // Initializes a models.SnippetModel instance containing the connection pool
  // and add it to the application dependencies.
  app := &application{ 
    logger:         logger, 
    snippets:       &models.SnippetModel{DB: db},
    users:          &models.UserModel{DB: db},
    templateCache:  templateCache,
    formDecoder:    formDecoder,
    sessionManager: sessionManager,
  }

  // Initialize a tls.Config struct to hold the non-default TLS setttings we
  // want the server to use. In this case that only thing that we're changing
  // is the curve preferences value, so that only elliptic curves with assembly
  // implementations are used.
  tlsConfig := &tls.Config{
    CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
  }

  // Initialize a new http.Server struct. We set the Addr and Handler fields so
  // that the server uses the same network address and routes as before.
  srv := &http.Server{
    Addr:       *addr,
    Handler:    app.routes(),
    // Create a *log.Logger from our structured logger handler, which writes
    // log entries at Error level, and assigns it to the ErrorLog field. If
    // you would prefer to log the server errors at Warn level instead, you
    // could pass slog.LevelWarn as the final parameter.
    ErrorLog:   slog.NewLogLogger(logger.Handler(), slog.LevelError),
    TLSConfig:  tlsConfig,
    // Add Idle, Read, and Write timeouts to the server.
    IdleTimeout:    time.Minute,
    ReadTimeout:    5 * time.Second,
    WriteTimeout:   10 * time.Second,
  }

  // Use the Info() method to log the starting server message at Info severity
  // (along with the listen address as an attribute).
  logger.Info("starting server", "addr", srv.Addr)

  // Use the ListenAndServeTLS() method to start the HTTPS server. We
  // pass in the paths to the TLS certificate and corresponding private key as
  // the two parameters.
  err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

  // And we also use the Error() method to log any error message returned by
  // http.ListenAndServe() at Error severity (with no additional attributes),
  // and then call os.Exit(1) to terminate the application with exit code 1.
  logger.Error(err.Error())
  os.Exit(1)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
  db, err := sql.Open("mysql", dsn)
  if err != nil {
    return nil, err
  }

  err = db.Ping()
  if err != nil {
    db.Close()
    return nil, err
  }

  return db, nil
}
