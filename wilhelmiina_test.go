package wilhelmiina

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func getTestDatabase(t *testing.T) *gorm.DB {
	db, _ := InitDatabase(t.TempDir() + "/test.db")
	CreateTables(db)
	return db
}

func TestUser(test *testing.T) {
	test.Log("Testing user creation and fetching:")
	database := getTestDatabase(test)
	user, err := CreateUser("test.user", "Test", "User", "password1", Student, database)
	if err != nil {
		test.Fatal(err)
	}
	identicalUser, err := GetUser(user.UUID, database)
	if err != nil {
		test.Fatal(err)
	}
	if identicalUser.UUID != user.UUID {
		test.Fatal("Wrong UUID returned...")
	}
	if identicalUser.Firstname != user.Firstname || identicalUser.Surname != user.Surname {
		test.Fatal("Incorrect data returned from database")
	}

	res, err := validatePassword("password1", identicalUser.Password)
	if err != nil {
		test.Fatal(err)
	}

	if !res {
		test.Fatal("Password matching failed")
	}

	u, err := GetUser("wronguid", database)
	if err != ErrUserNotFound {
		test.Log(u.UUID)
		test.Fatal("User found even though it doesn't exist")
	}

}

func TestSubjects(test *testing.T) {
	db := getTestDatabase(test)

	// Fill database with test data:
	maa, err := CreateSubject("Pitkä Matematiikka", "MAA", "Opi laskemaan paljon", db)
	if err != nil {
		test.Fatal(err)
	}
	NewCourse("Polynomifunktiot", "MAA.2", "y=x²", maa.SubjectID, db)
	NewCourse("Geometria", "MAA.3", "ympyrä on pyöreä ja kolmio on tärkeä xd", maa.SubjectID, db)
	NewCourse("Vektorit", "MAA.4", "Haluux nähä mun vektorin suunnan ja suuruuden", maa.SubjectID, db)

	ai, err := CreateSubject("Äidinkieli", "AI", "Opi puhumaan", db)
	if err != nil {
		test.Fatal(err)
	}

	matikankurssit, err := GetCoursesForSubject(maa.SubjectID, db)
	if err != nil {
		test.Fatal(err)
	}

	if len(matikankurssit) != 3 {
		test.Fatal("Returned the wrong amount of courses")
	}

	if matikankurssit[0].CourseName == "" {
		test.Fatal("No good info returned")
	}

	c, err := GetCoursesForSubject(ai.SubjectID, db)
	test.Log(c)
	if err != ErrNoCoursesFound {
		test.Fatal("Returned some courses or error even though shouldn't have")
	}

	maa_fromdb, err := GetSubject(maa.SubjectID, db)
	if err != nil {
		test.Fatal(err)
	}

	if maa_fromdb.ShortName != maa.ShortName {
		test.Fatal("Subject names don't match even though they should")
	}

	_, err = GetSubject("nonexistant", db)
	if err != ErrSubjNotFound {
		test.Fatal("Wrong error or subject returned")
	}
}

func TestMessageCreationAndSending(test *testing.T) {
	uid1 := "1"
	uid2 := "2"
	uid3 := "3"

	database := getTestDatabase(test)

	SendMessage(uid1, []string{uid1, uid2, uid3}, "Lmao", "Muija antaa bj altaas lmao", "", database)
	m1, err := GetMessagesForId(uid2, database)
	if err != nil {
		test.Fatal(err)
	}
	m2, err := GetMessagesForId(uid3, database)
	if err != nil {
		test.Fatal(err)
	}
	if m1[0].MessageID != m2[0].MessageID {
		test.Fatal("Wrong message returned...")
	}

	msg, _ := SendMessage(uid2, []string{uid1, uid2}, "VERTTI FOKUS", "Mikä niil mukeil kestää", "", database)

	messages, err := GetMessagesForId(uid1, database)
	if err != nil {
		test.Fatal(err)
	}
	if len(messages) < 2 {
		test.Fatal("Amount of messages returned was wrong")
	}

	_, err = GetMessagesForId("nonexistant", database)
	if err == nil {
		test.Fatal("Did not return error for nonexistant user")
	}
	reply, err := SendMessage(uid1, []string{uid2, uid1}, "Äijä chillaa", "Tuun iha just...", msg.MessageID, database)
	if err != nil {
		test.Fatal(err)
	}
	replies, err := GetReplies(msg.MessageID, database)
	if err != nil {
		test.Fatal(err)
	}
	if reply.MessageID != replies[0].MessageID {
		test.Fatal("Reply ids do not match")
	}

	_, err = GetReplies("", database)
	if err != ErrInvalidMessageID {
		test.Fatal("Wrong error or no error returned for invalid messageid")
	}

	_, err = GetReplies("thisiddoesnotexist", database)
	if err != ErrNoMessagesFound {
		test.Fatal("Messages found even though there should be none")
	}
}

