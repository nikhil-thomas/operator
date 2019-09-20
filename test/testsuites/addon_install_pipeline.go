package testsuites

import (
	"context"
	"testing"
	"time"

	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	"github.com/tektoncd/operator/pkg/controller/addon"
	"github.com/tektoncd/operator/pkg/controller/setup"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-sdk/pkg/test"
	testConfig "github.com/tektoncd/operator/test/config"
	"github.com/tektoncd/operator/test/helpers"
)

// ValidateAddonInstall creates an instance of addon.operator.tekton.dev
// and checks whether dashboard deployments are created
func ValidateAddonInstall(t *testing.T) {
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()

	checkPipeline(t, ctx)

	t.Run("creating-addon-with-version", addonCRWithVersion)
	t.Run("creating-addon-without-version", addonCRWithoutVersion)

	deletePipeline(t, ctx)

}

// ValidateAddonDeletion ensures that deleting the addon CR  deletes the already
// installed addon pipeline
func ValidateAddonDeletion(t *testing.T) {
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()


	t.Run("deleting-addon-cr", pipelineCRDeletion)
	t.Run("deleting-addon-cr", addonCRDeletion)
}

func installPipeline(t *testing.T, ctx *test.TestCtx) {
	configCR := &v1alpha1.Config{
		TypeMeta: v1.TypeMeta{
			Kind:       "config",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: setup.ClusterCRName,
		},
		Spec: v1alpha1.ConfigSpec{
			TargetNamespace: setup.DefaultTargetNs,
		},
	}

	cleanupOptions := &test.CleanupOptions{
		TestContext:   ctx,
		Timeout:       5 * time.Second,
		RetryInterval: 1 * time.Second,
	}
	err := test.Global.Client.Create(context.TODO(), configCR, cleanupOptions)
	helpers.AssertNoError(t, err)
	helpers.WaitForClusterCR(t, setup.ClusterCRName, configCR)
	helpers.ValidatePipelineSetup(t, configCR,
		setup.PipelineControllerName,
		setup.PipelineWebhookName)

}
func checkPipeline(t *testing.T, ctx *test.TestCtx) {
	configCR := &v1alpha1.Config{
		TypeMeta: v1.TypeMeta{
			Kind:       "config",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: setup.ClusterCRName,
		},
		Spec: v1alpha1.ConfigSpec{
			TargetNamespace: setup.DefaultTargetNs,
		},
	}

	helpers.WaitForClusterCR(t, setup.ClusterCRName, configCR)
	helpers.ValidatePipelineSetup(t, configCR,
		setup.PipelineControllerName,
		setup.PipelineWebhookName)

}

func deletePipeline(t *testing.T, ctx *test.TestCtx) {
	configCR := &v1alpha1.Config{}
	helpers.WaitForClusterCR(t, setup.ClusterCRName, configCR)
	helpers.DeleteClusterCR(t, setup.ClusterCRName)
	helpers.ValidatePipelineCleanup(t, configCR,
		setup.PipelineControllerName,
		setup.PipelineWebhookName)
}

func addonCRWithVersion(t *testing.T) {
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()

	addonCR := &v1alpha1.Addon{
		TypeMeta: v1.TypeMeta{
			Kind:       "Addon",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "dashboard",
		},
		Spec: v1alpha1.AddonSpec{
			Version: "v0.1.1",
		},
	}

	cleanupOpetions := &test.CleanupOptions{
		TestContext:   ctx,
		Timeout:       5 * time.Second,
		RetryInterval: 1 * time.Second,
	}

	err := test.Global.Client.Create(
		context.TODO(),
		addonCR,
		cleanupOpetions)

	helpers.AssertNoError(t, err)

	err = e2eutil.WaitForDeployment(
		t, test.Global.KubeClient, setup.DefaultTargetNs,
		"tekton-dashboard",
		1,
		testConfig.APIRetry,
		testConfig.APITimeout,
	)
	helpers.AssertNoError(t, err)

	helpers.WaitForClusterCR(t, "dashboard", addonCR)
	if code := addonCR.Status.Conditions[0].Code; code != v1alpha1.InstalledStatus {
		t.Errorf("Expected code to be %s but got %s", v1alpha1.InstalledStatus, code)
	}
}

