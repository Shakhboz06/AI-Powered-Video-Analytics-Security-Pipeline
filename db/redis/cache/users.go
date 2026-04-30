package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"


	"github.com/redis/go-redis/v9"
)

type UsersStore struct {
	rbd *redis.Client
}
type Users struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
}

type Password struct {
	text *string
	hash []byte
}

const ExpTime = time.Minute * 2

func (s *UsersStore) Get(ctx context.Context, userID int64)(*Users, error){
	cacheKey := fmt.Sprintf("user-%v", userID)

	data, err := s.rbd.Get(ctx, cacheKey).Result()

	if err != nil {
		return nil, err
	}

	var user Users
	if data != ""{
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, err
}

func (s *UsersStore) Set(ctx context.Context, user *Users) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)

	if err != nil {
		return  err
	}

	return s.rbd.Set(ctx, cacheKey, json, ExpTime).Err()
}	