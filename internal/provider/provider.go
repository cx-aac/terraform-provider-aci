package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ciscoecosystem/aci-go-client/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"username": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("ACI_USERNAME", nil),
					Description: "Username for the APIC Account. This can also be set as the ACI_USERNAME environment variable.",
				},
				"password": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ACI_PASSWORD", nil),
					Description: "Password for the APIC Account. This can also be set as the ACI_PASSWORD environment variable.",
				},
				"url": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("ACI_URL", nil),
					Description: "URL of the Cisco ACI web interface. This can also be set as the ACI_URL environment variable.",
				},
				"insecure": {
					Type:     schema.TypeBool,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						if v := os.Getenv("ACI_INSECURE"); v != "" {
							return strconv.ParseBool(v)
						}
						return true, nil
					},
					Description: "Allow insecure HTTPS client. This can also be set as the ACI_INSECURE environment variable. Defaults to `true`.",
				},
				"private_key": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ACI_PRIVATE_KEY", nil),
					Description: "Private key path for signature calculation. This can also be set as the ACI_PRIVATE_KEY environment variable.",
				},
				"cert_name": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ACI_CERT_NAME", nil),
					Description: "Certificate name for the User in Cisco ACI. This can also be set as the ACI_CERT_NAME environment variable.",
				},
				"proxy_url": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ACI_PROXY_URL", nil),
					Description: "Proxy Server URL with port number. This can also be set as the ACI_PROXY_URL environment variable.",
				},
				"retries": {
					Type:     schema.TypeInt,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						if v := os.Getenv("ACI_RETRIES"); v != "" {
							return strconv.Atoi(v)
						}
						return 3, nil
					},
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						v := val.(int)
						if v < 0 || v > 9 {
							errs = append(errs, fmt.Errorf("%q must be between 0 and 9 inclusive, got: %d", key, v))
						}
						return
					},
					Description: "Number of retries for REST API calls. This can also be set as the ACI_RETRIES environment variable. Defaults to `3`.",
				},
				"mock": {
					Type:     schema.TypeBool,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						if v := os.Getenv("ACI_MOCK"); v != "" {
							return strconv.ParseBool(v)
						}
						return false, nil
					},
					Description: "Only mock API calls. This is mainly for troubleshooting/debugging purposes. This can also be set as the ACI_MOCK environment variable. Defaults to `false`.",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"aci_rest": dataSourceAciRest(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"aci_rest": resourceAciRest(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	Username   string
	Password   string
	URL        string
	IsInsecure bool
	PrivateKey string
	Certname   string
	ProxyUrl   string
	Retries    int
	IsMock     bool
	Client     *client.Client
}

func (c apiClient) Valid() diag.Diagnostics {

	if c.Username == "" {
		return diag.FromErr(fmt.Errorf("Username must be provided for the ACI provider"))
	}

	if c.Password == "" {
		if c.PrivateKey == "" && c.Certname == "" {

			return diag.FromErr(fmt.Errorf("Either of private_key/cert_name or password is required"))
		} else if c.PrivateKey == "" || c.Certname == "" {
			return diag.FromErr(fmt.Errorf("private_key and cert_name both must be provided"))
		}
	}

	if c.URL == "" {
		return diag.FromErr(fmt.Errorf("The URL must be provided for the ACI provider"))
	}

	return nil
}

func (c apiClient) getClient() interface{} {
	if c.Password != "" {
		return client.GetClient(c.URL, c.Username, client.Password(c.Password), client.Insecure(c.IsInsecure), client.ProxyUrl(c.ProxyUrl))
	} else {
		return client.GetClient(c.URL, c.Username, client.PrivateKey(c.PrivateKey), client.AdminCert(c.Certname), client.Insecure(c.IsInsecure), client.ProxyUrl(c.ProxyUrl))
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(c context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		cl := apiClient{
			Username:   d.Get("username").(string),
			Password:   d.Get("password").(string),
			URL:        d.Get("url").(string),
			IsInsecure: d.Get("insecure").(bool),
			PrivateKey: d.Get("private_key").(string),
			Certname:   d.Get("cert_name").(string),
			ProxyUrl:   d.Get("proxy_url").(string),
			Retries:    d.Get("retries").(int),
			IsMock:     d.Get("mock").(bool),
		}

		if diag := cl.Valid(); diag != nil {
			return nil, diag
		}

		cl.Client = cl.getClient().(*client.Client)

		return cl, nil
	}
}
