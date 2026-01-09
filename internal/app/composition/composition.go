package composition

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	auth_http "farohq-core-app/internal/domains/auth/infra/http"
	brand_usecases "farohq-core-app/internal/domains/brand/app/usecases"
	brand_db "farohq-core-app/internal/domains/brand/infra/db"
	brand_http "farohq-core-app/internal/domains/brand/infra/http"
	files_usecases "farohq-core-app/internal/domains/files/app/usecases"
	files_services "farohq-core-app/internal/domains/files/domain/services"
	files_http "farohq-core-app/internal/domains/files/infra/http"
	"farohq-core-app/internal/domains/files/infra/s3"
	tenants_usecases "farohq-core-app/internal/domains/tenants/app/usecases"
	tenants_services "farohq-core-app/internal/domains/tenants/domain/services"
	tenants_db "farohq-core-app/internal/domains/tenants/infra/db"
	tenants_http "farohq-core-app/internal/domains/tenants/infra/http"
	users_usecases "farohq-core-app/internal/domains/users/app/usecases"
	users_db "farohq-core-app/internal/domains/users/infra/db"
	users_http "farohq-core-app/internal/domains/users/infra/http"
	"farohq-core-app/internal/platform/config"
)

// Composition wires all domains together
type Composition struct {
	TenantHandlers *tenants_http.Handlers
	BrandHandlers  *brand_http.Handlers
	FilesHandlers  *files_http.Handlers
	AuthHandlers   *auth_http.Handlers
	UserHandlers   *users_http.Handlers
}

// RegisterPublicRoutes registers public routes (no auth required)
func (c *Composition) RegisterPublicRoutes(r chi.Router) {
	// Public brand routes
	c.BrandHandlers.RegisterPublicRoutes(r)
}

// RegisterProtectedRoutes registers protected routes (auth required)
func (c *Composition) RegisterProtectedRoutes(r chi.Router) {
	// Register all domain routes
	c.TenantHandlers.RegisterRoutes(r)
	c.BrandHandlers.RegisterRoutes(r)
	c.FilesHandlers.RegisterRoutes(r)
	c.AuthHandlers.RegisterRoutes(r)
	c.UserHandlers.RegisterRoutes(r)
}

// RegisterProtectedRoutesWithTenant registers protected routes that require tenant context
// This excludes routes that don't need tenant context (e.g., POST /tenants, GET /auth/me)
func (c *Composition) RegisterProtectedRoutesWithTenant(r chi.Router) {
	// Register tenant routes (excluding POST /tenants which doesn't need tenant context)
	// Note: We register these directly since POST /tenants is already registered in main.go
	r.Get("/tenants/{id}", c.TenantHandlers.GetTenantHandler)
	r.Put("/tenants/{id}", c.TenantHandlers.UpdateTenantHandler)
	r.Post("/tenants/{id}/invites", c.TenantHandlers.InviteMemberHandler)
	r.Get("/tenants/{id}/members", c.TenantHandlers.ListMembersHandler)
	r.Delete("/tenants/{id}/members/{user_id}", c.TenantHandlers.RemoveMemberHandler)
	r.Get("/tenants/{id}/roles", c.TenantHandlers.ListRolesHandler)
	r.Get("/tenants/{id}/seat-usage", c.TenantHandlers.GetSeatUsageHandler)
	r.Post("/tenants/{id}/clients", c.TenantHandlers.CreateClientHandler)
	r.Get("/tenants/{id}/clients", c.TenantHandlers.ListClientsHandler)

	// Register client routes (all require tenant context)
	r.Route("/clients", func(r chi.Router) {
		r.Get("/{id}", c.TenantHandlers.GetClientHandler)
		r.Put("/{id}", c.TenantHandlers.UpdateClientHandler)
		r.Post("/{id}/members", c.TenantHandlers.AddClientMemberHandler)
		r.Get("/{id}/members", c.TenantHandlers.ListClientMembersHandler)
		r.Delete("/{id}/members/{memberId}", c.TenantHandlers.RemoveClientMemberHandler)
		r.Post("/{id}/locations", c.TenantHandlers.CreateLocationHandler)
		r.Get("/{id}/locations", c.TenantHandlers.ListLocationsHandler)
	})

	// Register location routes (all require tenant context)
	r.Route("/locations", func(r chi.Router) {
		r.Put("/{id}", c.TenantHandlers.UpdateLocationHandler)
	})

	// Register brand routes (all require tenant context)
	c.BrandHandlers.RegisterRoutes(r)

	// Register files routes (all require tenant context)
	c.FilesHandlers.RegisterRoutes(r)
}

