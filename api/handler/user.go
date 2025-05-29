package handler

import (
	"asynchronous/auth"
	"asynchronous/email"
	"asynchronous/model/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register godoc
// @Summary Register user
// @Description create new users
// @Tags auth
// @Param info body user.RegisterReq true "User info"
// @Success 200 {object} string "Token"
// @Failure 400 {object} string "Invalid data"
// @Failure 500 {object} string "Server error"
// @Router /auth/register [post]
func (h Handler) Register(c *gin.Context) {
	h.Log.Info("Register is starting")
	req := db.User{}
	if err := c.BindJSON(&req); err != nil {
		h.Log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !email.IsValidEmail(req.Email) {
		h.Log.Error("Invalid email")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}
	userId, err := h.User.Register(c, req)
	if err != nil {
		h.Log.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := auth.GenerateJWTToken(userId, string(req.Role))

	if err != nil {
		h.Log.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.Log.Info("Register ended")
	c.JSON(http.StatusOK, gin.H{
		"Token": token,
	})
}
