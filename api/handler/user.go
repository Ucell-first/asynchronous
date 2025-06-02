package handler

import (
	"asynchronous/auth"
	"asynchronous/email"
	"asynchronous/model/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterReq - Registratsiya so'rovi
type RegisterReq struct {
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Surname  string `json:"surname" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role,omitempty"`
}

// RegisterResp - Registratsiya javobi
type RegisterResp struct {
	Token string   `json:"token"`
	User  *db.User `json:"user"`
}

// Register godoc
// @Summary Register user
// @Description create new users
// @Tags auth
// @Param info body RegisterReq true "User info"
// @Success 200 {object} RegisterResp
// @Failure 400 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /auth/register [post]
func (h Handler) Register(c *gin.Context) {
	h.Log.Info("Register is starting")

	var req RegisterReq
	if err := c.BindJSON(&req); err != nil {
		h.Log.Error("Binding error: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Email validatsiyasi
	if !email.IsValidEmail(req.Email) {
		h.Log.Error("Invalid email")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Role validatsiya
	if req.Role != "" && req.Role != string(db.RoleAdmin) && req.Role != string(db.RoleWorker) {
		h.Log.Error("Invalid role: " + req.Role)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// User obyektini yaratish
	user := db.User{
		Email:        req.Email,
		Name:         req.Name,
		Surname:      req.Surname,
		PasswordHash: req.Password, // Service qismida hash qilinadi
		Role:         db.Role(req.Role),
	}

	// Registratsiya
	userID, err := h.User.Register(c, user)
	if err != nil {
		h.Log.Error("Registration error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	// User ma'lumotlarini olish
	registeredUser, err := h.User.GetUserProfile(c, userID)
	if err != nil {
		h.Log.Error("Get user error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Token generatsiya qilish
	token, err := auth.GenerateJWTToken(userID, string(registeredUser.Role))
	if err != nil {
		h.Log.Error("Token generation error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token creation failed"})
		return
	}

	// Parolni tozalash
	registeredUser.PasswordHash = ""

	h.Log.Info("Register ended successfully", "user_id", userID)
	c.JSON(http.StatusOK, RegisterResp{
		Token: token,
		User:  registeredUser,
	})
}
