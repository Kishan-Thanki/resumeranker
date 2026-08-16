package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kishan-thanki/resumeranker/api/internal/apikey"
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

type apiKeyService interface {
	GenerateKey(ctx context.Context, userID uint64, name string, quota uint64) (string, *apikey.APIKey, error)
}

type UserService struct {
	repo          UserRepository
	agreementRepo AgreementRepository
	auditService  auditService
	emailService  emailService
	apiKeyService apiKeyService
	cfg           *config.Config
	wg            *sync.WaitGroup
}

func NewUserService(
	repo UserRepository,
	agreementRepo AgreementRepository,
	auditService auditService,
	emailService emailService,
	apiKeyService apiKeyService,
	cfg *config.Config,
	wg *sync.WaitGroup,
) *UserService {
	if wg == nil {
		wg = &sync.WaitGroup{}
	}

	return &UserService{
		repo:          repo,
		agreementRepo: agreementRepo,
		auditService:  auditService,
		emailService:  emailService,
		apiKeyService: apiKeyService,
		cfg:           cfg,
		wg:            wg,
	}
}

func (s *UserService) Register(
	ctx context.Context,
	email string,
	passwordStr string,
	role Role,
	agreedToTerms bool,
) (*User, error) {
	if !agreedToTerms {
		return nil, ErrMustAgreeToTerms
	}

	hashedPassword, err := password.HashIt(passwordStr)
	if err != nil {
		return nil, errors.Join(errors.New("failed to hash password"), err)
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(
		time.Duration(s.cfg.VerifyTokenDurationHours) * time.Hour,
	)

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

	latestAgreements, err := s.agreementRepo.GetLatestAgreements(ctx)
	if err != nil {
		_ = s.repo.DeleteUser(ctx, createdUser.ID)
		return nil, fmt.Errorf("failed to load latest agreements: %w", err)
	}

	for _, agreement := range latestAgreements {
		if agreement == nil {
			_ = s.repo.DeleteUser(ctx, createdUser.ID)
			return nil, errors.New("failed to accept registration agreements: nil agreement")
		}

		_, err := s.agreementRepo.CreateUserAgreement(ctx, &UserAgreement{
			UserID:      createdUser.ID,
			AgreementID: agreement.ID,
			AcceptedAt:  time.Now(),
		})
		if err != nil {
			_ = s.repo.DeleteUser(ctx, createdUser.ID)
			return nil, fmt.Errorf(
				"failed to accept registration agreement %d: %w",
				agreement.ID,
				err,
			)
		}
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventUserRegistered,
		Description: "user registered successfully",
		UserID:      &createdUser.ID,
	})

	for _, agreement := range latestAgreements {
		_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
			Type: audit.AuditEventAgreementAccepted,
			Description: fmt.Sprintf(
				"user accepted agreement %d on registration",
				agreement.ID,
			),
			UserID: &createdUser.ID,
		})
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if s.apiKeyService != nil {
			if _, _, err := s.apiKeyService.GenerateKey(
				context.Background(),
				createdUser.ID,
				"Default Dashboard Key",
				100,
			); err != nil {
				slog.Error(
					"failed to generate default dashboard API key",
					"user_id", createdUser.ID,
					"error", err,
				)
			}
		}

		if s.emailService == nil {
			return
		}

		err := s.emailService.SendEmail(
			context.Background(),
			&emailpkg.SendEmailRequest{
				To:      []string{createdUser.Email},
				Subject: "Welcome to ResumeRanker - Verify your email",
				Text:    "Your verification token is: " + token,
				HTML: emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
					Title:   "Verify your email",
					Message: "<p>Welcome to ResumeRanker! To get started and gain access to your dashboard, please verify your email address.</p>",
					BtnText: "Verify Email",
					BtnLink: fmt.Sprintf(
						"%s/verify?token=%s",
						s.cfg.Domain,
						token,
					),
					SupportEmail: s.cfg.EmailContact,
					Domain:       s.cfg.Domain,
					FooterNote:   "If you did not request this email, you can safely ignore it.",
				}),
			},
		)
		if err != nil {
			slog.Error(
				"failed to send registration email",
				"user_id", createdUser.ID,
				"error", err,
			)
		}
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
	expiresAt := time.Now().Add(
		time.Duration(s.cfg.ResetTokenDurationHours) * time.Hour,
	)

	user.PasswordResetToken = &token
	user.PasswordResetExpiresAt = &expiresAt

	if _, err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if s.emailService == nil {
			return
		}

		link := fmt.Sprintf(
			"%s/reset-password?token=%s",
			s.cfg.Domain,
			token,
		)

		err := s.emailService.SendEmail(
			context.Background(),
			&emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "ResumeRanker - Password Reset",
				Text:    "Your password reset token is: " + token,
				HTML: emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
					Title: "Password Reset Request",
					Message: fmt.Sprintf(
						"<p>You recently requested to reset your password for your ResumeRanker account.</p><p>This link is valid for <strong>%d hours</strong>.</p>",
						s.cfg.ResetTokenDurationHours,
					),
					BtnText:      "Reset Password",
					BtnLink:      link,
					SupportEmail: s.cfg.EmailContact,
					Domain:       s.cfg.Domain,
					FooterNote:   "If you did not request this password reset, please ignore this email or contact support if you have concerns.",
				}),
			},
		)
		if err != nil {
			slog.Error(
				"failed to send password reset email",
				"user_id", user.ID,
				"error", err,
			)
		}
	}()

	return nil
}

