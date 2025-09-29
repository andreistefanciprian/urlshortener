// Package proto provides gRPC protocol buffer definitions for the URL Generator service.
//
// This package contains the generated Go code from the url-gen.proto file,
// which defines the URLGenerator service for creating short URLs from long URLs.
//
// The main service provided is:
//   - URLGenerator: Generates short URLs with optional expiration dates
//
// Example usage:
//
//	client := proto.NewURLGeneratorClient(conn)
//	response, err := client.GenerateShortURL(ctx, &proto.LongURLRequest{
//		LongUrl: "https://example.com",
//		Expiration: timestamppb.New(time.Now().Add(24 * time.Hour)),
//	})
package proto
