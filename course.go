package wilhelmiina

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Course struct {
	CourseID          string `gorm:"primaryKey"`
	CourseName        string
	CourseNameShort   string
	CourseDescription string
	SubjectID         string
}

func NewCourse(courseName string, courseNameShort string, courseDesc string, subjectID string, db *gorm.DB) (Course, error) {
	courseID := uuid.New().String()
	course := Course{
		CourseID:          courseID,
		CourseName:        courseName,
		CourseNameShort:   courseNameShort,
		CourseDescription: courseDesc,
		SubjectID:         subjectID,
	}
	tx := db.Begin()
	tx.Create(&course)
	err := tx.Commit().Error
	if err != nil {
		return Course{}, err
	}
	return course, nil
}

var ErrCourseNotFound = errors.New("no course found matching id")

func GetCourse(courseID string, db *gorm.DB) (Course, error) {
	var res Course
	tx := db.First(&res, "course_id = ?", courseID)
	if tx.RowsAffected == 0 {
		return Course{}, ErrCourseNotFound
	}
	if err := tx.Error; err != nil {
		return Course{}, err
	}
	return res, nil
}

func (c *Course) SetDescription(newDesc string, db *gorm.DB) error {
	tx := db.Begin()
	tx.Model(c).Where("course_id = ?", c.CourseID).Update("course_description", newDesc)
	err := tx.Commit().Error
	if err != nil {
		return err
	}
	return nil
}

func (c *Course) SetName(newName string, db *gorm.DB) error {
	tx := db.Begin()
	tx.Model(c).Where("course_id = ?", c.CourseID).Update("course_name", newName)
	err := tx.Commit().Error
	if err != nil {
		return err
	}
	return nil
}
func (c *Course) SetShortName(newName string, db *gorm.DB) error {
	tx := db.Begin()
	tx.Model(c).Where("course_id = ?", c.CourseID).Update("course_name_short", newName)
	err := tx.Commit().Error
	if err != nil {
		return err
	}
	return nil
}

func (c *Course) Delete(db *gorm.DB) error {
	return DeleteCourse(c.CourseID, db)
}

// Deletes course and all groups related to it
func DeleteCourse(courseID string, db *gorm.DB) error {
	groups, err := GetGroupsForCourse(courseID, db)
	if err != nil {
		return err
	}
	for _, g := range groups {
		g.GroupInfo.Delete(db)
	}
	tx := db.Begin()
	tx.Delete(Course{}, "course_id = ?", courseID)
	err = tx.Commit().Error
	if err != nil {
		return err
	}
	return nil
}
