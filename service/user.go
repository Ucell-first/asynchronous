// service/user_service.go
package service

import (
	"asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	storage storage.IStorage
	logger  *slog.Logger
}

func NewUserService(db storage.IStorage, logger *slog.Logger) *UserService {
	return &UserService{
		storage: db,
		logger:  logger,
	}
}

func isValidName(name string) bool {
	return len(name) >= 2 && len(name) <= 100
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && len(email) >= 5
}

func isValidPassword(password string) bool {
	return len(password) >= 8
}

// Register - Foydalanuvchini ro'yxatdan o'tkazish
func (s *UserService) Register(ctx context.Context, req db.User) (string, error) {
	s.logger.Info("Register metodi ishga tushdi", "email", req.Email)

	// Validatsiyalar
	if !isValidEmail(req.Email) {
		return "", errors.New("noto'g'ri email formati")
	}
	if !isValidName(req.Name) {
		return "", errors.New("ism 2-100 belgidan iborat bo'lishi kerak")
	}
	if !isValidName(req.Surname) {
		return "", errors.New("familiya 2-100 belgidan iborat bo'lishi kerak")
	}
	if !isValidPassword(req.PasswordHash) {
		return "", errors.New("parol kamida 8 ta belgidan iborat bo'lishi kerak")
	}
	if req.Role != db.RoleAdmin && req.Role != db.RoleWorker {
		req.Role = db.RoleWorker // Default role
	}

	// Email unikal ekanligini tekshirish
	existingUser, err := s.storage.User().GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser.ID != "" {
		s.logger.Warn("Bu email allaqachon ro'yxatdan o'tgan", "email", req.Email)
		return "", errors.New("email allaqachon mavjud")
	}

	// Parolni hash qilish
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Parolni hash qilishda xato", "error", err)
		return "", fmt.Errorf("parolni hash qilishda xato: %w", err)
	}
	req.PasswordHash = string(hashedPassword)

	// Bazaga yozish
	userID, err := s.storage.User().CreateUser(ctx, req)
	if err != nil {
		s.logger.Error("Foydalanuvchini saqlashda xato", "error", err)
		return "", fmt.Errorf("foydalanuvchini saqlashda xato: %w", err)
	}

	s.logger.Info("Foydalanuvchi muvaffaqiyatli ro'yxatdan o'tdi",
		"user_id", userID,
		"role", req.Role,
	)
	return userID, nil
}

// Login - Foydalanuvchini autentifikatsiya qilish
func (s *UserService) Login(ctx context.Context, email, password string) (*db.User, error) {
	s.logger.Info("Login metodi ishga tushdi", "email", email)

	// Validatsiya
	if !isValidEmail(email) || !isValidPassword(password) {
		return nil, errors.New("email yoki parol noto'g'ri")
	}

	user, err := s.storage.User().GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Foydalanuvchi topilmadi", "email", email, "error", err)
		return nil, errors.New("email yoki parol noto'g'ri")
	}

	// Parolni solishtirish
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("Noto'g'ri parol kiritildi", "email", email)
		return nil, errors.New("email yoki parol noto'g'ri")
	}

	// Parolni qaytarmaymiz
	user.PasswordHash = ""
	return &user, nil
}

// GetUserProfile - Foydalanuvchi profilini olish
func (s *UserService) GetUserProfile(ctx context.Context, userID string) (*db.User, error) {
	s.logger.Info("GetUserProfile metodi ishga tushdi", "user_id", userID)

	user, err := s.storage.User().GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Foydalanuvchi topilmadi", "user_id", userID, "error", err)
		return nil, fmt.Errorf("foydalanuvchi topilmadi: %w", err)
	}

	// Parolni qaytarmaymiz
	user.PasswordHash = ""
	return &user, nil
}

