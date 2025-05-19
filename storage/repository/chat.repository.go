package repository

import (
	"context"
	"database/sql"
	"log"
	"marketplace-messageing/storage"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

type Chat struct {
	ID           int64     `json:"id"`
	Participants [2]int64  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
	DeletedAt    time.Time `json:"updated_at"`
}

type ChatRepository struct {
	db *storage.DB
}

func NewChatRepository(db *storage.DB) *ChatRepository {
	return &ChatRepository{db}
}

func (repo *ChatRepository) CreateChat(ctx context.Context, chat *Chat) (*Chat, error) {
	query := repo.db.QueryBuilder.Insert("chat").
		Columns("participants").
		Values(chat.Participants).
		Suffix("RETURNING id, participants, created_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = repo.db.QueryRow(ctx, sql, args...).Scan(
		&chat.ID,
		&chat.Participants,
		&chat.CreatedAt,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return chat, nil
}

func (repo *ChatRepository) ChatExists(ctx context.Context, participants [2]int64) (bool, error) {
	query := repo.db.QueryBuilder.Select("*").
		From("chat").
		Where(
			sq.Or{
				sq.Eq{"participants": [2]int64{participants[0], participants[1]}},
				sq.Eq{"participants": [2]int64{participants[1], participants[0]}},
			},
		).
		Limit(1)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var chat Chat
	err = repo.db.QueryRow(ctx, sqlStr, args...).Scan(
		&chat.ID,
		&chat.Participants,
		&chat.CreatedAt,
		&chat.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
