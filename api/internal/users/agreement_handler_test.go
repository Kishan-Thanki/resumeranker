package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kishan-thanki/resumeranker/api/internal/users"
)

func TestAgreementService_PublishAgreement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		agType  users.AgreementType
		version string
		content string
		wantErr error
	}{
		{
			name:    "success",
			agType:  users.AgreementTypeTermsOfService,
			version: "1.0",
			content: "terms",
		},
		{
			name:    "invalid type",
			agType:  users.AgreementType("unknown"),
			version: "1.0",
			content: "terms",
			wantErr: users.ErrInvalidAgreementType,
		},
		{
			name:    "missing version",
			agType:  users.AgreementTypeTermsOfService,
			version: "",
			content: "terms",
			wantErr: users.ErrInvalidAgreementVersion,
		},
		{
			name:    "missing content",
			agType:  users.AgreementTypeTermsOfService,
			version: "1.0",
			content: "",
			wantErr: users.ErrInvalidAgreementContent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var created *users.Agreement

			repo := &MockAgreementRepository{
				CreateAgreementFunc: func(
					_ context.Context,
					agreement *users.Agreement,
				) (*users.Agreement, error) {
					created = agreement
					agreement.ID = 1
					return agreement, nil
				},
			}

			svc := users.NewAgreementService(
				repo,
				&MockEmailService{},
				nil,
			)

			got, err := svc.PublishAgreement(
				context.Background(),
				tt.agType,
				tt.version,
				tt.content,
				false,
				nil,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == nil || created == nil {
				t.Fatal("expected created agreement")
			}

			if got.ID != 1 {
				t.Fatalf("expected ID 1, got %d", got.ID)
			}

			if got.Type != tt.agType {
				t.Fatalf("expected type %q, got %q", tt.agType, got.Type)
			}
		})
	}
}

