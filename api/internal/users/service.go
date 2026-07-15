package users

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kishan-thanki/resumeranker/api/internal/audit"
	"github.com/kishan-thanki/resumeranker/api/internal/config"
	emailpkg "github.com/kishan-thanki/resumeranker/api/internal/email"
	"github.com/kishan-thanki/resumeranker/api/internal/password"
)

type auditService interface {
	LogEvent(ctx context.Context, event *audit.AuditEvent) error
}

type emailService interface {
	SendEmail(ctx context.Context, req *emailpkg.SendEmailRequest) error
}

type UserService struct {
	repo          UserRepository
	agreementRepo AgreementRepository
	auditService  auditService
	emailService  emailService
	cfg           *config.Config
	wg            *sync.WaitGroup
}

func NewUserService(repo UserRepository, agreementRepo AgreementRepository, auditService auditService, emailService emailService, cfg *config.Config, wg *sync.WaitGroup) *UserService {
	return &UserService{
		repo:          repo,
		agreementRepo: agreementRepo,
		auditService:  auditService,
		emailService:  emailService,
		cfg:           cfg,
		wg:            wg,
	}
}

func (s *UserService) Register(ctx context.Context, email, passwordStr string, role Role, agreedToTerms bool) (*User, error) {

	if !agreedToTerms {
		return nil, errors.New("must agree to terms of service and privacy policy")
	}

	hashedPassword, err := password.HashIt(passwordStr)
	if err != nil {
		return nil, errors.Join(errors.New("failed to hash password"), err)
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(s.cfg.VerifyTokenDurationHours) * time.Hour)

	user := &User{
		Email:                 email,
		PasswordHash:          hashedPassword,
		Role:                  role,
		Status:                AccountStatusActive,
		IsVerified:            false,
		VerificationToken:     &token,
		VerificationExpiresAt: &expiresAt,
	}

	createdUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventUserRegistered,
		Description: "user registered successfully",
		UserID:      &createdUser.ID,
	})

	latestAgreements, err := s.agreementRepo.GetLatestAgreements(ctx)
	if err == nil {
		for _, agreement := range latestAgreements {
			_, _ = s.agreementRepo.CreateUserAgreement(ctx, &UserAgreement{
				UserID:      createdUser.ID,
				AgreementID: agreement.ID,
			})
		}
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		link := fmt.Sprintf("%s/verify?token=%s", s.cfg.Domain, token)
		htmlBody := emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
			Title:        "Verify your email",
			Message:      "<p>Welcome to ResumeRanker! To get started and gain access to your dashboard, please verify your email address.</p>",
			BtnText:      "Verify Email",
			BtnLink:      link,
			SupportEmail: s.cfg.EmailContact,
			Domain:       s.cfg.Domain,
			FooterNote:   "If you did not request this email, you can safely ignore it.",
		})
		_ = s.emailService.SendEmail(context.Background(), &emailpkg.SendEmailRequest{
			To:      []string{createdUser.Email},
			Subject: "Welcome to ResumeRanker - Verify your email",
			Text:    "Your verification token is: " + token,
			HTML:    htmlBody,
		})
	}()

	return createdUser, nil
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.repo.GetUserByVerificationToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	if user.VerificationExpiresAt == nil || time.Now().After(*user.VerificationExpiresAt) {
		return errors.New("verification token has expired")
	}

	_, err = s.repo.VerifyUserEmail(ctx, user.ID)
	return err
}

