package handler

import (
	"asynchronous/auth"
	"asynchronous/email"
	"asynchronous/model/db"
	"net/http"
	"strconv"

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

// LoginReq - Kirish so'rovi
type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp - Kirish javobi
type LoginResp struct {
	Token string   `json:"token"`
	User  *db.User `json:"user"`
}

// Login godoc
// @Summary Login user
// @Description authenticate users
// @Tags auth
// @Param credentials body LoginReq true "Login credentials"
// @Success 200 {object} LoginResp
// @Failure 400 {object} ErrorResp
// @Failure 401 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	h.Log.Info("Login is starting")

	var req LoginReq
	if err := c.BindJSON(&req); err != nil {
		h.Log.Error("Binding error: " + err.Error())
		c.JSON(http.StatusBadRequest, ErrorResp{Error: "Noto'g'ri so'rov formati"})
		return
	}

	// Autentifikatsiya
	user, err := h.User.Login(c, req.Email, req.Password)
	if err != nil {
		h.Log.Error("Authentication error: " + err.Error())
		c.JSON(http.StatusUnauthorized, ErrorResp{Error: "Kirish ma'lumotlari noto'g'ri"})
		return
	}

	// Token generatsiya qilish
	token, err := auth.GenerateJWTToken(user.ID, string(user.Role))
	if err != nil {
		h.Log.Error("Token generation error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Token yaratishda xato"})
		return
	}

	h.Log.Info("Login muvaffaqiyatli", "user_id", user.ID)
	c.JSON(http.StatusOK, LoginResp{
		Token: token,
		User:  user,
	})
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description get current user profile
// @Tags user
// @Security ApiKeyAuth
// @Success 200 {object} db.User
// @Failure 401 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /user/profile [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
	h.Log.Info("GetUserProfile is starting")

	// Kontekstdan userID ni olish
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResp{Error: "Foydalanuvchi autentifikatsiyadan o'tmagan"})
		return
	}

	// Profilni olish
	user, err := h.User.GetUserProfile(c, userID.(string))
	if err != nil {
		h.Log.Error("Get profile error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Profilni olishda xato"})
		return
	}

	h.Log.Info("Profil muvaffaqiyatli olindi", "user_id", userID)
	c.JSON(http.StatusOK, user)
}

// UpdateUserReq - Foydalanuvchini yangilash so'rovi
type UpdateUserReq struct {
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
	Surname string `json:"surname,omitempty"`
}

// UpdateUser godoc
// @Summary Update user profile
// @Description update user profile information
// @Tags user
// @Security ApiKeyAuth
// @Param updates body UpdateUserReq true "Update fields"
// @Success 200 {object} db.User
// @Failure 400 {object} ErrorResp
// @Failure 401 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /user [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	h.Log.Info("UpdateUser is starting")

	// Kontekstdan userID ni olish
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResp{Error: "Foydalanuvchi autentifikatsiyadan o'tmagan"})
		return
	}

	var req UpdateUserReq
	if err := c.BindJSON(&req); err != nil {
		h.Log.Error("Binding error: " + err.Error())
		c.JSON(http.StatusBadRequest, ErrorResp{Error: "Noto'g'ri so'rov formati"})
		return
	}

	// Yangilanishlarni mapga aylantirish
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Surname != "" {
		updates["surname"] = req.Surname
	}

	// Foydalanuvchini yangilash
	if err := h.User.UpdateUser(c, userID.(string), updates); err != nil {
		h.Log.Error("Update error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Foydalanuvchini yangilashda xato"})
		return
	}

	// Yangilangan foydalanuvchi ma'lumotlarini olish
	updatedUser, err := h.User.GetUserProfile(c, userID.(string))
	if err != nil {
		h.Log.Error("Get user error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Yangilangan foydalanuvchini olishda xato"})
		return
	}

	h.Log.Info("Foydalanuvchi ma'lumotlari yangilandi", "user_id", userID)
	c.JSON(http.StatusOK, updatedUser)
}

