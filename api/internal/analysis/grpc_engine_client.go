package analysis

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kishan-thanki/resumeranker/api/internal/analysis/pb"
)

type GrpcEngineClient struct {
	conn   *grpc.ClientConn
	client pb.AnalysisEngineClient
}

func NewGrpcEngineClient(target string) (*GrpcEngineClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to engine gRPC server at %s: %w", target, err)
	}

	client := pb.NewAnalysisEngineClient(conn)
	return &GrpcEngineClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *GrpcEngineClient) Close() error {
	return c.conn.Close()
}

func (c *GrpcEngineClient) Analyze(ctx context.Context, req *EngineRequest) (*EngineResponse, error) {
	pbReq := &pb.AnalyzeRequest{
		ResumePdf:         req.ResumePDF,
		JobDescriptionPdf: req.JobDescriptionPDF,
		RequestId:         req.RequestID,
	}

	pbRes, err := c.client.Analyze(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("engine gRPC Analyze call failed: %w", err)
	}

	if !pbRes.Success {
		return nil, fmt.Errorf("analysis engine error: %s", pbRes.ErrorMessage)
	}

	return &EngineResponse{
		ResultJSON:       pbRes.ResultJson,
		Model:            pbRes.Model,
		PromptTokens:     pbRes.PromptTokens,
		CompletionTokens: pbRes.CompletionTokens,
		TotalTokens:      pbRes.TotalTokens,
	}, nil
}
