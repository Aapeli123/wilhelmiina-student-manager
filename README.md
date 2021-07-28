# Wilhelmiina Student Manager
## Student management database written in golang
## Current features:
* Create and delete users with different roles
* Create and delete subjects, courses for subjects and groups for courses
* Add users to groups
* Send and delete messages between users, reply to messages
* Users have passwords stored by hashing them with argon2id and salted using 128 byte salt
* Currently uses sqlite as database but it is really easy to change in db.go
## Pull requests welcome!

# Testing:
* Unit tests can be run using `go test .`

# Examples:
## Database Creation:
```go
package main

import 	(
    // Import wilhelmiina
    "github.com/Aapeli123/wilhelmiina-student-manager"

    // Import gorm and necessary database drivers
	"gorm.io/driver/sqlite"
    "gorm.io/gorm"

    ) 

func main() {
	db, err := gorm.Open(sqlite.Open("./databasepath.db"))
	if err != nil {
		panic(err)
	}
	err := wilhelmiina.CreateTables(db) // Migrate all object schemas to database
	if err != nil {
		panic(err)
	}
    // Database needs to be supplied to any method as the last parameter
}
```