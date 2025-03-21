package models

import (
  "testing"

  "github.com/kjloveless/snippetbox/internal/assert"
)

func TestUserModelExists(t *testing.T) {
  // Skip the test if the "-short" flag is provided when running the tests.
  if testing.Short() {
    t.Skip("models: skipping integration test")
  }

  // Set up a suite of table-driven tests and expected results.
  tests := []struct {
    name    string
    userID  int
    want    bool
  }{
    {
      name:   "valid ID",
      userID: 1,
      want:   true,
    },
    {
      name:   "zero id",
      userID: 0,
      want:   false,
    },
    {
      name:   "non-existent id",
      userID: 2,
      want:   false,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      // Call the newTestDB() helper function to get a connection pool to our
      // test database. Calling this here -- inside t.Run() -- means that fresh
      // database tables and data will be set up and torn down for each
      // sub-test.
      db := newTestDB(t)

      // Create a new instance of the UserModel.
      m := UserModel{db}

      // Call the UserModel.Exists() method and check that the return value and
      // error match the expected values for the sub-test.
      exists, err := m.Exists(tt.userID)

      assert.Equal(t, exists, tt.want)
      assert.NilError(t, err)
    })
  }
}
