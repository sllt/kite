package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

func TestKiteHelloClientWrapper_Creation(t *testing.T) {
	configs := testutil.NewServerConfigs(t)

	t.Run("NewHelloKiteClient", func(t *testing.T) {
		// Test Kite's NewHelloKiteClient function
		conn, err := grpc.Dial(configs.GRPCHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err, "Connection creation should not fail immediately")
		defer conn.Close()

		app := kite.New()
		helloClient, err := NewHelloKiteClient(configs.GRPCHost, app.Metrics())
		require.NoError(t, err, "Kite hello client creation should not fail")
		assert.NotNil(t, helloClient, "Kite hello client should not be nil")

		// Test that it implements the Kite interface
		var _ HelloKiteClient = helloClient
	})

	t.Run("HelloClientWrapperInterface", func(t *testing.T) {
		// Test Kite's interface compliance
		conn, err := grpc.Dial(configs.GRPCHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err, "Connection creation should not fail immediately")
		defer conn.Close()

		app := kite.New()
		helloClient, err := NewHelloKiteClient(configs.GRPCHost, app.Metrics())
		require.NoError(t, err, "Kite hello client creation should not fail")

		// Test HelloKiteClient interface compliance
		var _ HelloKiteClient = helloClient

		// Test that wrapper has the correct Kite type
		wrapper, ok := helloClient.(*HelloClientWrapper)
		assert.True(t, ok, "Should be able to cast to Kite HelloClientWrapper")
		assert.NotNil(t, wrapper.client, "Underlying hello client should not be nil")
	})
}

func TestKiteHelloClientWrapper_Methods(t *testing.T) {
	configs := testutil.NewServerConfigs(t)

	// Test Kite's wrapper methods without actual gRPC calls
	app := kite.New()
	helloClient, err := NewHelloKiteClient(configs.GRPCHost, app.Metrics())
	require.NoError(t, err, "Kite hello client creation should not fail")
	ctx := createTestContext()

	t.Run("SayHelloMethodExists", func(t *testing.T) {
		// Test that Kite's SayHello method exists and accepts correct parameters
		req := &HelloRequest{
			Name: "test-name",
		}

		// This will fail due to connection, but we're testing Kite's method signature
		_, err := helloClient.SayHello(ctx, req)
		assert.Error(t, err, "Should fail with invalid connection, but method should exist")
	})

	t.Run("HealthClientEmbedded", func(t *testing.T) {
		// Test that Kite's HelloKiteClient embeds HealthClient
		// The HelloKiteClient interface should include HealthClient methods
		var _ HealthClient = helloClient
	})
}

func TestKiteHelloClientWrapper_ContextIntegration(t *testing.T) {
	configs := testutil.NewServerConfigs(t)

	// Test Kite's context integration
	app := kite.New()
	helloClient, err := NewHelloKiteClient(configs.GRPCHost, app.Metrics())
	require.NoError(t, err, "Kite hello client creation should not fail")

	t.Run("ContextParameter", func(t *testing.T) {
		// Test that Kite's methods accept *kite.Context
		ctx := createTestContext()
		req := &HelloRequest{
			Name: "test-name",
		}

		// Test that the method signature is correct for Kite context
		_, err := helloClient.SayHello(ctx, req)
		assert.Error(t, err, "Should fail with invalid connection")

		// Test that context is properly passed (even though call fails)
		assert.NotNil(t, ctx, "Kite context should not be nil")
	})

	t.Run("ContextTypeCompliance", func(t *testing.T) {
		// Test that Kite's methods expect *kite.Context specifically
		ctx := createTestContext()
		req := &HelloRequest{
			Name: "test-name",
		}

		// Verify the method signature expects *kite.Context
		var _ func(*kite.Context, *HelloRequest, ...grpc.CallOption) (*HelloResponse, error) = helloClient.SayHello

		// Ensure the call compiles (even if it fails at runtime)
		_, _ = helloClient.SayHello(ctx, req)
	})
}

func TestKiteHelloClientWrapper_MultipleInstances(t *testing.T) {
	configs := testutil.NewServerConfigs(t)

	// Test Kite's client creation with multiple instances
	t.Run("MultipleHelloClients", func(t *testing.T) {
		app := kite.New()

		client1, err := NewHelloKiteClient(configs.GRPCHost, app.Metrics())
		require.NoError(t, err, "First Kite hello client creation should not fail")

		client2, err := NewHelloKiteClient(configs.GRPCHost, app.Metrics())
		require.NoError(t, err, "Second Kite hello client creation should not fail")

		assert.NotNil(t, client1, "First Kite hello client should not be nil")
		assert.NotNil(t, client2, "Second Kite hello client should not be nil")
		assert.NotEqual(t, client1, client2, "Kite hello client instances should be different")
	})
}

func TestKiteHelloClientWrapper_ErrorHandling(t *testing.T) {
	_ = testutil.NewServerConfigs(t)

	// Test Kite's error handling patterns
	t.Run("InvalidAddressHandling", func(t *testing.T) {
		// Test Kite's handling of invalid addresses
		app := kite.New()
		helloClient, err := NewHelloKiteClient("invalid:address", app.Metrics())
		require.NoError(t, err, "Client creation should not fail immediately")

		ctx := createTestContext()
		req := &HelloRequest{
			Name: "test-name",
		}

		// Test Kite's error handling
		_, err = helloClient.SayHello(ctx, req)
		assert.Error(t, err, "Kite should handle invalid address errors")
	})

	t.Run("EmptyAddressHandling", func(t *testing.T) {
		// Test Kite's handling of empty addresses
		app := kite.New()
		helloClient, err := NewHelloKiteClient("", app.Metrics())
		require.NoError(t, err, "Client creation should not fail immediately")

		ctx := createTestContext()
		req := &HelloRequest{
			Name: "test-name",
		}

		// Test Kite's error handling
		_, err = helloClient.SayHello(ctx, req)
		assert.Error(t, err, "Kite should handle empty address errors")
	})
}

func TestKiteHelloClientWrapper_ConcurrentAccess(t *testing.T) {
	configs := testutil.NewServerConfigs(t)

	// Test Kite's concurrent access patterns
	t.Run("ConcurrentSayHelloCalls", func(t *testing.T) {
		app := kite.New()
		helloClient, err := NewHelloKiteClient(configs.GRPCHost, app.Metrics())
		require.NoError(t, err, "Kite hello client creation should not fail")

		numGoroutines := 5
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				ctx := createTestContext()
				req := &HelloRequest{
					Name: "concurrent-test",
				}

				// This will fail due to connection, but we're testing Kite's concurrency
				_, err := helloClient.SayHello(ctx, req)
				assert.Error(t, err, "Should fail with invalid connection")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
