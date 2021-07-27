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