// UpdatePasswordReq - Parolni yangilash so'rovi
type UpdatePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UpdatePassword godoc
// @Summary Update user password
// @Description update user password
// @Tags user
// @Security ApiKeyAuth
// @Param passwords body UpdatePasswordReq true "Password details"
// @Success 200 {object} SuccessResp
// @Failure 400 {object} ErrorResp
// @Failure 401 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /user/password [put]
func (h *Handler) UpdatePassword(c *gin.Context) {
	h.Log.Info("UpdatePassword is starting")

	// Kontekstdan userID ni olish
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResp{Error: "Foydalanuvchi autentifikatsiyadan o'tmagan"})
		return
	}

	var req UpdatePasswordReq
	if err := c.BindJSON(&req); err != nil {
		h.Log.Error("Binding error: " + err.Error())
		c.JSON(http.StatusBadRequest, ErrorResp{Error: "Noto'g'ri so'rov formati"})
		return
	}

	// Parolni yangilash
	if err := h.User.ChangePassword(c, userID.(string), req.OldPassword, req.NewPassword); err != nil {
		h.Log.Error("Password update error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Parolni yangilashda xato"})
		return
	}

	h.Log.Info("Parol muvaffaqiyatli yangilandi", "user_id", userID)
	c.JSON(http.StatusOK, SuccessResp{Message: "Parol muvaffaqiyatli yangilandi"})
}

// ListUsers godoc
// @Summary List users
// @Description list all users (admin only)
// @Tags admin
// @Security ApiKeyAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} db.User
// @Failure 401 {object} ErrorResp
// @Failure 403 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /admin/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	h.Log.Info("ListUsers is starting")

	// Query parametrlarini olish
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Foydalanuvchilarni olish
	users, err := h.User.ListUsers(c, limit, offset)
	if err != nil {
		h.Log.Error("List users error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Foydalanuvchilarni olishda xato"})
		return
	}

	h.Log.Info("Foydalanuvchilar ro'yxati olindi", "count", len(users))
	c.JSON(http.StatusOK, users)
}

// UpdateUserRoleReq - Foydalanuvchi rolini yangilash so'rovi
type UpdateUserRoleReq struct {
	Role string `json:"role" binding:"required"`
}

// UpdateUserRole godoc
// @Summary Update user role
// @Description update user role (admin only)
// @Tags admin
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param role body UpdateUserRoleReq true "New role"
// @Success 200 {object} SuccessResp
// @Failure 400 {object} ErrorResp
// @Failure 401 {object} ErrorResp
// @Failure 403 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /admin/users/{id}/role [put]
func (h *Handler) UpdateUserRole(c *gin.Context) {
	h.Log.Info("UpdateUserRole is starting")

	// Path parametrlarini olish
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResp{Error: "User ID talab qilinadi"})
		return
	}

	var req UpdateUserRoleReq
	if err := c.BindJSON(&req); err != nil {
		h.Log.Error("Binding error: " + err.Error())
		c.JSON(http.StatusBadRequest, ErrorResp{Error: "Noto'g'ri so'rov formati"})
		return
	}

	// Rolni yangilash
	if err := h.User.UpdateUserRole(c, userID, db.Role(req.Role)); err != nil {
		h.Log.Error("Update role error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Rolni yangilashda xato"})
		return
	}

	h.Log.Info("Foydalanuvchi roli yangilandi", "user_id", userID, "new_role", req.Role)
	c.JSON(http.StatusOK, SuccessResp{Message: "Rol muvaffaqiyatli yangilandi"})
}

// DeleteUser godoc
// @Summary Delete user
// @Description delete user (admin only)
// @Tags admin
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResp
// @Failure 400 {object} ErrorResp
// @Failure 401 {object} ErrorResp
// @Failure 403 {object} ErrorResp
// @Failure 500 {object} ErrorResp
// @Router /admin/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	h.Log.Info("DeleteUser is starting")

	// Path parametrlarini olish
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResp{Error: "User ID talab qilinadi"})
		return
	}

	// Foydalanuvchini o'chirish
	if err := h.User.DeleteUser(c, userID); err != nil {
		h.Log.Error("Delete user error: " + err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResp{Error: "Foydalanuvchini o'chirishda xato"})
		return
	}

	h.Log.Info("Foydalanuvchi o'chirildi", "user_id", userID)
	c.JSON(http.StatusOK, SuccessResp{Message: "Foydalanuvchi muvaffaqiyatli o'chirildi"})
}