func addonCRWithoutVersion(t *testing.T) {
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()

	addonCR := &v1alpha1.Addon{
		TypeMeta: v1.TypeMeta{
			Kind:       "Addon",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "dashboard",
		},
	}

	cleanupOpetions := &test.CleanupOptions{
		TestContext:   ctx,
		Timeout:       5 * time.Second,
		RetryInterval: 1 * time.Second,
	}

	err := test.Global.Client.Create(
		context.TODO(),
		addonCR,
		cleanupOpetions)

	helpers.AssertNoError(t, err)

	err = e2eutil.WaitForDeployment(
		t, test.Global.KubeClient, setup.DefaultTargetNs,
		"tekton-dashboard",
		1,
		testConfig.APIRetry,
		testConfig.APITimeout,
	)
	helpers.AssertNoError(t, err)

	helpers.WaitForClusterCR(t, "dashboard", addonCR)

	version, err := addon.GetLatestVersion(addonCR)
	if addonCR.Spec.Version != version {
		t.Errorf("Expected version to be %s but got %s", version, addonCR.Spec.Version)
	}

	if code := addonCR.Status.Conditions[0].Code; code != v1alpha1.InstalledStatus {
		t.Errorf("Expected code to be %s but got %s", v1alpha1.InstalledStatus, code)
	}
}

func addonCRDeletion(t *testing.T) {
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()
	installPipeline(t, ctx)

	addonCR := &v1alpha1.Addon{
		TypeMeta: v1.TypeMeta{
			Kind:       "Addon",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "dashboard",
		},
	}

	cleanupOpetions := &test.CleanupOptions{
		TestContext:   ctx,
		Timeout:       5 * time.Second,
		RetryInterval: 1 * time.Second,
	}

	err := test.Global.Client.Create(
		context.TODO(),
		addonCR,
		cleanupOpetions)

	helpers.AssertNoError(t, err)

	err = e2eutil.WaitForDeployment(
		t, test.Global.KubeClient, setup.DefaultTargetNs,
		"tekton-dashboard",
		1,
		testConfig.APIRetry,
		testConfig.APITimeout,
	)
	helpers.AssertNoError(t, err)

	helpers.DeleteAddonCR(t, "dashboard")
	helpers.WaitForDeploymentDeletion(t, setup.DefaultTargetNs, "tekton-dashboard")

	deletePipeline(t, ctx)
}

func pipelineCRDeletion(t *testing.T) {
	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()
	installPipeline(t, ctx)

	addonCR := &v1alpha1.Addon{
		TypeMeta: v1.TypeMeta{
			Kind:       "Addon",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "dashboard",
		},
	}

	cleanupOpetions := &test.CleanupOptions{
		TestContext:   ctx,
		Timeout:       5 * time.Second,
		RetryInterval: 1 * time.Second,
	}

	err := test.Global.Client.Create(
		context.TODO(),
		addonCR,
		cleanupOpetions)

	helpers.AssertNoError(t, err)

	err = e2eutil.WaitForDeployment(
		t, test.Global.KubeClient, setup.DefaultTargetNs,
		"tekton-dashboard",
		1,
		testConfig.APIRetry,
		testConfig.APITimeout,
	)
	helpers.AssertNoError(t, err)

	cr := &v1alpha1.Config{}
	helpers.WaitForClusterCR(t, setup.ClusterCRName, cr)

	deletePipeline(t, ctx)
	helpers.ValidatePipelineCleanup(t, cr,
		setup.PipelineControllerName,
		setup.PipelineWebhookName)

	helpers.WaitForDeploymentDeletion(t, setup.DefaultTargetNs, "tekton-dashboard")
}
