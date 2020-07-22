package configmaps

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var state testSetup

func TestMain(m *testing.M) {
	var err error
	state, err = setup()
	if err != nil {
		os.Exit(1)
	}
	code := m.Run()
	if err = teardown(state); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}

func TestReactsOnCreate(t *testing.T) {
	cm := &corev1.ConfigMap{}
	cm.ObjectMeta.Name = "test"
	cm.Data = make(map[string]string)
	cm.Data["created1.txt"] = "data1"
	cm.Data["created2.txt"] = "data2"

	_, ctrl, err := testWith(cm)
	if err != nil {
		t.Fatalf("Failed to setup up the test. %s", err)
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cm.ObjectMeta.Name,
			Namespace: cm.ObjectMeta.Namespace,
		},
	}
	ctrl.Reconcile(req)

	files, err := ioutil.ReadDir(state.workDir)
	if err != nil {
		t.Fatalf("Failed to read the sync dir. %s", err)
	}

	if len(files) != 2 {
		t.Fatalf("There should have been exactly 2 files in the sync dir but there were %d.", len(files))
	}

	foundCreated1 := false
	foundCreated2 := false

	for _, f := range files {
		if f.Name() == "created1.txt" {
			contents, err := ioutil.ReadFile(filepath.Join(state.workDir, f.Name()))
			if err == nil && string(contents) == "data1" {
				foundCreated1 = true
			}
		} else if f.Name() == "created2.txt" {
			contents, err := ioutil.ReadFile(filepath.Join(state.workDir, f.Name()))
			if err == nil && string(contents) == "data2" {
				foundCreated2 = true
			}
		}
	}

	if !foundCreated1 {
		t.Error("Failed to find the expected created1.txt with matching contents.")
	}

	if !foundCreated2 {
		t.Error("Failed to find the expected created2.txt with matching contents.")
	}
}

func TestReactsOnUpdate(t *testing.T) {
}

func TestReactsOnDelete(t *testing.T) {
}

func TestReactsOnLabelChange(t *testing.T) {
}

func testWith(cms ...runtime.Object) (client.Client, reconcile.Reconciler, error) {
	cl := fake.NewFakeClient(cms...)

	cfg := rest.Config{}

	opts := manager.Options{
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
			return cl, nil
		},
		MapperProvider: func(c *rest.Config) (meta.RESTMapper, error) {
			gv := schema.GroupVersion{
				Group:   "",
				Version: "v1",
			}

			gvs := make([]schema.GroupVersion, 1)
			gvs[0] = gv
			mapper := meta.NewDefaultRESTMapper(gvs)
			mapper.Add(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap",
			}, meta.RESTScopeRoot)

			return mapper, nil
		},
		NewCache: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
			return cache.Cache(&informertest.FakeInformers{}), nil
		},
	}
	mgr, err := manager.New(&cfg, opts)

	if err != nil {
		return nil, nil, err
	}

	ctrl, err := New(mgr, ConfigMapReconcilerConfig{
		BaseDir: state.workDir,
		NewClient: func(*rest.Config) (client.Client, error) {
			return cl, nil
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return cl, ctrl, nil
}

func setup() (testSetup, error) {
	workDir, err := ioutil.TempDir("", "config-bump-test")
	if err != nil {
		return testSetup{}, err
	}

	return testSetup{
		workDir: workDir,
	}, nil
}

func teardown(s testSetup) error {
	return os.RemoveAll(s.workDir)
}

type testSetup struct {
	workDir string
}
