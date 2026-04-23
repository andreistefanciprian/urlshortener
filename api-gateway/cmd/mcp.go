package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ugen "github.com/andreistefanciprian/urlshortener/api-gateway/url-gen/proto"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newMCPServer(gw *URLGateway) *server.MCPServer {
	s := server.NewMCPServer("urlshortener", "1.0.0")

	s.AddTool(
		mcp.NewTool("list_urls",
			mcp.WithDescription("List all short URLs in the system"),
		),
		gw.mcpListURLs,
	)

	s.AddTool(
		mcp.NewTool("create_short_url",
			mcp.WithDescription("Create a short URL from a long URL"),
			mcp.WithString("long_url",
				mcp.Required(),
				mcp.Description("The long URL to shorten"),
			),
			mcp.WithNumber("expires_in",
				mcp.Description("Days until expiry (default: 7)"),
			),
		),
		gw.mcpCreateShortURL,
	)

	s.AddTool(
		mcp.NewTool("delete_short_url",
			mcp.WithDescription("Delete a short URL by its 7-character short code"),
			mcp.WithString("short_code",
				mcp.Required(),
				mcp.Description("The 7-character short code (e.g. gt0PFD8)"),
			),
		),
		gw.mcpDeleteShortURL,
	)

	return s
}

func (h *URLGateway) mcpListURLs(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := h.urlGenClient.GetAllURLs(ctx, &emptypb.Empty{})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to list URLs: %v", err), nil
	}

	urls := make([]URLInfo, 0, len(resp.Urls))
	for _, u := range resp.Urls {
		info := URLInfo{
			ShortUrl: URLScheme + "://" + URLShortenerDomainName + "/" + u.ShortUrlCode,
			LongUrl:  u.OriginalUrl,
		}
		if u.CreatedAt != nil {
			info.CreatedAt = u.CreatedAt.AsTime().Format(time.RFC3339)
		}
		if u.ExpireTime != nil {
			info.Expiration = u.ExpireTime.AsTime().Format(time.RFC3339)
		}
		urls = append(urls, info)
	}

	out, err := json.Marshal(urls)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response"), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}

func (h *URLGateway) mcpCreateShortURL(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	longURL, err := req.RequireString("long_url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	expiresIn := req.GetInt("expires_in", 7)

	urlReq := &CreateShortURLRequest{LongUrl: longURL, ExpiresIn: expiresIn}
	if err := urlReq.validateURL(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := urlReq.validateExpiration(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	grpcResp, err := h.urlGenClient.GenerateShortURL(ctx, &ugen.ShortURLRequest{
		LongUrl:    urlReq.LongUrl,
		Expiration: timestamppb.New(time.Now().Add(time.Duration(urlReq.ExpiresIn) * 24 * time.Hour)),
	})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to create short URL: %v", err), nil
	}

	grpcResp.ShortUrl = URLScheme + "://" + URLShortenerDomainName + "/" + grpcResp.ShortUrl

	out, err := json.Marshal(grpcResp)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response"), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}

func (h *URLGateway) mcpDeleteShortURL(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	shortCode, err := req.RequireString("short_code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := validateShortURL(shortCode); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	grpcResp, err := h.urlGenClient.DeleteShortURL(ctx, &ugen.DeleteShortURLRequest{
		ShortUrlCode: shortCode,
	})
	if err != nil {
		return mcp.NewToolResultErrorf("failed to delete short URL: %v", err), nil
	}
	if !grpcResp.Success {
		return mcp.NewToolResultError(fmt.Sprintf("short URL not found: %s", grpcResp.Message)), nil
	}

	out, err := json.Marshal(grpcResp)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response"), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}
