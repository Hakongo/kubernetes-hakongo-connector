package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	hakongov1alpha1 "github.com/hakongo/kubernetes-connector/api/v1alpha1"
)

func TestConnectorConfig(t *testing.T) {
	// Register our scheme
	s := scheme.Scheme
	hakongov1alpha1.AddToScheme(s)

	// Create a fake client
	cli := fake.NewClientBuilder().WithScheme(s).Build()

	// Create a test ConnectorConfig
	cfg := &hakongov1alpha1.ConnectorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: hakongov1alpha1.ConnectorConfigSpec{
			HakonGo: hakongov1alpha1.HakonGoConfig{
				BaseURL: "https://api.hakongo.io",
				APIKey: hakongov1alpha1.SecretKeyRef{
					Name: "hakongo-secret",
					Key:  "api-key",
				},
			},
			ClusterContext: hakongov1alpha1.ClusterContextConfig{
				Name:   "test-cluster",
				Type:   "aws",
				Region: "us-west-2",
			},
		},
	}

	// Create the config
	err := cli.Create(context.Background(), cfg)
	assert.NoError(t, err)

	// Verify we can get it back
	var found hakongov1alpha1.ConnectorConfig
	err = cli.Get(context.Background(), client.ObjectKey{
		Name:      "test-config",
		Namespace: "default",
	}, &found)
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", found.Spec.ClusterContext.Name)
}
