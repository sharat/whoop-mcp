package mcp

import (
	"context"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sharat/whoop-mcp/pkg/whoop"
)

type Server struct {
	mcpServer *mcp.Server
	client    *whoop.Client
}

type EmptyParams struct{}

type QueryParams struct {
	Start     string `json:"start,omitempty" jsonschema:"ISO 8601 start date-time, e.g., 2026-06-01T00:00:00.000Z"`
	End       string `json:"end,omitempty" jsonschema:"ISO 8601 end date-time, e.g., 2026-06-07T23:59:59.000Z"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Max number of records to return (up to 25)"`
	NextToken string `json:"next_token,omitempty" jsonschema:"Token to retrieve the next page of results"`
}

func NewServer(client *whoop.Client) *Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "whoop-mcp",
		Version: "1.0.0",
	}, nil)

	srv := &Server{
		mcpServer: s,
		client:    client,
	}

	srv.registerTools()

	return srv
}

func (s *Server) Run(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	// 1. get_profile
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_profile",
		Description: "Retrieve basic WHOOP user profile information (e.g. name, user ID, email)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args EmptyParams) (*mcp.CallToolResult, any, error) {
		body, err := s.client.GetProfile(ctx)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(body)},
			},
		}, nil, nil
	})

	// 2. get_body_measurements
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_body_measurements",
		Description: "Retrieve WHOOP user body measurements (e.g. height, weight, max heart rate)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args EmptyParams) (*mcp.CallToolResult, any, error) {
		body, err := s.client.GetBodyMeasurements(ctx)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(body)},
			},
		}, nil, nil
	})

	// 3. get_cycles
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_cycles",
		Description: "Retrieve physiological cycles containing daily strain, sleep performance, and recovery data",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryParams) (*mcp.CallToolResult, any, error) {
		limitStr := ""
		if args.Limit > 0 {
			limitStr = strconv.Itoa(args.Limit)
		}
		body, err := s.client.GetCycles(ctx, args.Start, args.End, limitStr, args.NextToken)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(body)},
			},
		}, nil, nil
	})

	// 4. get_sleeps
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_sleeps",
		Description: "Retrieve sleep activity records containing detailed sleep states, disturbances, and efficiency",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryParams) (*mcp.CallToolResult, any, error) {
		limitStr := ""
		if args.Limit > 0 {
			limitStr = strconv.Itoa(args.Limit)
		}
		body, err := s.client.GetSleeps(ctx, args.Start, args.End, limitStr, args.NextToken)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(body)},
			},
		}, nil, nil
	})

	// 5. get_recoveries
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_recoveries",
		Description: "Retrieve daily recovery metrics including HRV, resting heart rate, and recovery score",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryParams) (*mcp.CallToolResult, any, error) {
		limitStr := ""
		if args.Limit > 0 {
			limitStr = strconv.Itoa(args.Limit)
		}
		body, err := s.client.GetRecoveries(ctx, args.Start, args.End, limitStr, args.NextToken)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(body)},
			},
		}, nil, nil
	})

	// 6. get_workouts
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_workouts",
		Description: "Retrieve workout activities including strain, duration, average heart rate, and calories burned",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryParams) (*mcp.CallToolResult, any, error) {
		limitStr := ""
		if args.Limit > 0 {
			limitStr = strconv.Itoa(args.Limit)
		}
		body, err := s.client.GetWorkouts(ctx, args.Start, args.End, limitStr, args.NextToken)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(body)},
			},
		}, nil, nil
	})
}
