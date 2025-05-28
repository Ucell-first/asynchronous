package middleware

import (
	"asynchronous/auth"
	"errors"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

type CasbinPermission interface {
	GetRole(*gin.Context) (string, int)
	CheckPermission(*gin.Context) (bool, error)
	CheckPermissionMiddleware() gin.HandlerFunc
}

type casbinPermission struct {
	enforcer *casbin.Enforcer
}

func NewCasbinPermission(enforcer *casbin.Enforcer) CasbinPermission {
	return &casbinPermission{enforcer: enforcer}
}

func Check(c *gin.Context) {

	refresh := c.GetHeader("Authorization")
	if refresh == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Authorization is required",
		})
		return
	}

	_, err := auth.ValidateToken(refresh)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token provided",
		})
		return
	}
	c.Next()
}

func (casb *casbinPermission) GetRole(c *gin.Context) (string, int) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return "unauthorized", http.StatusUnauthorized
	}
	_, role, err := auth.GetUserIdFromToken(token)
	if err != nil {
		return "error while reding role", 500
	}

	return role, 0
}

func (casb *casbinPermission) CheckPermission(c *gin.Context) (bool, error) {

	act := c.Request.Method
	sub, status := casb.GetRole(c)
	if status != 0 {
		return false, errors.New("error in get role")
	}
	obj := c.FullPath()

	ok, err := casb.enforcer.Enforce(sub, obj, act)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Internal server error",
		})
		c.Abort()
		return false, err
	}
	return ok, nil
}

func (casb *casbinPermission) CheckPermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := casb.CheckPermission(c)

		if err != nil {
			c.AbortWithError(500, err)
		}
		if !result {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "Forbidden",
			})
		}

		c.Next()
	}
}
