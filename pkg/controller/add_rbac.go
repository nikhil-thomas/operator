package controller

import (
	"github.com/tektoncd/operator/pkg/controller/openshift/rbac"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, rbac.Add)
}
