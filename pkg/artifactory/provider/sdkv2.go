package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

// SdkV2 Artifactory provider that supports configuration via Access Token
// Supported resources are repos, users, groups, replications, and permissions
func SdkV2() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
				Description:      "Artifactory URL.",
			},
			"api_key": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				ValidateDiagFunc: validator.StringIsNotEmpty,
				Description:      "API key. If `access_token` attribute, `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` environment variable is set, the provider will ignore this attribute.",
				Deprecated: "An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform).\n" +
					"In a future version (scheduled for end of Q3, 2023), the option to disable the usage/creation of API Keys will be available and set to disabled by default. Admins will be able to enable the usage/creation of API Keys.\n" +
					"By end of Q1 2024, API Keys will be deprecated all together and the option to use them will no longer be available.",
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "This is a access token that can be given to you by your admin under `User Management -> Access Tokens`. If not set, the 'api_key' attribute value will be used.",
			},
			"check_license": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Toggle for pre-flight checking of Artifactory Pro and Enterprise license. Default to `true`.",
			},
		},

		ResourcesMap:   resourcesMap(),
		DataSourcesMap: datasourcesMap(),
	}

	p.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var ds diag.Diagnostics
		meta, d := providerConfigure(ctx, data, p.TerraformVersion)
		if d != nil {
			ds = append(ds, d...)
		}
		return meta, ds
	}

	return p
}

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	// Check environment variables, first available OS variable will be assigned to the var
	url := CheckEnvVars([]string{"JFROG_URL", "ARTIFACTORY_URL"}, "")
	accessToken := CheckEnvVars([]string{"JFROG_ACCESS_TOKEN", "ARTIFACTORY_ACCESS_TOKEN"}, "")

	if v, ok := d.GetOk("url"); ok {
		url = v.(string)
	}
	if url == "" {
		return nil, diag.Errorf("missing URL Configuration")
	}

	if v, ok := d.GetOk("access_token"); ok && v != "" {
		accessToken = v.(string)
	}

	apiKey := d.Get("api_key").(string)

	restyBase, err := client.Build(url, productId)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	restyBase, err = client.AddAuth(restyBase, apiKey, accessToken)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// Due to migration from SDK v2 to plugin framework, we have to remove defaults from the provider configuration.
	// https://discuss.hashicorp.com/t/muxing-upgraded-tfsdk-and-framework-provider-with-default-provider-configuration/43945
	checkLicense := true
	if v, ok := d.GetOk("check_license"); ok {
		checkLicense = v.(bool)
	}
	if checkLicense {
		licenseErr := util.CheckArtifactoryLicense(restyBase, "Enterprise", "Commercial", "Edge")
		if licenseErr != nil {
			return nil, diag.FromErr(licenseErr)
		}
	}

	version, err := util.GetArtifactoryVersion(restyBase)
	if err != nil {
		return nil, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Error getting Artifactory version",
			Detail:   fmt.Sprintf("The provider functionality might be affected by the absence of Artifactory version in the context. %v", err),
		}}
	}

	featureUsage := fmt.Sprintf("Terraform/%s", terraformVersion)
	util.SendUsage(ctx, restyBase, productId, featureUsage)

	return util.ProviderMetadata{
		Client:             restyBase,
		ArtifactoryVersion: version,
	}, nil
}
