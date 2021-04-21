// +build tools

package tools

import (
	// used in `update-codegen.sh`
	_ "k8s.io/code-generator/cmd/client-gen"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "k8s.io/code-generator/cmd/defaulter-gen"
	_ "k8s.io/code-generator/cmd/informer-gen"
	_ "k8s.io/code-generator/cmd/lister-gen"
	_ "k8s.io/code-generator/pkg/util"

	// need to have a local clone of googleapi .proto files to generate protos
	_ "github.com/googleapis/googleapis/google/example/endpointsapis/goapp"
)
