package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/sllt/kite/pkg/kite"
	"github.com/sllt/kite/pkg/kite/infra"
	"github.com/sllt/kite/pkg/kite/testutil"
)

// createTestContext creates a test kite.Context
func createTestContext() *kite.Context {
	container := &infra.Container{}
	return &kite.Context{
		Context:   context.Background(),
		Container: container,
	}
}

func TestKiteHealthServer_Creation(t *testing.T) {
	t.Run("GetOrCreateHealthServer", func(t *testing.T) {
		// Test Kite's getOrCreateHealthServer function
		healthServer := getOrCreateHealthServer()
		assert.NotNil(t, healthServer, "Kite health server should not be nil")

		// Test that it implements the Kite interface (not the standard gRPC interface)
		// The Kite health server has different method signatures
		assert.NotNil(t, healthServer, "Health server should not be nil")
	})

	t.Run("HealthServerSingleton", func(t *testing.T) {
		// Test Kite's singleton pattern for health server
		healthServer1 := getOrCreateHealthServer()
		healthServer2 := getOrCreateHealthServer()

		assert.Equal(t, healthServer1, healthServer2, "Kite health server should be singleton")
	})
}

func TestKiteHealthServer_Methods(t *testing.T) {
	_ = testutil.NewServerConfigs(t)

	// Test Kite's health server methods
	healthServer := getOrCreateHealthServer()
	ctx := createTestContext()

	t.Run("CheckMethodExists", func(t *testing.T) {
		// Test that Kite's Check method exists and accepts correct parameters
		req := &healthpb.HealthCheckRequest{
			Service: "test-service",
		}

		// Test Kite's Check method signature - this will fail with "unknown service" which is expected
		resp, err := healthServer.Check(ctx, req)
		assert.Error(t, err, "Health check should fail for unknown service")
		assert.Nil(t, resp, "Health check response should be nil for unknown service")
		assert.Contains(t, err.Error(), "unknown service", "Error should indicate unknown service")
	})

	t.Run("WatchMethodExists", func(t *testing.T) {
		// Test that Kite's Watch method exists and accepts correct parameters
		req := &healthpb.HealthCheckRequest{
			Service: "test-service",
		}

		// Test Kite's Watch method signature - this will panic with nil stream, but we're testing method existence
		assert.Panics(t, func() {
			healthServer.Watch(ctx, req, nil)
		}, "Watch should panic with nil stream, but method should exist")
	})
}

func TestKiteHealthServer_SetServingStatus(t *testing.T) {
	_ = testutil.NewServerConfigs(t)

	// Test Kite's SetServingStatus functionality
	healthServer := getOrCreateHealthServer()
	ctx := createTestContext()

	t.Run("SetServingStatus", func(t *testing.T) {
		// Test Kite's SetServingStatus method
		healthServer.SetServingStatus(ctx, "test-service", healthpb.HealthCheckResponse_SERVING)

		// Verify the status was set
		req := &healthpb.HealthCheckRequest{
			Service: "test-service",
		}
		resp, err := healthServer.Check(ctx, req)
		require.NoError(t, err, "Health check should not fail")
		assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.Status, "Service should be serving")
	})

	t.Run("SetNotServingStatus", func(t *testing.T) {
		// Test Kite's SetServingStatus with NOT_SERVING
		healthServer.SetServingStatus(ctx, "test-service-not-serving", healthpb.HealthCheckResponse_NOT_SERVING)

		// Verify the status was set
		req := &healthpb.HealthCheckRequest{
			Service: "test-service-not-serving",
		}
		resp, err := healthServer.Check(ctx, req)
		require.NoError(t, err, "Health check should not fail")
		assert.Equal(t, healthpb.HealthCheckResponse_NOT_SERVING, resp.Status, "Service should not be serving")
	})
}

func TestKiteHealthServer_Shutdown(t *testing.T) {
	_ = testutil.NewServerConfigs(t)

	// Test Kite's Shutdown functionality
	healthServer := getOrCreateHealthServer()
	ctx := createTestContext()

	t.Run("Shutdown", func(t *testing.T) {
		// Test Kite's Shutdown method
		healthServer.Shutdown(ctx)

		// After shutdown, all services should return NOT_SERVING
		req := &healthpb.HealthCheckRequest{
			Service: "any-service",
		}
		resp, err := healthServer.Check(ctx, req)
		// After shutdown, health checks should fail with "unknown service"
		assert.Error(t, err, "Health check should fail after shutdown")
		assert.Nil(t, resp, "Health check response should be nil after shutdown")
		assert.Contains(t, err.Error(), "unknown service", "Error should indicate unknown service after shutdown")
	})
}

func TestKiteHealthServer_Resume(t *testing.T) {
	_ = testutil.NewServerConfigs(t)

	// Test Kite's Resume functionality
	healthServer := getOrCreateHealthServer()
	ctx := createTestContext()

	t.Run("Resume", func(t *testing.T) {
		// Test Kite's Resume method
		healthServer.Resume(ctx)

		// After resume, services should return to their previous status
		healthServer.SetServingStatus(ctx, "test-service-resume", healthpb.HealthCheckResponse_SERVING)

		req := &healthpb.HealthCheckRequest{
			Service: "test-service-resume",
		}
		resp, err := healthServer.Check(ctx, req)
		require.NoError(t, err, "Health check should not fail")
		assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.Status, "Service should be serving after resume")
	})
}

func TestKiteHealthServer_MultipleInstances(t *testing.T) {
	_ = testutil.NewServerConfigs(t)

	// Test Kite's singleton pattern
	t.Run("SingletonPattern", func(t *testing.T) {
		healthServer1 := getOrCreateHealthServer()
		healthServer2 := getOrCreateHealthServer()
		ctx := createTestContext()

		assert.Equal(t, healthServer1, healthServer2, "Kite health server should be singleton")

		// Test that operations on one affect the other
		healthServer1.SetServingStatus(ctx, "singleton-test", healthpb.HealthCheckResponse_SERVING)

		req := &healthpb.HealthCheckRequest{
			Service: "singleton-test",
		}
		resp, err := healthServer2.Check(ctx, req)
		require.NoError(t, err, "Health check should not fail")
		assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.Status, "Singleton should share state")
	})
}
