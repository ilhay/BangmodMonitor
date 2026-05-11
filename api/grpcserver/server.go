// Package grpcserver implements the gRPC MetricsService that agents use in
// Phase 5+. It runs on a separate port alongside the HTTP REST API so both
// protocols are supported during migration.
package grpcserver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"time"

	bangmodv1 "github.com/bangmodmonitor/gen/bangmod/v1"
	"github.com/bangmodmonitor/api/cache"
	"github.com/bangmodmonitor/api/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	bangmodv1.UnimplementedMetricsServiceServer
	ch         *storage.CH
	maria      *storage.Maria
	tokenCache *cache.TokenCache
	nodeSecret string
}

func New(ch *storage.CH, maria *storage.Maria, tc *cache.TokenCache, nodeSecret string) *MetricsServer {
	return &MetricsServer{ch: ch, maria: maria, tokenCache: tc, nodeSecret: nodeSecret}
}

// Start launches the gRPC server on addr (e.g. ":9090") in a goroutine.
func Start(srv *MetricsServer, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("grpc listen %s: %v", addr, err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	bangmodv1.RegisterMetricsServiceServer(s, srv)
	log.Printf("gRPC server listening on %s", addr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("grpc server: %v", err)
		}
	}()
}

func (s *MetricsServer) IngestMetrics(ctx context.Context, req *bangmodv1.IngestRequest) (*bangmodv1.IngestResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	hash := hashToken(req.Token)
	hostID, _, ok := s.tokenCache.Get(ctx, hash)
	if !ok {
		var orgID string
		hostID, orgID, ok = s.maria.ValidateToken(ctx, hash)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		s.tokenCache.Set(ctx, hash, hostID, orgID)
	}

	m := req.Metrics
	if m == nil {
		return &bangmodv1.IngestResponse{Ok: true}, nil
	}

	ts := time.Unix(m.Timestamp, 0)

	var cpuCores uint8
	var cpuPct float32
	if m.Cpu != nil {
		cpuCores = uint8(m.Cpu.Cores)
		cpuPct = float32(m.Cpu.UsagePercent)
	}
	var memTotal, memUsed uint64
	var memPct float32
	if m.Memory != nil {
		memTotal = m.Memory.TotalBytes
		memUsed = m.Memory.UsedBytes
		memPct = float32(m.Memory.UsagePercent)
	}

	if err := s.ch.InsertAgentMetric(ctx, storage.AgentMetricRow{
		Timestamp: ts, HostID: hostID, Hostname: m.Hostname, Region: "grpc",
		CPUPercent: cpuPct, CPUCores: cpuCores,
		MemTotal: memTotal, MemUsed: memUsed, MemPercent: memPct,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "store: %v", err)
	}

	var diskRows []storage.DiskRow
	for _, d := range m.Disks {
		diskRows = append(diskRows, storage.DiskRow{
			Timestamp: ts, HostID: hostID, Path: d.Path,
			TotalBytes: d.TotalBytes, UsedBytes: d.UsedBytes, UsePct: float32(d.UsagePercent),
		})
	}
	if len(diskRows) > 0 {
		_ = s.ch.InsertDiskMetrics(ctx, diskRows)
	}

	var netRows []storage.NetRow
	for _, n := range m.Networks {
		netRows = append(netRows, storage.NetRow{
			Timestamp: ts, HostID: hostID, Interface: n.Interface,
			BytesSent: n.BytesSent, BytesRecv: n.BytesRecv,
			PacketsSent: n.PacketsSent, PacketsRecv: n.PacketsRecv,
		})
	}
	if len(netRows) > 0 {
		_ = s.ch.InsertNetMetrics(ctx, netRows)
	}

	return &bangmodv1.IngestResponse{Ok: true, Message: "stored"}, nil
}

func (s *MetricsServer) IngestProbe(ctx context.Context, req *bangmodv1.ProbeRequest) (*bangmodv1.IngestResponse, error) {
	if s.nodeSecret != "" && req.NodeSecret != s.nodeSecret {
		return nil, status.Error(codes.Unauthenticated, "invalid node secret")
	}

	ts := time.Now()
	for _, r := range req.Results {
		isUp := uint8(0)
		if r.IsUp {
			isUp = 1
		}
		_ = s.ch.InsertProbeResult(ctx, storage.ProbeRow{
			Timestamp:  ts,
			HostID:     "probe",
			TargetURL:  r.Url,
			Region:     req.Region,
			StatusCode: uint16(r.StatusCode),
			ResponseMS: uint32(r.ResponseMs),
			IsUp:       isUp,
		})
	}
	return &bangmodv1.IngestResponse{Ok: true, Message: fmt.Sprintf("stored %d results", len(req.Results))}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	_ = md
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("gRPC %s %v", info.FullMethod, time.Since(start))
	return resp, err
}