// NewComposition creates a new composition with all dependencies wired
func NewComposition(
	db *pgxpool.Pool,
	cfg *config.Config,
	logger zerolog.Logger,
) *Composition {
		// Initialize repositories
	tenantRepo := tenants_db.NewTenantRepository(db)
	tenantMemberRepo := tenants_db.NewTenantMemberRepository(db)
	inviteRepo := tenants_db.NewInviteRepository(db)
	clientRepo := tenants_db.NewClientRepository(db)
	locationRepo := tenants_db.NewLocationRepository(db)
	clientMemberRepo := tenants_db.NewClientMemberRepository(db)
	brandRepo := brand_db.NewBrandRepository(db)
	userRepo := users_db.NewUserRepository(db)

	// Initialize services
	seatValidator := tenants_services.NewSeatValidator()
	assetValidator := files_services.NewAssetValidator()
	keyGenerator := files_services.NewKeyGenerator()

	// Initialize storage
	storage, err := s3.NewStorage(cfg.AWSRegion, cfg.S3BucketName)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize S3 storage")
	}

	// Initialize tenant use cases
	createTenant := tenants_usecases.NewCreateTenant(tenantRepo)
	getTenant := tenants_usecases.NewGetTenant(tenantRepo)
	updateTenant := tenants_usecases.NewUpdateTenant(tenantRepo)
	inviteMember := tenants_usecases.NewInviteMember(inviteRepo, tenantMemberRepo, tenantRepo, seatValidator, 7*24*time.Hour)
	acceptInvite := tenants_usecases.NewAcceptInvite(inviteRepo, tenantMemberRepo)
	listMembers := tenants_usecases.NewListMembers(tenantMemberRepo, tenantRepo)
	removeMember := tenants_usecases.NewRemoveMember(tenantMemberRepo, tenantRepo)
	listRoles := tenants_usecases.NewListRoles(tenantRepo)
	createClient := tenants_usecases.NewCreateClient(clientRepo, tenantRepo, seatValidator)
	listClients := tenants_usecases.NewListClients(clientRepo, tenantRepo)
	getClient := tenants_usecases.NewGetClient(clientRepo)
	updateClient := tenants_usecases.NewUpdateClient(clientRepo)
	addClientMember := tenants_usecases.NewAddClientMember(clientMemberRepo, locationRepo, seatValidator)
	listClientMembers := tenants_usecases.NewListClientMembers(clientMemberRepo)
	removeClientMember := tenants_usecases.NewRemoveClientMember(clientMemberRepo)
	createLocation := tenants_usecases.NewCreateLocation(locationRepo, clientRepo)
	listLocations := tenants_usecases.NewListLocations(locationRepo)
	updateLocation := tenants_usecases.NewUpdateLocation(locationRepo)
	getSeatUsage := tenants_usecases.NewGetSeatUsage(tenantRepo, clientRepo, clientMemberRepo, locationRepo)

	// Initialize brand use cases
	getByDomain := brand_usecases.NewGetByDomain(brandRepo)
	getByHost := brand_usecases.NewGetByHost(brandRepo, tenantRepo)
	listBrands := brand_usecases.NewListBrands(brandRepo)
	createBrand := brand_usecases.NewCreateBrand(brandRepo)
	getBrand := brand_usecases.NewGetBrand(brandRepo)
	updateBrand := brand_usecases.NewUpdateBrand(brandRepo)
	deleteBrand := brand_usecases.NewDeleteBrand(brandRepo)

	// Initialize files use cases
	signUpload := files_usecases.NewSignUpload(storage, assetValidator, keyGenerator, cfg.S3BucketName, 10*time.Minute)
	deleteFile := files_usecases.NewDeleteFile(storage, keyGenerator, cfg.S3BucketName)

	// Initialize user use cases
	syncUser := users_usecases.NewSyncUser(userRepo)

	// Initialize handlers
	tenantHandlers := tenants_http.NewHandlers(
		logger,
		createTenant,
		getTenant,
		updateTenant,
		inviteMember,
		acceptInvite,
		listMembers,
		removeMember,
		listRoles,
		createClient,
		listClients,
		getClient,
		updateClient,
		addClientMember,
		listClientMembers,
		removeClientMember,
		createLocation,
		listLocations,
		updateLocation,
		getSeatUsage,
	)

	brandHandlers := brand_http.NewHandlers(
		logger,
		getByDomain,
		getByHost,
		listBrands,
		createBrand,
		getBrand,
		updateBrand,
		deleteBrand,
	)

	filesHandlers := files_http.NewHandlers(
		logger,
		signUpload,
		deleteFile,
	)

	authHandlers := auth_http.NewHandlers(logger)

	userHandlers := users_http.NewHandlers(
		logger,
		syncUser,
	)

	return &Composition{
		TenantHandlers: tenantHandlers,
		BrandHandlers:  brandHandlers,
		FilesHandlers:  filesHandlers,
		AuthHandlers:   authHandlers,
		UserHandlers:   userHandlers,
	}
}

