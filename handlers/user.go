package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserHandlers struct {
	DB *pgxpool.Pool
}

func NewUserHandler(pool *pgxpool.Pool) *UserHandlers {
	return &UserHandlers{DB: pool}
}

func (h *UserHandlers) Signin(ctx *gin.Context) {
	var signInUser struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := ctx.ShouldBind(&signInUser)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "invalid username or password"})
		return
	}
	if signInUser.Username == "" || signInUser.Password == "" || len(signInUser.Username) > 15 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "invalid username or password"})
		return
	}

	row := h.DB.QueryRow(context.Background(), "SELECT id,username,hash FROM users where username=$1", signInUser.Username)
	var existingUser struct {
		Id       string
		Username string
		Hash     string
	}
	err = row.Scan(&existingUser.Id, &existingUser.Username, &existingUser.Hash)
	if err != nil {
		log.Println(err)
		ctx.Status(http.StatusNotFound)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Hash), []byte(signInUser.Password))
	if err != nil {
		log.Println("password matchin failed: ", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "invalid username or password"})
		return
	}

	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": existingUser.Id, "username": existingUser.Username})
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("error loading signing secret")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	signedToken, err := rawToken.SignedString(secret)
	ctx.JSON(http.StatusOK, gin.H{"token": signedToken})
}

func (h *UserHandlers) Signup(ctx *gin.Context) {
	var signUpUser struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := ctx.ShouldBind(&signUpUser)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad body payload"})
		return
	}
	if signUpUser.Username == "" || signUpUser.Password == "" || len(signUpUser.Username) > 15 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad body payload"})
		return
	}
	signUpUser.Username = strings.ToLower(signUpUser.Username)
	reg, err := regexp.Compile(`^[a-z][a-z0-9]{5,14}$`)
	if err != nil {
		log.Println("Error compiling regex:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}
	valid := reg.MatchString(signUpUser.Username)
	if !valid {
		log.Println("invalid email")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "the username must be alphanumeric and between 6-15 start with alphabets"})
		return
	}

	if len(signUpUser.Password) < 6 || len(signUpUser.Password) > 20 {
		log.Println("password too short")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "password lenght should be between 6 and 20"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(signUpUser.Password), 12)
	if err != nil {
		log.Println("error hasing password")
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}
	_, err = h.DB.Exec(context.Background(), "INSERT INTO USERS (username,hash) VALUES ($1,$2)", signUpUser.Username, hash)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == "23505" {
			log.Println("username already exists")
			ctx.JSON(http.StatusConflict, gin.H{"message": "username already exists"})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong while creating user"})
			return
		}
	}
	ctx.Status(http.StatusCreated)
}