func TestGroups(test *testing.T) {
	db := getTestDatabase(test)

	teacherID := "t1"
	courseId := "1"

	u, _ := CreateUser("test.user", "Test", "User", "password1", Student, db)
	userID := u.UUID

	now := time.Now().Unix()
	tomorrow := time.Now().Add(time.Hour * 24).Unix()
	timedata := []GroupTimeData{
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 0,
		},
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 3,
		},
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 5,
		},
	}

	g, err := NewGroup("Group 1", courseId, now, tomorrow, timedata, db)
	if err != nil {
		test.Fatal(err)
	}

	err = g.AssingTeacher(teacherID, db)
	if err != nil {
		test.Fatal(err)
	}
	if g.TeacherID != teacherID {
		test.Fatal("AssingTeacher did not change teacherid")
	}

	_, err = CreateReservation(userID, g.GroupID, db)
	if err != nil {
		test.Fatal(err)
	}

	user_groups, err := GetUserGroups(userID, db)
	if err != nil {
		test.Fatal(err)
	}
	if user_groups[0].Name != g.Name {
		test.Fatal("Wrong group returned")
	}
	if len(user_groups) != 1 {
		test.Fatal("Wrong amount of groups returned")
	}

	_, err = GetUserGroups("nonexistantUserID", db)
	if err != ErrUserHasNoGroups {
		test.Fatal("Wrong error or amount of groups returned")
	}
}

func TestCourses(test *testing.T) {
	db := getTestDatabase(test)
	subjID := "s1"
	subjID2 := "s2"
	course, err := NewCourse("test", "te", "test_desc", subjID, db)

	if err != nil {
		test.Fatal(err)
	}

	_, err = NewCourse("test2", "te2", "test_desc2", subjID2, db)
	if err != nil {
		test.Fatal(err)
	}

	db_course, err := GetCourse(course.CourseID, db)
	if err != nil {
		test.Fatal(err)
	}
	if db_course.CourseName != course.CourseName {
		test.Fatal("Wrong course returned")
	}

	_, err = GetCourse("notid", db)
	if err != ErrCourseNotFound {
		test.Fatal("Wrong error or amount of courses found")
	}

}

func assert(val interface{}, expected interface{}, t *testing.T) {
	if val != expected {
		t.Fatalf("Assertion failed! %v was not %v", val, expected)
	}
}

func assert_not(val interface{}, expected interface{}, t *testing.T) {
	if val == expected {
		t.Fatalf("Assertion failed! %v was %v", val, expected)
	}
}