func TestAgreementService_PublishAgreementRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database failure")

	repo := &MockAgreementRepository{
		CreateAgreementFunc: func(
			_ context.Context,
			_ *users.Agreement,
		) (*users.Agreement, error) {
			return nil, expectedErr
		},
	}

	svc := users.NewAgreementService(
		repo,
		&MockEmailService{},
		nil,
	)

	_, err := svc.PublishAgreement(
		context.Background(),
		users.AgreementTypeTermsOfService,
		"1.0",
		"terms",
		false,
		nil,
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestAgreementService_GetLatestAgreements(t *testing.T) {
	t.Parallel()

	expected := []*users.Agreement{
		{
			ID:      1,
			Type:    users.AgreementTypeTermsOfService,
			Version: "2.0",
		},
		{
			ID:      2,
			Type:    users.AgreementTypePrivacyPolicy,
			Version: "2.0",
		},
	}

	svc := users.NewAgreementService(
		&MockAgreementRepository{
			GetLatestAgreementsFunc: func(
				_ context.Context,
			) ([]*users.Agreement, error) {
				return expected, nil
			},
		},
		&MockEmailService{},
		nil,
	)

	got, err := svc.GetLatestAgreements(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 agreements, got %d", len(got))
	}
}

func TestAgreementService_GetPendingAgreements(t *testing.T) {
	t.Parallel()

	svc := users.NewAgreementService(
		&MockAgreementRepository{
			GetPendingAgreementsForUserFunc: func(
				_ context.Context,
				userID uint64,
			) ([]*users.Agreement, error) {
				if userID != 42 {
					t.Fatalf("expected user ID 42, got %d", userID)
				}

				return []*users.Agreement{
					{
						ID:      10,
						Type:    users.AgreementTypeTermsOfService,
						Version: "2.0",
					},
				}, nil
			},
		},
		&MockEmailService{},
		nil,
	)

	got, err := svc.GetPendingAgreements(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 || got[0].ID != 10 {
		t.Fatalf("unexpected agreements: %#v", got)
	}
}

func TestAgreementService_AcceptTerms(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var accepted []uint64

		repo := &MockAgreementRepository{
			GetAgreementByTypeAndVersionFunc: func(
				_ context.Context,
				agType users.AgreementType,
				version string,
			) (*users.Agreement, error) {
				switch agType {
				case users.AgreementTypeTermsOfService:
					if version != "1.0" {
						t.Fatalf("unexpected terms version %q", version)
					}
					return &users.Agreement{ID: 10}, nil
				case users.AgreementTypePrivacyPolicy:
					if version != "1.0" {
						t.Fatalf("unexpected privacy version %q", version)
					}
					return &users.Agreement{ID: 11}, nil
				default:
					t.Fatalf("unexpected type %q", agType)
					return nil, nil
				}
			},
			CreateUserAgreementFunc: func(
				_ context.Context,
				userAgreement *users.UserAgreement,
			) (*users.UserAgreement, error) {
				accepted = append(accepted, userAgreement.AgreementID)
				return userAgreement, nil
			},
		}

		svc := users.NewAgreementService(
			repo,
			&MockEmailService{},
			nil,
		)

		if err := svc.AcceptTerms(
			context.Background(),
			42,
			"1.0",
			"1.0",
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(accepted) != 2 {
			t.Fatalf("expected 2 accepted agreements, got %d", len(accepted))
		}

		if accepted[0] != 10 || accepted[1] != 11 {
			t.Fatalf("unexpected accepted IDs: %#v", accepted)
		}
	})

	t.Run("invalid terms version", func(t *testing.T) {
		t.Parallel()

		repo := &MockAgreementRepository{
			GetAgreementByTypeAndVersionFunc: func(
				_ context.Context,
				agType users.AgreementType,
				_ string,
			) (*users.Agreement, error) {
				if agType == users.AgreementTypeTermsOfService {
					return nil, errors.New("not found")
				}
				return &users.Agreement{}, nil
			},
		}

		svc := users.NewAgreementService(
			repo,
			&MockEmailService{},
			nil,
		)

		err := svc.AcceptTerms(
			context.Background(),
			42,
			"bad",
			"1.0",
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("privacy acceptance failure", func(t *testing.T) {
		t.Parallel()

		repo := &MockAgreementRepository{
			GetAgreementByTypeAndVersionFunc: func(
				_ context.Context,
				_ users.AgreementType,
				_ string,
			) (*users.Agreement, error) {
				return &users.Agreement{ID: 10}, nil
			},
			CreateUserAgreementFunc: func(
				_ context.Context,
				userAgreement *users.UserAgreement,
			) (*users.UserAgreement, error) {
				if userAgreement.AgreementID == 11 {
					return nil, errors.New("privacy insert failed")
				}
				return userAgreement, nil
			},
		}

		svc := users.NewAgreementService(
			repo,
			&MockEmailService{},
			nil,
		)

		// Both lookups return ID 10 in this mock, so the second insert
		// must be distinguished by call order in a realistic mock.
		callCount := 0
		repo.GetAgreementByTypeAndVersionFunc = func(
			_ context.Context,
			_ users.AgreementType,
			_ string,
		) (*users.Agreement, error) {
			callCount++
			if callCount == 1 {
				return &users.Agreement{ID: 10}, nil
			}
			return &users.Agreement{ID: 11}, nil
		}

		err := svc.AcceptTerms(
			context.Background(),
			42,
			"1.0",
			"1.0",
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAgreementService_HasAcceptedTerms(t *testing.T) {
	t.Parallel()

	t.Run("both accepted", func(t *testing.T) {
		t.Parallel()

		repo := &MockAgreementRepository{
			GetAgreementByTypeAndVersionFunc: func(
				_ context.Context,
				agType users.AgreementType,
				_ string,
			) (*users.Agreement, error) {
				if agType == users.AgreementTypeTermsOfService {
					return &users.Agreement{ID: 10}, nil
				}
				return &users.Agreement{ID: 11}, nil
			},
			HasUserAcceptedAgreementFunc: func(
				_ context.Context,
				_ uint64,
				agreementID uint64,
			) (bool, error) {
				if agreementID != 10 && agreementID != 11 {
					t.Fatalf("unexpected agreement ID %d", agreementID)
				}
				return true, nil
			},
		}

		svc := users.NewAgreementService(
			repo,
			&MockEmailService{},
			nil,
		)

		got, err := svc.HasAcceptedTerms(
			context.Background(),
			42,
			"1.0",
			"1.0",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Fatal("expected accepted=true")
		}
	})

	t.Run("missing one agreement", func(t *testing.T) {
		t.Parallel()

		repo := &MockAgreementRepository{
			GetAgreementByTypeAndVersionFunc: func(
				_ context.Context,
				agType users.AgreementType,
				_ string,
			) (*users.Agreement, error) {
				if agType == users.AgreementTypeTermsOfService {
					return &users.Agreement{ID: 10}, nil
				}
				return &users.Agreement{ID: 11}, nil
			},
			HasUserAcceptedAgreementFunc: func(
				_ context.Context,
				_ uint64,
				agreementID uint64,
			) (bool, error) {
				return agreementID == 10, nil
			},
		}

		svc := users.NewAgreementService(
			repo,
			&MockEmailService{},
			nil,
		)

		got, err := svc.HasAcceptedTerms(
			context.Background(),
			42,
			"1.0",
			"1.0",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Fatal("expected accepted=false")
		}
	})
}

func TestAgreementService_AcceptAgreements(t *testing.T) {
	t.Parallel()

	t.Run("accepts unique IDs", func(t *testing.T) {
		t.Parallel()

		var accepted []uint64

		repo := &MockAgreementRepository{
			CreateUserAgreementFunc: func(
				_ context.Context,
				userAgreement *users.UserAgreement,
			) (*users.UserAgreement, error) {
				accepted = append(accepted, userAgreement.AgreementID)
				return userAgreement, nil
			},
		}

		svc := users.NewAgreementService(
			repo,
			&MockEmailService{},
			nil,
		)

		if err := svc.AcceptAgreements(
			context.Background(),
			42,
			[]uint64{10, 10, 11},
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(accepted) != 2 {
			t.Fatalf("expected 2 unique agreements, got %d", len(accepted))
		}

		if accepted[0] != 10 || accepted[1] != 11 {
			t.Fatalf("unexpected accepted IDs: %#v", accepted)
		}
	})

	t.Run("rejects empty list", func(t *testing.T) {
		t.Parallel()

		svc := users.NewAgreementService(
			&MockAgreementRepository{},
			&MockEmailService{},
			nil,
		)

		if !errors.Is(
			svc.AcceptAgreements(context.Background(), 42, nil),
			users.ErrNoAgreementsProvided,
		) {
			t.Fatal("expected ErrNoAgreementsProvided")
		}
	})

	t.Run("rejects zero ID", func(t *testing.T) {
		t.Parallel()

		svc := users.NewAgreementService(
			&MockAgreementRepository{},
			&MockEmailService{},
			nil,
		)

		if !errors.Is(
			svc.AcceptAgreements(context.Background(), 42, []uint64{0}),
			users.ErrInvalidAgreementID,
		) {
			t.Fatal("expected ErrInvalidAgreementID")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("insert failed")

		svc := users.NewAgreementService(
			&MockAgreementRepository{
				CreateUserAgreementFunc: func(
					_ context.Context,
					_ *users.UserAgreement,
				) (*users.UserAgreement, error) {
					return nil, expectedErr
				},
			},
			&MockEmailService{},
			nil,
		)

		err := svc.AcceptAgreements(
			context.Background(),
			42,
			[]uint64{10},
		)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected wrapped repository error, got %v", err)
		}
	})
}

func TestAgreementService_PublishAgreementNotifications(t *testing.T) {
	t.Parallel()

	t.Skip("notification execution is intentionally asynchronous; cover with an integration/lifecycle test once service shutdown coordination is added")
}
