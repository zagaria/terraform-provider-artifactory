package federated

import (
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/federated"
)

const rclass = "federated"

var federatedSchemaV4 = federated.SchemaGeneratorV4(false)
