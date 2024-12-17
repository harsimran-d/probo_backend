package handlers

import (
	"log"
	"net/http"
	"probo_backend/models"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type ProboHandler struct {
	ob *models.Orderbook
	sb *models.StockBalances
	ib *models.InrBalances
}

var userReg *regexp.Regexp

func init() {
	var err error
	userReg, err = regexp.Compile(`^[a-z][a-z0-9]{5,14}$`)
	if err != nil {
		log.Fatal("error compiling user regex in probo hanlders")
	}
}

func NewProboHandler() *ProboHandler {
	ob := models.Orderbook{
		Mu:     sync.Mutex{},
		Orders: make(map[models.MarketID]interface{}),
	}
	sb := models.StockBalances{
		Mu:       sync.Mutex{},
		Balances: make(map[models.UserID]interface{}),
	}
	ib := models.InrBalances{
		Mu:       sync.Mutex{},
		Balances: make(map[models.UserID]models.UserBalance),
	}
	return &ProboHandler{
		ob: &ob,
		sb: &sb,
		ib: &ib,
	}
}

func (h *ProboHandler) CreateUser(ctx *gin.Context) {
	userId := ctx.Param("userId")
	h.ib.Mu.Lock()
	defer h.ib.Mu.Unlock()
	userId = strings.Trim(userId, " ")
	if userId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "enter valid id"})
		return
	}
	valid := userReg.MatchString(userId)
	if !valid {
		log.Println("invalid user id: ", userId)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "the username must be alphanumeric and between 6-15 start with alphabets"})
		return
	}
	if _, exists := h.ib.Balances[models.UserID(userId)]; exists {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}
	h.ib.Balances[models.UserID(userId)] = models.UserBalance{
		Balance: 0,
		Locked:  0,
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "user created with 0 balances"})
}
