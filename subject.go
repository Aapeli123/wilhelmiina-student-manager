package wilhelmiina

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subject struct {
	SubjectID   string `gorm:"primaryKey"`
	SubjectName string
	ShortName   string
}

func CreateSubject(name string, shortname string, db *gorm.DB) (Subject, error) {
	id := uuid.New().String()
	s := Subject{
		SubjectID:   id,
		SubjectName: name,
		ShortName:   shortname,
	}

	tx := db.Begin()
	tx.Create(&s)
	err := tx.Commit().Error

	if err != nil {
		return Subject{}, err
	}

	return s, nil
}

var ErrNoCoursesFound = errors.New("no courses found for this subject")

// Deletes Subject and all courses for it:
func DeleteSubject(subjectID string, db *gorm.DB) error {
	courses, err := GetCoursesForSubject(subjectID, db)
	if err != nil {
		return err
	}
	for _, c := range courses {
		c.Delete(db)
	}
	tx := db.Begin()
	tx.Where("subject_id = ?", subjectID).Delete(Subject{})
	return tx.Commit().Error
}

var ErrNoSubjectsFound = errors.New("no subjects found")

func GetSubjects(db *gorm.DB) ([]Subject, error) {
	var data []Subject
	tx := db.Model(&Subject{}).Select("*").Scan(&data)
	err := tx.Error
	if err != nil {
		return nil, err
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNoSubjectsFound
	}
	return data, nil
}

func GetCoursesForSubject(subjectID string, db *gorm.DB) ([]Course, error) {
	var res []Course
	result := db.Model(&Subject{}).Select("*").Where("subjects.subject_id = ?", subjectID).Joins("INNER JOIN courses on courses.subject_id = subjects.subject_id").Scan(&res)
	err := result.Error
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, ErrNoCoursesFound
	}
	return res, nil
}

func (s *Subject) GetCourses(db *gorm.DB) ([]Course, error) {
	return GetCoursesForSubject(s.SubjectID, db)
}

var ErrSubjNotFound = errors.New("subject not found")

func GetSubject(id string, db *gorm.DB) (Subject, error) {
	var res Subject
	tx := db.First(&res, "subject_id = ?", id)
	if tx.RowsAffected == 0 {
		return Subject{}, ErrSubjNotFound
	}
	return res, nil
}

func (s *Subject) Delete(db *gorm.DB) error {
	return DeleteSubject(s.SubjectID, db)
}
