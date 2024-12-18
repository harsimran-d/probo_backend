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
	sync.Mutex
	markets map[models.MarketID]models.Market
	sb      *models.StockBalances
	ib      *models.InrBalances
}

var userRegexp *regexp.Regexp
var marketRegexp *regexp.Regexp

func init() {
	var err error
	userRegexp, err = regexp.Compile(`^[a-z][a-z0-9]{5,14}$`)
	if err != nil {
		log.Fatal("error compiling user regex in probo hanlders")
	}
	marketRegexp, err = regexp.Compile(`^[A-Z][A-Z0-9_]{4,20}$`)
	if err != nil {
		log.Fatal("error compiling marketRegexp in probo handlers")
	}
}

func NewProboHandler() *ProboHandler {
	markets := make(map[models.MarketID]models.Market)

	sb := models.StockBalances{
		StockBalances: make(map[models.UserID]interface{}),
	}
	ib := models.InrBalances{
		InrBalances: make(map[models.UserID]models.UserBalance),
	}
	return &ProboHandler{
		markets: markets,
		sb:      &sb,
		ib:      &ib,
	}
}

func (h *ProboHandler) CreateUser(ctx *gin.Context) {
	userId := ctx.Param("userId")
	userId = strings.Trim(userId, " ")
	if userId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "enter valid id"})
		return
	}
	valid := userRegexp.MatchString(userId)
	if !valid {
		log.Println("invalid user id: ", userId)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "the username must be alphanumeric and between 6-15 start with alphabets"})
		return
	}
	h.ib.Lock()
	defer h.ib.Unlock()
	if _, exists := h.ib.InrBalances[models.UserID(userId)]; exists {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}
	h.ib.InrBalances[models.UserID(userId)] = models.UserBalance{
		Balance: 0,
		Locked:  0,
	}
	h.sb.Lock()
	defer h.sb.Unlock()
	h.sb.StockBalances[models.UserID(userId)] = make(map[models.MarketID]any)
	ctx.JSON(http.StatusCreated, gin.H{"message": "user created with 0 balances"})
}

func (h *ProboHandler) CreateMarket(ctx *gin.Context) {
	marketId := ctx.Param("marketId")
	marketId = strings.ToUpper(strings.Trim(marketId, " "))
	correct := marketRegexp.MatchString(marketId)
	if !correct {
		log.Println("invalid market symbol: ", marketId)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "the market symbol is should be all uppercase alphanumeric 5-20 characters"})
		return
	}
	h.sb.Lock()
	defer h.sb.Unlock()
	if _, exists := h.markets[models.MarketID(marketId)]; exists {
		log.Println("market symbol already exists: ", marketId)
		ctx.JSON(http.StatusConflict, gin.H{"error": "market symbol already exists"})
		return
	}

	h.markets[models.MarketID(marketId)] = models.Market{
		Book: &models.Orderbook{
			YesOrders: make(map[models.Price]interface{}),
			NoOrders:  make(map[models.Price]interface{}),
		},
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "symbol created successfully"})
}

func (h *ProboHandler) GetOrderBook(ctx *gin.Context) {
	marketId := ctx.Param("marketId")
	marketId = strings.ToUpper(strings.Trim(marketId, " "))
	market, exists := h.markets[models.MarketID(marketId)]
	if exists {
		ctx.JSON(http.StatusOK, market.Book)

	} else {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "order book not found"})
	}
}

func (h *ProboHandler) GetInrBalances(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.ib)
}
func (h *ProboHandler) GetStockBalances(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.sb)
}

func (h *ProboHandler) Reset(ctx *gin.Context) {
	h.Lock()
	h.sb.Lock()
	h.ib.Lock()
	defer h.Unlock()
	defer h.sb.Unlock()
	defer h.ib.Unlock()

	h.markets = make(map[models.MarketID]models.Market)
	h.sb.StockBalances = make(map[models.UserID]interface{})
	h.ib.InrBalances = make(map[models.UserID]models.UserBalance)
	ctx.Status(http.StatusOK)
}
