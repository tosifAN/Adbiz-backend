package handlers

import (
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type FavHandler struct {
	db        *gorm.DB
	rateLimit *time.Ticker
}

func NewFevHandler(db *gorm.DB) *FavHandler {
	requestsPerSecond, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS_PER_SECOND"))
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10 // Default value
	}
	return &FavHandler{
		db:        db,
		rateLimit: time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
	}
}
