package wilhelmiina

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Group struct {
	GroupID   string `gorm:"primaryKey"`
	Name      string
	CourseID  string
	TeacherID string
	StartDate int64
	EndDate   int64
}

type GroupReservation struct {
	gorm.Model
	GroupID      string
	ReserverUUID string
}

type GroupTime struct {
	gorm.Model
	GroupID      string
	StartTime    int64
	EndTime      int64
	DayOfTheWeek int64
}

func CreateReservation(UUID string, GroupID string, db *gorm.DB) (GroupReservation, error) {
	reservation := GroupReservation{
		GroupID:      GroupID,
		ReserverUUID: UUID,
	}

	tx := db.Begin()
	tx.Create(&reservation)
	err := tx.Commit().Error

	if err != nil {
		return GroupReservation{}, err
	}
	return reservation, nil
}

func CancelReservation(UUID string, GroupID string, db *gorm.DB) error {
	tx := db.Begin()
	tx.Where("reserver_uuid = ? AND group_id = ?", UUID, GroupID).Delete(&GroupReservation{})
	err := tx.Commit().Error
	return err
}

func DeleteGroup(groupID string, db *gorm.DB) error {
	tx := db.Begin()
	tx.Where("group_id = ?", groupID).Delete(&GroupReservation{})
	tx.Where("group_id = ?", groupID).Delete(&GroupTime{})
	tx.Where("group_id = ?", groupID).Delete(&Group{})

	err := tx.Commit().Error
	return err
}

var ErrGroupNotFound = errors.New("group not found")

func GetGroupTimes(groupID string, db *gorm.DB) ([]GroupTimeData, error) {
	var data []GroupTimeData
	tx := db.Model(GroupTime{}).Where("group_id = ?", groupID).Scan(&data)
	if tx.RowsAffected == 0 {
		return nil, ErrGroupNotFound
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return data, nil
}

func UpdateGroupTimes(groupID string, newTD []GroupTimeData, db *gorm.DB) error {
	var groupTimes []GroupTime
	for _, gt := range newTD {
		time := GroupTime{
			GroupID:      groupID,
			StartTime:    gt.StartTime,
			EndTime:      gt.EndTime,
			DayOfTheWeek: gt.DayOfTheWeek,
		}
		groupTimes = append(groupTimes, time)
	}
	tx := db.Begin()
	tx.Where("group_id = ?", groupID).Delete(&GroupTime{})
	tx.Create(&groupTimes)
	err := tx.Commit().Error
	return err
}

func GetGroupReservations(groupID string, db *gorm.DB) ([]GroupReservation, error) {
	var data []GroupReservation
	tx := db.Model(GroupReservation{}).Where("group_id = ?", groupID).Scan(&data)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrGroupNotFound
	}
	return data, nil
}

type GroupTimeData struct {
	StartTime    int64
	EndTime      int64
	DayOfTheWeek int64
}

func NewGroup(name string, CourseID string, startDate int64, endDate int64, times []GroupTimeData, db *gorm.DB) (Group, error) {
	groupID := uuid.New().String()
	var groupTimes []GroupTime
	for _, gt := range times {
		time := GroupTime{
			GroupID:      groupID,
			StartTime:    gt.StartTime,
			EndTime:      gt.EndTime,
			DayOfTheWeek: gt.DayOfTheWeek,
		}
		groupTimes = append(groupTimes, time)
	}

	group := Group{
		GroupID:   groupID,
		Name:      name,
		CourseID:  CourseID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	tx := db.Begin()
	tx.Create(&group)
	tx.Create(&groupTimes)
	err := tx.Commit().Error
	if err != nil {
		return Group{}, err
	}
	return group, nil
}

func (g *Group) AssingTeacher(teacherID string, db *gorm.DB) error {
	prev := g.TeacherID
	g.TeacherID = teacherID
	tx := db.Begin()
	tx.Save(g)
	err := tx.Commit().Error
	if err != nil {
		g.TeacherID = prev
		return err
	}
	return nil
}

var ErrUserHasNoGroups = errors.New("user has no groups")

func GetUserGroups(UUID string, db *gorm.DB) ([]Group, error) {
	var result []Group
	tx := db.Model(&GroupReservation{}).
		Where("uuid = ?", UUID).Select("*").
		Joins("LEFT JOIN users ON users.uuid = group_reservations.reserver_uuid").
		Joins("LEFT JOIN groups ON groups.group_id = group_reservations.group_id").
		Scan(&result)

	err := tx.Error
	if err != nil {
		return nil, err
	}
	if tx.RowsAffected == 0 {
		return nil, ErrUserHasNoGroups
	}
	return result, nil
}

var ErrEmptyGroup = errors.New("group has no users")

func GetGroupUsers(groupID string, db *gorm.DB) ([]User, error) {
	var data []User
	tx := db.Model(&GroupReservation{}).
		Where("group_id = ?", groupID).Select("*").
		Joins("LEFT JOIN users ON users.uuid = group_reservations.reserver_uuid").
		Scan(&data)
	err := tx.Error
	if err != nil {
		return nil, err
	}
	if tx.RowsAffected == 0 {
		return nil, ErrEmptyGroup
	}
	return data, nil
}

func GetGroupsForCourse(courseID string, db *gorm.DB) ([]Group, error) {
	var data []Group
	tx := db.Model(Group{}).Where("course_id = ?", courseID).Scan(&data)
	err := tx.Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

type GroupData struct {
	GroupInfo  Group
	GroupTimes []GroupTimeData
}

func GetGroup(groupID string, db *gorm.DB) (GroupData, error) {
	var group Group
	tx := db.Model(Group{}).Where("group_id = ?", groupID).Scan(&group)
	err := tx.Error
	if err != nil {
		return GroupData{}, err
	}

	times, err := GetGroupTimes(groupID, db)
	if err != nil {
		return GroupData{}, err
	}
	return GroupData{
		GroupInfo:  group,
		GroupTimes: times,
	}, nil
}

func (g *Group) GetUsers(db *gorm.DB) ([]User, error) {
	return GetGroupUsers(g.GroupID, db)
}
func (g *Group) Delete(db *gorm.DB) error {
	return DeleteGroup(g.GroupID, db)
}