func (s *UserService) ResetPassword(
	ctx context.Context,
	token string,
	newPassword string,
) error {
	user, err := s.repo.GetUserByPasswordResetToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if user.PasswordResetExpiresAt == nil ||
		time.Now().After(*user.PasswordResetExpiresAt) {
		return errors.New("password reset token has expired")
	}

	hashedPassword, err := password.HashIt(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = hashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpiresAt = nil

	if _, err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventUserPasswordChanged,
		Description: "user reset password successfully",
		UserID:      &user.ID,
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if s.emailService == nil {
			return
		}

		err := s.emailService.SendEmail(
			context.Background(),
			&emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Your Password Was Reset",
				Text:    "This is a confirmation that the password for your ResumeRanker account has just been reset. If you did not make this change, please contact support immediately.",
				HTML: emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
					Title:        "Password Reset Successful",
					Message:      "<p>This is a confirmation that the password for your ResumeRanker account has just been reset.</p><p>If you did not make this change, please contact support immediately.</p>",
					SupportEmail: s.cfg.EmailContact,
					Domain:       s.cfg.Domain,
				}),
			},
		)
		if err != nil {
			slog.Error(
				"failed to send password reset confirmation email",
				"user_id", user.ID,
				"error", err,
			)
		}
	}()

	return nil
}

func (s *UserService) Authenticate(
	ctx context.Context,
	email string,
	passwordStr string,
) (*User, error) {
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

func (s *UserService) ListUsers(
	ctx context.Context,
	limit, offset int32,
) ([]*User, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	userList, err := s.repo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	return userList, count, nil
}

func (s *UserService) ChangePassword(
	ctx context.Context,
	userID uint64,
	oldPassword,
	newPassword string,
) error {
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

	if _, err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	_ = s.auditService.LogEvent(ctx, &audit.AuditEvent{
		Type:        audit.AuditEventUserPasswordChanged,
		Description: "user changed password successfully",
		UserID:      &userID,
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if s.emailService == nil {
			return
		}

		err := s.emailService.SendEmail(
			context.Background(),
			&emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Your Password Was Changed",
				Text:    "This is a confirmation that the password for your ResumeRanker account has just been changed. If you did not make this change, please contact support immediately.",
				HTML: emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
					Title:        "Password Changed",
					Message:      "<p>This is a confirmation that the password for your ResumeRanker account has just been changed.</p><p>If you did not make this change, please contact support immediately.</p>",
					SupportEmail: s.cfg.EmailContact,
					Domain:       s.cfg.Domain,
				}),
			},
		)
		if err != nil {
			slog.Error(
				"failed to send password changed email",
				"user_id", user.ID,
				"error", err,
			)
		}
	}()

	return nil
}

func (s *UserService) ToggleStatus(
	ctx context.Context,
	userID uint64,
	status AccountStatus,
) error {
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

	if _, err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if s.emailService == nil {
			return
		}

		err := s.emailService.SendEmail(
			context.Background(),
			&emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Account Status Update",
				Text: fmt.Sprintf(
					"Your account status has been updated to: %s.",
					status,
				),
				HTML: emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
					Title: "Account Status Update",
					Message: fmt.Sprintf(
						"<p>Your ResumeRanker account status has been updated by an administrator.</p><p>New Status: <strong>%s</strong></p>",
						status,
					),
					SupportEmail: s.cfg.EmailContact,
					Domain:       s.cfg.Domain,
				}),
			},
		)
		if err != nil {
			slog.Error(
				"failed to send account status email",
				"user_id", user.ID,
				"error", err,
			)
		}
	}()

	return nil
}

func (s *UserService) DeleteAccount(ctx context.Context, userID uint64) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		if s.emailService == nil {
			return
		}

		err := s.emailService.SendEmail(
			context.Background(),
			&emailpkg.SendEmailRequest{
				To:      []string{user.Email},
				Subject: "Account Deletion Confirmation",
				Text:    "Your account and all associated data have been permanently deleted from ResumeRanker.",
				HTML: emailpkg.BuildHTMLTemplate(emailpkg.HTMLTemplateParams{
					Title:        "Account Deleted",
					Message:      "<p>Your account and all associated data have been permanently deleted from ResumeRanker.</p><p>We're sorry to see you go.</p>",
					SupportEmail: s.cfg.EmailContact,
					Domain:       s.cfg.Domain,
				}),
			},
		)
		if err != nil {
			slog.Error(
				"failed to send account deletion email",
				"user_id", user.ID,
				"error", err,
			)
		}
	}()

	return nil
}
