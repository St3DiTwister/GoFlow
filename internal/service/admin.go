package service

import (
	"GoFlow/internal/gen/admin"
	"GoFlow/internal/storage"
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
)

type AdminServer struct {
	admin.UnimplementedAdminServiceServer
	pg *storage.PostgresStorage
}

func NewAdminServer(pg *storage.PostgresStorage) *AdminServer {
	return &AdminServer{pg: pg}
}

func (s *AdminServer) GetValidSites(ctx context.Context, _ *admin.GetValidSitesRequest) (*admin.GetValidSitesResponse, error) {
	sitesMap, err := s.pg.GetValidSiteIDs(ctx)
	if err != nil {
		slog.Error("grpc_get_sites_failed", "error", err)
		return nil, err
	}

	ids := make([]string, 0, len(sitesMap))
	for id := range sitesMap {
		ids = append(ids, id)
	}

	return &admin.GetValidSitesResponse{SiteIds: ids}, nil
}

func (s *AdminServer) RegisterSite(ctx context.Context, req *admin.RegisterSiteRequest) (*admin.RegisterSiteResponse, error) {
	apiKey := generateAPIKey()
	siteID, err := s.pg.SaveSite(ctx, req, apiKey)
	if err != nil {
		slog.Error("grpc_register_site_failed", "error", err)
		return nil, err

	}

	slog.Info("site_registered", "name", req.Name, "id", siteID)

	return &admin.RegisterSiteResponse{
		SiteId: siteID,
		ApiKey: apiKey,
	}, nil
}

func generateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
