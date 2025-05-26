// service/user_service.go
package service

import (
	"asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"

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

// Register - Foydalanuvchini ro'yxatdan o'tkazish
func (s *UserService) Register(ctx context.Context, req db.User) (string, error) {
	s.logger.Info("Register metodi ishga tushdi", "email", req.Email)

	// Parolni hash qilish
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Parolni hash qilishda xato", "error", err)
		return "", fmt.Errorf("parolni hash qilishda xato: %w", err)
	}
	req.PasswordHash = string(hashedPassword)

	// Email unikal ekanligini tekshirish
	existingUser, err := s.storage.User().GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser.ID != "" {
		s.logger.Warn("Bu email allaqachon ro'yxatdan o'tgan", "email", req.Email)
		return "", errors.New("email allaqachon mavjud")
	}

	// Bazaga yozish
	userID, err := s.storage.User().CreateUser(ctx, req)
	if err != nil {
		s.logger.Error("Foydalanuvchini saqlashda xato", "error", err)
		return "", fmt.Errorf("foydalanuvchini saqlashda xato: %w", err)
	}

	s.logger.Info("Foydalanuvchi muvaffaqiyatli ro'yxatdan o'tdi", "user_id", userID)
	return userID, nil
}

// Login - Foydalanuvchini autentifikatsiya qilish
func (s *UserService) Login(ctx context.Context, email, password string) (*db.User, error) {
	s.logger.Info("Login metodi ishga tushdi", "email", email)

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

	// Parolni xavfsizligi uchun o'chirib qo'yamiz
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
	if email, ok := updates["email"].(string); ok {
		existingUser.Email = email
	}
	if name, ok := updates["name"].(string); ok {
		existingUser.Name = name
	}

	if err := s.storage.User().UpdateUser(ctx, existingUser); err != nil {
		s.logger.Error("Yangilashda xato", "error", err)
		return fmt.Errorf("yangilashda xato: %w", err)
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