// UpdateUser - Foydalanuvchi ma'lumotlarini yangilash
func (s *UserService) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	s.logger.Info("UpdateUser metodi ishga tushdi", "user_id", userID)

	existingUser, err := s.storage.User().GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("foydalanuvchi topilmadi: %w", err)
	}

	// Yangilanishlarni qo'llash
	emailChanged := false

	if email, ok := updates["email"].(string); ok {
		if !isValidEmail(email) {
			return errors.New("noto'g'ri email formati")
		}
		// Yangi email unikal ekanligini tekshirish
		if email != existingUser.Email {
			if _, err := s.storage.User().GetUserByEmail(ctx, email); err == nil {
				return errors.New("email allaqachon mavjud")
			}
			emailChanged = true
		}
		existingUser.Email = email
	}

	if name, ok := updates["name"].(string); ok {
		if !isValidName(name) {
			return errors.New("ism noto'g'ri formatda")
		}
		existingUser.Name = name
	}

	if surname, ok := updates["surname"].(string); ok {
		if !isValidName(surname) {
			return errors.New("familiya noto'g'ri formatda")
		}
		existingUser.Surname = surname
	}

	if role, ok := updates["role"].(string); ok {
		if role != string(db.RoleAdmin) && role != string(db.RoleWorker) {
			return errors.New("noto'g'ri rol")
		}
		existingUser.Role = db.Role(role)
	}

	// Faqat o'zgargan bo'lsa yangilash
	if emailChanged || updates["name"] != nil || updates["surname"] != nil || updates["role"] != nil {
		if err := s.storage.User().UpdateUser(ctx, existingUser); err != nil {
			s.logger.Error("Yangilashda xato", "error", err)
			return fmt.Errorf("yangilashda xato: %w", err)
		}
	}

	return nil
}

// DeleteUser - Foydalanuvchini o'chirish (soft delete)
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	s.logger.Info("DeleteUser metodi ishga tushdi", "user_id", userID)

	if err := s.storage.User().DeleteUser(ctx, userID); err != nil {
		s.logger.Error("O'chirishda xato", "error", err)
		return fmt.Errorf("o'chirishda xato: %w", err)
	}
	return nil
}

// ListUsers - Foydalanuvchilar ro'yxati
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]db.User, error) {
	s.logger.Info("ListUsers metodi ishga tushdi")

	users, err := s.storage.User().ListUsers(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Ro'yxatni olishda xato", "error", err)
		return nil, fmt.Errorf("ro'yxatni olishda xato: %w", err)
	}

	// Parollarni tozalash
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}

// ChangePassword - Parolni o'zgartirish
func (s *UserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	s.logger.Info("ChangePassword metodi ishga tushdi", "user_id", userID)

	// Validatsiya
	if !isValidPassword(newPassword) {
		return errors.New("yangi parol kamida 8 ta belgidan iborat bo'lishi kerak")
	}

	user, err := s.storage.User().GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("foydalanuvchi topilmadi: %w", err)
	}

	// Eski parolni tekshirish
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("hozirgi parol noto'g'ri")
	}

	// Yangi parolni hash qilish
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("parolni hash qilishda xato: %w", err)
	}

	// Yangilash
	user.PasswordHash = string(newHash)
	if err := s.storage.User().UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("parolni yangilashda xato: %w", err)
	}

	return nil
}

// UpdateUserRole - Foydalanuvchi rolini yangilash
func (s *UserService) UpdateUserRole(ctx context.Context, userID string, newRole db.Role) error {
	s.logger.Info("Foydalanuvchi rolini yangilash", "user_id", userID, "new_role", newRole)

	// Rolni tekshirish
	if newRole != db.RoleAdmin && newRole != db.RoleWorker {
		return errors.New("noto'g'ri rol")
	}

	user, err := s.storage.User().GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("foydalanuvchi topilmadi: %w", err)
	}

	user.Role = newRole
	return s.storage.User().UpdateUser(ctx, user)
}

// ListUsersByRole - Rol bo'yicha foydalanuvchilarni olish
func (s *UserService) ListUsersByRole(ctx context.Context, role db.Role, limit, offset int) ([]db.User, error) {
	s.logger.Info("ListUsersByRole metodi ishga tushdi", "role", role)

	users, err := s.storage.User().ListUsersByRole(ctx, role, limit, offset)
	if err != nil {
		s.logger.Error("Ro'yxatni olishda xato", "error", err)
		return nil, fmt.Errorf("ro'yxatni olishda xato: %w", err)
	}

	// Parollarni tozalash
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}
