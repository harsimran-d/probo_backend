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
var stockRegexp *regexp.Regexp

func init() {
	var err error
	userReg, err = regexp.Compile(`^[a-z][a-z0-9]{5,14}$`)
	if err != nil {
		log.Fatal("error compiling user regex in probo hanlders")
	}
	stockRegexp, err = regexp.Compile(`^[A-Z][A-Z0-9_]{4,20}$`)
	if err != nil {
		log.Fatal("error compiling stockRegexp in probo handlers")
	}
}

func NewProboHandler() *ProboHandler {
	ob := models.Orderbook{
		Mu:     sync.Mutex{},
		Orders: make(map[models.MarketID]interface{}),
	}
	sb := models.StockBalances{
		Mu:       sync.Mutex{},
		Balances: make(map[models.StockID]interface{}),
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
	h.ib.Mu.Lock()
	defer h.ib.Mu.Unlock()
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

func (h *ProboHandler) CreateSymbol(ctx *gin.Context) {
	stockSymbol := ctx.Param("stockSymbol")
	stockSymbol = strings.ToUpper(strings.Trim(stockSymbol, " "))
	correct := stockRegexp.MatchString(stockSymbol)
	if !correct {
		log.Println("error matching the stock symbol: ", stockSymbol)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "the stock symbol is should be alphanumeric 5-20 characters"})
		return
	}
	h.sb.Mu.Lock()
	defer h.sb.Mu.Unlock()
	if _, exists := h.sb.Balances[models.StockID(stockSymbol)]; exists {
		log.Println("stock symbol already exists: ", stockSymbol)
		ctx.JSON(http.StatusConflict, gin.H{"error": "stock symbol already exists"})
		return
	}
	h.sb.Balances[models.StockID(stockSymbol)] = make(map[models.UserID]any)
	ctx.JSON(http.StatusCreated, gin.H{"message": "symbol created successfully"})
}
