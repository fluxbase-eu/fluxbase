package api

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/email"
	"github.com/nimbleflux/fluxbase/internal/middleware"
	"github.com/nimbleflux/fluxbase/internal/tenantdb"
)

type TenancyModule struct {
	Handlers   *TenancyHandlers
	Middleware *MiddlewareComponents

	// ctx/cancel scope the module's fire-and-forget startup goroutines so
	// Shutdown can stop them instead of leaving them running (and sleeping)
	// for the lifetime of the process.
	ctx    context.Context
	cancel context.CancelFunc
}

func (m *TenancyModule) Name() string { return "tenancy" }

func (m *TenancyModule) Init(ctx context.Context, registry *ServiceRegistry) error {
	m.ctx, m.cancel = context.WithCancel(context.Background())

	cfg := registry.Config
	db := registry.DB

	m.Handlers = &TenancyHandlers{}
	m.Handlers.ServiceKey = NewServiceKeyHandler(db)

	if !cfg.Tenants.Enabled {
		m.Handlers.Tenant = NewTenantHandler(db, nil, nil, nil, nil, cfg)
		return nil
	}

	tenantStorage := tenantdb.NewRepository(db.Pool())
	dbURL := cfg.Database.RuntimeConnectionString()
	tenantCfg := tenantdb.Config{
		Enabled:        cfg.Tenants.Enabled,
		DatabasePrefix: cfg.Tenants.DatabasePrefix,
		MaxTenants:     cfg.Tenants.MaxTenants,
		Pool: tenantdb.PoolConfig{
			MaxTotalConnections: cfg.Tenants.Pool.MaxTotalConnections,
			EvictionAge:         cfg.Tenants.Pool.EvictionAge,
		},
		Migrations: tenantdb.MigrationsConfig{
			CheckInterval: cfg.Tenants.Migrations.CheckInterval,
			OnCreate:      cfg.Tenants.Migrations.OnCreate,
			OnAccess:      cfg.Tenants.Migrations.OnAccess,
			Background:    cfg.Tenants.Migrations.Background,
		},
	}
	// Prefer a dedicated admin pool for tenant-database management (FDW role
	// creation, schema grants): the runtime user typically cannot grant on
	// bootstrap-owned schemas. Falls back to the application pool when admin
	// credentials are not configured.
	tenantAdminPool := db.Pool()
	if adminDBURL := cfg.Database.AdminConnectionString(); adminDBURL != "" {
		if pool, err := pgxpool.New(ctx, adminDBURL); err == nil {
			tenantAdminPool = pool
		} else {
			log.Warn().Err(err).Msg("Failed to create admin pool for tenant management; using the application pool")
		}
	}
	tenantManager := tenantdb.NewManager(tenantStorage, tenantCfg, tenantAdminPool, dbURL)
	tenantManager.SetAdminDBURL(cfg.Database.AdminConnectionString())

	if adminDBURL := cfg.Database.AdminConnectionString(); adminDBURL != "" {
		fdwCfg, fdwErr := tenantdb.ParseFDWConfig(adminDBURL)
		if fdwErr != nil {
			log.Warn().Err(fdwErr).Msg("Failed to parse FDW config, FDW disabled for tenant databases")
		} else {
			tenantManager.SetFDWConfig(fdwCfg)
			log.Info().Msg("FDW enabled for tenant databases")

			go func() {
				defer func() {
					if rec := recover(); rec != nil {
						log.Error().
							Interface("panic", rec).
							Str("module", m.Name()).
							Msg("Panic in tenant FDW upgrade goroutine - recovered")
					}
				}()
				select {
				case <-time.After(5 * time.Second):
				case <-m.ctx.Done():
					return
				}
				upgCtx, cancel := context.WithTimeout(m.ctx, 10*time.Minute)
				defer cancel()
				tenantManager.UpgradeAllTenantsFDW(upgCtx)
			}()
		}
	}

	tenantRouter := tenantdb.NewRouter(tenantStorage, tenantCfg, db.Pool(), db.Pool(), dbURL)
	tenantRouter.SetManager(tenantManager)
	tenantManager.SetRouter(tenantRouter)

	log.Info().Msg("Multi-tenancy enabled")

	if cfg.Tenants.Declarative.Enabled && cfg.Tenants.Declarative.SchemaDir != "" {
		declarativeCfg := tenantdb.DeclarativeConfig{
			Enabled:          cfg.Tenants.Declarative.Enabled,
			SchemaDir:        cfg.Tenants.Declarative.SchemaDir,
			OnCreate:         cfg.Tenants.Declarative.OnCreate,
			OnStartup:        cfg.Tenants.Declarative.OnStartup,
			AllowDestructive: cfg.Tenants.Declarative.AllowDestructive,
		}
		declarativeSvc := tenantdb.NewDeclarativeService(
			declarativeCfg,
			"pgschema",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.AdminUser,
			cfg.Database.AdminPassword,
			db.Pool(),
		)
		tenantManager.SetDeclarativeService(declarativeSvc)
		tenantManager.SetDeclarativeConfig(declarativeCfg)
		log.Info().
			Str("schema_dir", cfg.Tenants.Declarative.SchemaDir).
			Bool("on_create", cfg.Tenants.Declarative.OnCreate).
			Bool("on_startup", cfg.Tenants.Declarative.OnStartup).
			Msg("Tenant declarative schema service initialized")

		if cfg.Tenants.Declarative.OnStartup {
			go func() {
				defer func() {
					if rec := recover(); rec != nil {
						log.Error().
							Interface("panic", rec).
							Str("module", m.Name()).
							Msg("Panic in tenant declarative schema startup goroutine - recovered")
					}
				}()
				schemaCtx, cancel := context.WithTimeout(m.ctx, 5*time.Minute)
				defer cancel()
				if err := tenantManager.ApplyDeclarativeSchemas(schemaCtx); err != nil {
					log.Error().Err(err).Msg("Failed to apply tenant declarative schemas on startup")
				}
			}()
		}
	}

	m.Handlers.Manager = tenantManager
	m.Handlers.Storage = tenantStorage

	invitationService := GetService[*auth.InvitationService](registry)
	lazyEmail := GetService[*email.LazyService](registry)
	var emailSvc email.Service
	if lazyEmail != nil {
		emailSvc = lazyEmail
	}
	m.Handlers.Tenant = NewTenantHandler(db, tenantManager, tenantStorage, invitationService, emailSvc, cfg)

	if tenantManager.GetRouter() != nil {
		m.Middleware = &MiddlewareComponents{
			TenantDB: middleware.TenantDBMiddleware(middleware.TenantDBConfig{
				Router:     tenantManager.GetRouter(),
				Repository: tenantStorage,
			}),
		}
	}

	registry.Register(tenantManager)
	registry.Register(tenantStorage)
	return nil
}

func (m *TenancyModule) Shutdown(ctx context.Context) error {
	// Stop any pending fire-and-forget startup goroutines.
	if m.cancel != nil {
		m.cancel()
	}
	if m.Handlers != nil && m.Handlers.Manager != nil && m.Handlers.Manager.GetRouter() != nil {
		m.Handlers.Manager.GetRouter().Close()
	}
	return nil
}
