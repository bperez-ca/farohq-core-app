package usecases

import (
	"context"
	"testing"
	"time"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInviteRepository and MockTenantRepository are defined in list_invites_test.go
// Reusing them here

func TestRevokeInvite_Execute(t *testing.T) {
	tests := []struct {
		name          string
		inviteID      uuid.UUID
		tenantID      uuid.UUID
		mockSetup     func(*MockInviteRepository, *MockTenantRepository, uuid.UUID, uuid.UUID)
		expectedError error
	}{
		{
			name:     "successfully revokes invite",
			inviteID: uuid.New(),
			tenantID: uuid.New(),
			mockSetup: func(inviteRepo *MockInviteRepository, tenantRepo *MockTenantRepository, tenantID, inviteID uuid.UUID) {
				tenant := &model.Tenant{}
				tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenant, nil)
				
				invite := model.NewInvite(tenantID, "test@example.com", model.RoleViewer, "token1", uuid.New(), 7*24*time.Hour)
				inviteRepo.On("FindByID", mock.Anything, inviteID).Return(invite, nil)
				inviteRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "returns error when tenant not found",
			inviteID: uuid.New(),
			tenantID: uuid.New(),
			mockSetup: func(inviteRepo *MockInviteRepository, tenantRepo *MockTenantRepository, tenantID, inviteID uuid.UUID) {
				tenantRepo.On("FindByID", mock.Anything, tenantID).Return(nil, domain.ErrTenantNotFound)
			},
			expectedError: domain.ErrTenantNotFound,
		},
		{
			name:     "returns error when invite not found",
			inviteID: uuid.New(),
			tenantID: uuid.New(),
			mockSetup: func(inviteRepo *MockInviteRepository, tenantRepo *MockTenantRepository, tenantID, inviteID uuid.UUID) {
				tenant := &model.Tenant{}
				tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenant, nil)
				inviteRepo.On("FindByID", mock.Anything, inviteID).Return(nil, domain.ErrInviteNotFound)
			},
			expectedError: domain.ErrInviteNotFound,
		},
		{
			name:     "returns error when invite already accepted",
			inviteID: uuid.New(),
			tenantID: uuid.New(),
			mockSetup: func(inviteRepo *MockInviteRepository, tenantRepo *MockTenantRepository, tenantID, inviteID uuid.UUID) {
				tenant := &model.Tenant{}
				tenantRepo.On("FindByID", mock.Anything, tenantID).Return(tenant, nil)
				
				invite := model.NewInvite(tenantID, "test@example.com", model.RoleViewer, "token1", uuid.New(), 7*24*time.Hour)
				invite.Accept()
				inviteRepo.On("FindByID", mock.Anything, inviteID).Return(invite, nil)
			},
			expectedError: domain.ErrInviteAlreadyAccepted,
		},
	}

	// MockTenantRepository is defined in list_invites_test.go
	// Reusing it here

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inviteRepo := new(MockInviteRepository)
			tenantRepo := new(MockTenantRepository)
			
			if tt.mockSetup != nil {
				tt.mockSetup(inviteRepo, tenantRepo, tt.tenantID, tt.inviteID)
			}

			uc := NewRevokeInvite(inviteRepo, tenantRepo)
			req := &RevokeInviteRequest{
				InviteID: tt.inviteID,
				TenantID: tt.tenantID,
			}

			resp, err := uc.Execute(context.Background(), req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Invite.IsRevoked())
			}

			inviteRepo.AssertExpectations(t)
			tenantRepo.AssertExpectations(t)
		})
	}
}
