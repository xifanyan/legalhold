package client

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

type MsGraphClientBuilder struct {
	tenantID   string
	clientID   string
	certFile   string
	certSecret string
}

type MsGraphClient struct {
	AuthProvider *auth.AzureIdentityAuthenticationProvider
	Adapter      *msgraphsdk.GraphRequestAdapter
	*msgraphsdk.GraphServiceClient
}

func NewMsGraphClientBuilder() *MsGraphClientBuilder {
	return &MsGraphClientBuilder{}
}

func (b *MsGraphClientBuilder) WithTenantID(tenantID string) *MsGraphClientBuilder {
	b.tenantID = tenantID
	return b
}

func (b *MsGraphClientBuilder) WithClientID(clientID string) *MsGraphClientBuilder {
	b.clientID = clientID
	return b
}

func (b *MsGraphClientBuilder) WithCertFile(certFile string) *MsGraphClientBuilder {
	b.certFile = certFile
	return b
}

func (b *MsGraphClientBuilder) WithCertSecret(secret string) *MsGraphClientBuilder {
	b.certSecret = secret
	return b
}

func (b *MsGraphClientBuilder) Build() (*MsGraphClient, error) {
	certs, err := os.ReadFile(b.certFile)
	if err != nil {
		return nil, fmt.Errorf("error reading certificate file: %v", err)
	}

	certData, key, err := azidentity.ParseCertificates(certs, []byte(b.certSecret))
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate: %v", err)
	}

	// Create a new instance of the certificate credential
	cred, err := azidentity.NewClientCertificateCredential(b.tenantID, b.clientID, certData, key, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating credentials: %v", err)
	}

	// Create an instance of the Azure identity authentication provider
	authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("error creating authentication provider: %v", err)
	}

	// Create a request adapter using the authentication provider
	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return nil, fmt.Errorf("error creating request adapter: %v", err)
	}

	// Create a new instance of the Graph client
	client := &MsGraphClient{
		AuthProvider:       authProvider,
		Adapter:            adapter,
		GraphServiceClient: msgraphsdk.NewGraphServiceClient(adapter),
	}

	return client, nil
}