func TestWilhelmiina(t *testing.T) {
	db := getTestDatabase(t)

	admin, err := CreateUser("admin", "Arto", "Admini", "admin", Admin, db)
	assert(err, nil, t)

	adminFromDB, err := GetUserByUn("admin", db)
	assert(err, nil, t)
	assert(admin.UUID, adminFromDB.UUID, t)

	adminFromDB2, err := GetUser(admin.UUID, db)
	assert(err, nil, t)
	assert(admin.UUID, adminFromDB2.UUID, t)
	assert(adminFromDB2.UUID, adminFromDB.UUID, t)

	pw_cmp_res, err := admin.CheckPassword("admin")
	assert(err, nil, t)
	assert(pw_cmp_res, true, t)

	pw_cmp_res, err = admin.CheckPassword("admin2")
	assert(err, nil, t)
	assert(pw_cmp_res, false, t)

	moderator, err := CreateUser("moderator", "Make", "Moderaattori", "moderator", Moderator, db)
	assert(err, nil, t)

	teacher, err := CreateUser("teacher", "Olli", "Opettaja", "teacher", Teacher, db)
	assert(err, nil, t)

	student1, err := CreateUser("student1", "Oona", "Oppilas", "password1", Student, db)
	assert(err, nil, t)

	student2, err := CreateUser("student2", "Eetu", "Esimerkki", "password2", Student, db)
	assert(err, nil, t)

	_, err = CreateUser("student2", "Should", "fail", "password2", Student, db)
	assert_not(err, nil, t)

	maa, err := CreateSubject("Pitkä Matematiikka", "MAA", "Opi laskemaan paljon", db)
	assert(err, nil, t)
	mab, err := CreateSubject("Lyhyt Matematiikka", "MAB", "Opi laskemaan vähemmän", db)
	assert(err, nil, t)

	maa2, err := NewCourse("Paraabelit", "MAA2", "y=x²", maa.SubjectID, db)
	assert(err, nil, t)
	mab2, err := NewCourse("Geometria", "MAB2", "pi = 3", mab.SubjectID, db)
	assert(err, nil, t)

	timedata_maa2 := []GroupTimeData{
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 0,
		},
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 3,
		},
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 4,
		},
	}

	timedata_mab2 := []GroupTimeData{
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 1,
		},
		{
			StartTime:    int64(time.Hour * 8),
			EndTime:      int64(time.Hour * 9),
			DayOfTheWeek: 2,
		},
		{
			StartTime:    int64(time.Hour * 10),
			EndTime:      int64(time.Hour * 11),
			DayOfTheWeek: 3,
		},
	}
	maa2group, err := NewGroup("MAA2.1", maa2.CourseID, time.Now().Unix(), time.Now().Add(24*time.Hour*30).Unix(), timedata_maa2, db)
	assert(err, nil, t)

	err = maa2group.AssingTeacher(teacher.UUID, db)
	assert(err, nil, t)

	mab2group, err := NewGroup("MAB2.1", mab2.CourseID, time.Now().Unix(), time.Now().Add(24*time.Hour*30).Unix(), timedata_mab2, db)
	assert(err, nil, t)

	err = mab2group.AssingTeacher(teacher.UUID, db)
	assert(err, nil, t)

	maa_courses, err := maa.GetCourses(db)
	assert(err, nil, t)
	assert(len(maa_courses), 1, t)
	assert(maa_courses[0].CourseID, maa2.CourseID, t)

	maa2_groups, err := GetGroupsForCourse(maa2.CourseID, db)
	assert(err, nil, t)
	assert(len(maa2_groups), 1, t)

	_, err = student1.JoinGroup(maa2group.GroupID, db)
	assert(err, nil, t)

	_, err = student1.JoinGroup(mab2group.GroupID, db)
	assert(err, nil, t)

	_, err = student2.JoinGroup(mab2group.GroupID, db)
	assert(err, nil, t)

	testmsg, err := admin.SendMessage([]string{moderator.UUID}, "Test Message", "Testing message sending", "", db)
	assert(err, nil, t)

	modMessages, err := moderator.GetMessages(db)
	assert(err, nil, t)
	assert(len(modMessages), 1, t)
	assert(modMessages[0].MessageID, testmsg.MessageID, t)

	reply, err := moderator.SendMessage([]string{admin.UUID}, "Test reply", "Testing replying", modMessages[0].MessageID, db)
	assert(err, nil, t)

	replies, err := testmsg.GetReplies(db)
	assert(err, nil, t)
	assert(replies[0].MessageID, reply.MessageID, t)

	err = reply.Delete(db)
	assert(err, nil, t)

	_, err = testmsg.GetReplies(db)
	assert_not(err, nil, t)

	s1g, err := student1.GetGroups(db)
	assert(err, nil, t)
	assert(len(s1g), 2, t)

	s1t1, err := GetUser(s1g[0].TeacherID, db)
	assert(err, nil, t)
	assert(s1t1.UUID, teacher.UUID, t)

	mab2members, err := mab2group.GetUsers(db)
	assert(err, nil, t)
	assert(len(mab2members), 2, t)

	maa2members, err := maa2group.GetUsers(db)
	assert(err, nil, t)
	assert(len(maa2members), 1, t)

	err = admin.UpdatePassword("admin2", db)
	assert(err, nil, t)

	admin, err = GetUser(admin.UUID, db)
	assert(err, nil, t)

	res, err := admin.CheckPassword("admin")
	assert(err, nil, t)
	assert(res, false, t)

	res, err = admin.CheckPassword("admin2")
	assert(err, nil, t)
	assert(res, true, t)

	err = CancelReservation(student1.UUID, mab2group.GroupID, db)
	assert(err, nil, t)

	reservations, err := GetGroupReservations(mab2group.GroupID, db)
	assert(err, nil, t)
	assert(len(reservations), 1, t)

	err = mab.Delete(db)
	assert(err, nil, t)

	_, err = mab2group.GetUsers(db)
	assert_not(err, nil, t)
}