func (s *UserService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(s.cfg.ResetTokenDurationHours) * time.Hour)

	user.PasswordResetToken = &token
	user.PasswordResetExpiresAt = &expiresAt

	_, err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.Domain, token)
		htmlBody := emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
			Title:        "Password Reset Request",
			Message:      fmt.Sprintf("<p>You recently requested to reset your password for your ResumeRanker account.</p><p>This link is valid for <strong>%d hours</strong>.</p>", s.cfg.ResetTokenDurationHours),
			BtnText:      "Reset Password",
			BtnLink:      link,
			SupportEmail: s.cfg.EmailContact,
			Domain:       s.cfg.Domain,
			FooterNote:   "If you did not request this password reset, please ignore this email or contact support if you have concerns.",
		})
		_ = s.emailService.SendEmail(context.Background(), &emailpkg.SendEmailRequest{
			To:      []string{user.Email},
			Subject: "ResumeRanker - Password Reset",
			Text:    "Your password reset token is: " + token,
			HTML:    htmlBody,
		})
	}()

	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.repo.GetUserByPasswordResetToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if user.PasswordResetExpiresAt == nil || time.Now().After(*user.PasswordResetExpiresAt) {
		return errors.New("password reset token has expired")
	}

	hashedPassword, err := password.HashIt(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = hashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpiresAt = nil

	_, err = s.repo.UpdateUser(ctx, user)
	if err == nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventUserPasswordChanged,
			Description: "user reset password successfully",
			UserID:      &user.ID,
		})
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			htmlBody := emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
				Title:        "Password Reset Successful",
				Message:      "<p>This is a confirmation that the password for your ResumeRanker account has just been reset.</p><p>If you did not make this change, please contact support immediately.</p>",
				SupportEmail: s.cfg.EmailContact,
				Domain:       s.cfg.Domain,
			})
			_ = s.emailService.SendEmail(context.Background(), &emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Your Password Was Reset",
				Text:    "This is a confirmation that the password for your ResumeRanker account has just been reset. If you did not make this change, please contact support immediately.",
				HTML:    htmlBody,
			})
		}()
	}
	return err
}

func (s *UserService) Authenticate(ctx context.Context, email, passwordStr string) (*User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	match, err := password.VerifyHash(passwordStr, user.PasswordHash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	if !user.IsVerified {
		return nil, errors.New("email is not verified")
	}

	if user.Status != AccountStatusActive {
		if user.Status == AccountStatusSuspended {
			return nil, ErrAccountSuspended
		}
		return nil, errors.New("account is not active")
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventUserLoggedIn,
		Description: "user logged in successfully",
		UserID:      &user.ID,
	})
	return user, nil
}

func (s *UserService) GetMe(ctx context.Context, userID uint64) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int32) ([]*User, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	users, err := s.repo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := password.VerifyHash(oldPassword, user.PasswordHash)
	if err != nil || !match {
		return ErrIncorrectPassword
	}

	hashedPassword, err := password.HashIt(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = hashedPassword
	_, err = s.repo.UpdateUser(ctx, user)
	if err == nil {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type:        audit.AuditEventUserPasswordChanged,
			Description: "user changed password successfully",
			UserID:      &userID,
		})
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			htmlBody := emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
				Title:        "Password Changed",
				Message:      "<p>This is a confirmation that the password for your ResumeRanker account has just been changed.</p><p>If you did not make this change, please contact support immediately.</p>",
				SupportEmail: s.cfg.EmailContact,
				Domain:       s.cfg.Domain,
			})
			_ = s.emailService.SendEmail(context.Background(), &emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Your Password Was Changed",
				Text:    "This is a confirmation that the password for your ResumeRanker account has just been changed. If you did not make this change, please contact support immediately.",
				HTML:    htmlBody,
			})
		}()
	}
	return err
}

func (s *UserService) ToggleStatus(ctx context.Context, userID uint64, status AccountStatus) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Status == status {
		return nil
	}

	user.Status = status
	_, err = s.repo.UpdateUser(ctx, user)
	if err == nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			htmlBody := emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
				Title:        "Account Status Update",
				Message:      fmt.Sprintf("<p>Your ResumeRanker account status has been updated by an administrator.</p><p>New Status: <strong>%s</strong></p>", status),
				SupportEmail: s.cfg.EmailContact,
				Domain:       s.cfg.Domain,
			})
			_ = s.emailService.SendEmail(context.Background(), &emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Account Status Update",
				Text:    fmt.Sprintf("Your account status has been updated to: %s.", status),
				HTML:    htmlBody,
			})
		}()
	}
	return err
}

func (s *UserService) DeleteAccount(ctx context.Context, userID uint64) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	err = s.repo.DeleteUser(ctx, userID)
	if err == nil {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			htmlBody := emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
				Title:        "Account Deleted",
				Message:      "<p>Your account and all associated data have been permanently deleted from ResumeRanker.</p><p>We're sorry to see you go.</p>",
				SupportEmail: s.cfg.EmailContact,
				Domain:       s.cfg.Domain,
			})
			_ = s.emailService.SendEmail(context.Background(), &emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Account Deletion Confirmation",
				Text:    "Your account and all associated data have been permanently deleted from ResumeRanker.",
				HTML:    htmlBody,
			})
		}()
	}
	return err
}
