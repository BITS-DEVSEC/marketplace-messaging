package repository

import (
	"context"
	"log"
	"marketplace-messageing/storage"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Message struct {
	ID int64 `json:"id"`

	ChatID  int64     `json:"chat_id"`
	From    int64     `json:"from"`
	To      int64     `json:"to"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type MessageRepository struct {
	db *storage.DB
}

func NewMessageRepository(db *storage.DB) *MessageRepository {
	return &MessageRepository{db}
}

func (repo *MessageRepository) CreateMessage(ctx context.Context, msg *Message) (*Message, error) {
	query := repo.db.QueryBuilder.Insert("message").
		Columns("chat_id", `"from"`, `"to"`, "content").
		Values(msg.ChatID, msg.From, msg.To, msg.Content).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = repo.db.QueryRow(ctx, sql, args...).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.From,
		&msg.To,
		&msg.Content,
		&msg.Time,
	)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return msg, nil
}

func (repo *MessageRepository) GetChatLastMessage(ctx context.Context, chatID int64) (*Message, error) {
	query := repo.db.QueryBuilder.Select("*").
		From("message").
		Where(sq.Eq{"chat_id": chatID}).
		OrderBy("time DESC").
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var msg Message
	err = repo.db.QueryRow(ctx, sql, args...).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.From,
		&msg.To,
		&msg.Content,
		&msg.Time,
	)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (repo *MessageRepository) GetAllMessages(ctx context.Context, chatID int64) ([]Message, error) {
	messages := []Message{}
	var msg Message

	query := repo.db.QueryBuilder.Select("*").
		From("message").
		Where(sq.Eq{"chat_id": chatID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := repo.db.Query(ctx, sql, args...)

	for result.Next() {
		err = result.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.From,
			&msg.To,
			&msg.Content,
			&msg.Time,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}
