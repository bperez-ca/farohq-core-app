package usecases

import (
	"context"
	"testing"

	"farohq-core-app/internal/domains/tenants/domain"
	"farohq-core-app/internal/domains/tenants/domain/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInviteRepository is a mock implementation of InviteRepository
type MockInviteRepository struct {
	mock.Mock
}

func (m *MockInviteRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Invite, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Invite), args.Error(1)
}

func (m *MockInviteRepository) FindByToken(ctx context.Context, token string) (*model.Invite, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Invite), args.Error(1)
}

func (m *MockInviteRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*model.Invite, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Invite), args.Error(1)
}

func (m *MockInviteRepository) FindByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*model.Invite, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Invite), args.Error(1)
}

func (m *MockInviteRepository) FindPendingInvitesByEmail(ctx context.Context, email string) ([]*model.Invite, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Invite), args.Error(1)
}

func (m *MockInviteRepository) Save(ctx context.Context, invite *model.Invite) error {
	args := m.Called(ctx, invite)
	return args.Error(0)
}

func (m *MockInviteRepository) Update(ctx context.Context, invite *model.Invite) error {
	args := m.Called(ctx, invite)
	return args.Error(0)
}

func (m *MockInviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTenantRepository is a mock implementation of TenantRepository
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Save(ctx context.Context, tenant *model.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestListInvites_Execute(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      uuid.UUID
		mockSetup     func(*MockInviteRepository, *MockTenantRepository)
		expectedError error
		expectedCount int
	}{
		{
			name:     "successfully lists invites",
			tenantID: uuid.New(),
			mockSetup: func(inviteRepo *MockInviteRepository, tenantRepo *MockTenantRepository) {
				tenant := &model.Tenant{}
				tenantRepo.On("FindByID", mock.Anything, mock.Anything).Return(tenant, nil)

				invites := []*model.Invite{
					model.NewInvite(uuid.New(), "test@example.com", model.RoleViewer, "token1", uuid.New(), 7*24*60*60*1000*1000*1000),
				}
				inviteRepo.On("FindByTenantID", mock.Anything, mock.Anything).Return(invites, nil)
			},
			expectedError: nil,
			expectedCount: 1,
		},
		{
			name:     "returns error when tenant not found",
			tenantID: uuid.New(),
			mockSetup: func(inviteRepo *MockInviteRepository, tenantRepo *MockTenantRepository) {
				tenantRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, domain.ErrTenantNotFound)
			},
			expectedError: domain.ErrTenantNotFound,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inviteRepo := new(MockInviteRepository)
			tenantRepo := new(MockTenantRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(inviteRepo, tenantRepo)
			}

			uc := NewListInvites(inviteRepo, tenantRepo)
			req := &ListInvitesRequest{
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
				assert.Equal(t, tt.expectedCount, len(resp.Invites))
			}

			inviteRepo.AssertExpectations(t)
			tenantRepo.AssertExpectations(t)
		})
	}
}
