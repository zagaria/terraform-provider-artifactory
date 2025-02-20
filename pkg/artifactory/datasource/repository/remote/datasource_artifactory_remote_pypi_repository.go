package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func DataSourceArtifactoryRemotePypiRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.PyPiPackageType)
		if err != nil {
			return nil, err
		}

		return &remote.PypiRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.PyPiPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	pypiSchema := getSchema(remote.PyPiSchemas)

	return &schema.Resource{
		Schema:      pypiSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(pypiSchema), constructor),
		Description: "Provides a data source for a remote Pypi repository",
	}
}
