package wilhelmiina

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	MessageID  string `gorm:"primaryKey"`
	Title      string
	Contents   string
	From       string
	RespondsTo string
}

type MessageReciever struct {
	gorm.Model
	MessageID string
	UUID      string
}

func createRecieverList(messageID string, recievers []string) []MessageReciever {
	var list []MessageReciever
	for _, reciever := range recievers {
		list = append(list, MessageReciever{MessageID: messageID, UUID: reciever})
	}
	return list
}

// Creates and sends a message by saving it to database:
func SendMessage(from string, to []string, title string, contents string, respondsTo string, db *gorm.DB) (Message, error) {
	messageID := uuid.New().String()
	mr := createRecieverList(messageID, to)
	message := Message{
		MessageID:  messageID,
		Title:      title,
		Contents:   contents,
		From:       from,
		RespondsTo: respondsTo,
	}

	tx := db.Begin()
	db.Create(&message)
	db.Create(&mr)
	err := tx.Commit().Error

	if err != nil {
		return Message{}, err
	}
	return message, nil
}

var ErrNoMessagesFound = errors.New("no messages found")

func GetMessagesForId(uid string, db *gorm.DB) ([]Message, error) {
	var r []Message
	n := db.Model(&MessageReciever{}).Where("uuid = ?", uid).Select("*").Joins("left join messages on messages.message_id = message_recievers.message_id").Find(&r)
	if n.RowsAffected == 0 {
		return nil, ErrNoMessagesFound
	}
	return r, nil
}

var ErrInvalidMessageID = errors.New("invalid MessageID")

func GetReplies(messageid string, db *gorm.DB) ([]Message, error) {
	if messageid == "" {
		return nil, ErrInvalidMessageID
	}
	var r []Message
	n := db.Model(Message{}).Where("responds_to = ?", messageid).Find(&r)
	if n.RowsAffected == 0 {
		return nil, ErrNoMessagesFound
	}
	return r, nil
}

var ErrMessageNotFound = errors.New("message not found")

func GetMessage(id string, db *gorm.DB) (Message, error) {
	var m Message
	tx := db.First(&m, "message_id = ?", id)

	if tx.RowsAffected == 0 {
		return Message{}, ErrMessageNotFound
	}
	return m, nil
}

func DeleteMessage(messageID string, db *gorm.DB) error {

	tx := db.Begin()
	tx.Where("message_id = ?", messageID).Delete(&Message{})
	tx.Where("message_id = ?", messageID).Delete(&MessageReciever{})
	err := tx.Commit().Error

	return err
}

func (m *Message) GetReplies(db *gorm.DB) ([]Message, error) {
	return GetReplies(m.MessageID, db)
}

func (m *Message) Delete(db *gorm.DB) error {
	return DeleteMessage(m.MessageID, db)
}
