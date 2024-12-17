package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bearer := ctx.GetHeader("Authorization")
		if bearer == "" {
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}
		parts := strings.Split(bearer, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SigningSecret), nil
		})
		if err != nil {
			fmt.Println("hitting this error here")
			log.Println(err)
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			sub := claims["sub"].(string)
			if sub == "" {
				ctx.Status(http.StatusUnauthorized)
				ctx.Abort()
				return
			}
			ctx.Set("sub", sub)
			ctx.Next()
		} else {
			fmt.Println()
			ctx.JSON(http.StatusInternalServerError, "somehting went wrong")
			ctx.Abort()
			return
		}
	}
}

func RejectTrailingSpaces() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasSuffix(path, " ") || strings.HasSuffix(path, "\t") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: Trailing spaces are not allowed in the URL.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
