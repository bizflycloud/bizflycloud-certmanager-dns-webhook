package bizflycloud

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bizflycloud/gobizfly"
	"github.com/jetstack/cert-manager/pkg/issuer/acme/dns/util"
)

const (
	// ProviderName specifies the name for the Bizfly provider
	ProviderName  string = "bizflycloud"
	defaultRegion string = "HN"
	authPassword  string = "password"
	authAppCred   string = "application_credential"
	defaultApiUrl string = "https://manage.bizflycloud.vn"

	bizflyCloudAuthMethod      string = "BIZFLYCLOUD_AUTH_METHOD"
	bizflyCloudEmailEnvName    string = "BIZFLYCLOUD_EMAIL"
	bizflyCloudPasswordEnvName string = "BIZFLYCLOUD_PASSWORD"
	bizflyCloudRegionEnvName   string = "BIZFLYCLOUD_REGION"
	bizflyCloudAppCredID       string = "BIZFLYCLOUD_APP_CREDENTIAL_ID"
	bizflyCloudAppCredSecret   string = "BIZFLYCLOUD_APP_CREDENTIAL_SECRET"
	bizflyCloudApiUrl          string = "BIZFLYCLOUD_API_URL"
	bizflyCloudTenantID        string = "BIZFLYCLOUD_TENANT_ID"
)

type Client struct {
	dnsc *gobizfly.Client
}

func newClient() (*Client, error) {
	authMethod := os.Getenv(bizflyCloudAuthMethod)
	username := os.Getenv(bizflyCloudEmailEnvName)
	password := os.Getenv(bizflyCloudPasswordEnvName)
	region := os.Getenv(bizflyCloudRegionEnvName)
	appCredId := os.Getenv(bizflyCloudAppCredID)
	appCredSecret := os.Getenv(bizflyCloudAppCredSecret)
	apiUrl := os.Getenv(bizflyCloudApiUrl)
	tenantId := os.Getenv(bizflyCloudTenantID)

	switch authMethod {
	case authPassword:
		{
			if username == "" {
				return nil, errors.New("you have to provide username variable")
			}
			if password == "" {
				return nil, errors.New("you have to provide password variable")
			}
		}
	case authAppCred:
		{
			if appCredId == "" {
				return nil, errors.New("you have to provide application credential ID")
			}
			if appCredSecret == "" {
				return nil, errors.New("you have to provide application credential secret")
			}
		}
	}

	if region == "" {
		region = defaultRegion
	}

	if apiUrl == "" {
		apiUrl = defaultApiUrl
	}

	bizflyClient, err := gobizfly.NewClient(gobizfly.WithTenantName(username), gobizfly.WithAPIUrl(apiUrl), gobizfly.WithTenantID(tenantId), gobizfly.WithRegionName(region))
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize Bizflycloud client: %s", err)
	}

	token, err := bizflyClient.Token.Create(
		context.Background(),
		&gobizfly.TokenCreateRequest{
			AuthMethod:    authMethod,
			Username:      username,
			Password:      password,
			AppCredID:     appCredId,
			AppCredSecret: appCredSecret})
	if err != nil {
		return nil, fmt.Errorf("cannot create token: %w", err)
	}

	bizflyClient.SetKeystoneToken(token.KeystoneToken)

	return &Client{dnsc: bizflyClient}, nil
}

func (c *Client) domainNameToZoneID(fqdn string) (string, error) {

	var zoneID string
	opts := &gobizfly.ListOptions{}

	zoneName := fqdn

	allZone, err := c.dnsc.DNS.ListZones(context.Background(), opts)
	if err != nil {
		return "", err
	}

	if last := len(zoneName) - 1; last >= 0 && zoneName[last] == '.' {
		zoneName = zoneName[:last]
	}
	for _, i := range allZone.Zones {
		if i.Name == zoneName {
			zoneID = i.ID
		}
	}

	return zoneID, err
}

func (c *Client) findTxtRecord(zonename string, fqdn string) ([]gobizfly.RecordSet, string, error) {

	var ID string

	zoneName := zonename
	if last := len(zoneName) - 1; last >= 0 && zoneName[last] == '.' {
		zoneName = zoneName[:last]
	}

	zoneID, err := c.domainNameToZoneID(zoneName)
	if err != nil {
		return nil, "", err
	}

	getZone, err := c.dnsc.DNS.GetZone(
		context.Background(),
		zoneID,
	)
	if err != nil {
		return nil, "", err
	}

	allRecords := getZone.RecordsSet

	targetName := fqdn
	if last := len(targetName) - 1; last >= 0 && targetName[last] == '.' {
		targetName = targetName[:last]
	}
	targetName = targetName[0 : len(targetName)-len(zoneName)]

	for _, record := range allRecords {
		if util.ToFqdn(record.Name) == targetName {
			ID = record.ID
		}
	}

	return allRecords, ID, err
}
