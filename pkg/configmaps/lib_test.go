package configmaps

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
	cm.Data = make(map[string]string)
	cm.Data["created1.txt"] = "data1"
	cm.Data["created2.txt"] = "data2"

	state.client.Create(context.TODO(), cm)

	// TODO we should wait here for the sync to happen... Probably can use on OnReconcileDone function?

	files, err := ioutil.ReadDir(state.workDir)
	if err != nil {
		t.Fatalf("Failed to read the sync dir. %s", err)
	}

	if len(files) != 2 {
		t.Fatalf("There should have been exactly 2 files in the sync dir but there were %d.", len(files))
	}
}

func TestReactsOnUpdate(t *testing.T) {
}

func TestReactsOnDelete(t *testing.T) {
}

func TestReactsOnLabelChange(t *testing.T) {
}

func setup() (testSetup, error) {
	workDir, err := ioutil.TempDir("", "config-bump-test")
	if err != nil {
		return testSetup{}, err
	}

	cl := fake.NewFakeClient(&corev1.ConfigMap{})

	cfg := rest.Config{}

	opts := manager.Options{
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
			return cl, nil
		},
		MapperProvider: func(c *rest.Config) (meta.RESTMapper, error) {
			// TODO this fake rest mapper is probably the cause why we're not getting anything
			// in the tests...
			gv := schema.GroupVersion{
				Group:   "",
				Version: "v1",
			}

			gvs := make([]schema.GroupVersion, 1)
			gvs[0] = gv
			return meta.NewDefaultRESTMapper(gvs), nil
		},
	}
	manager, err := manager.New(&cfg, opts)

	if err != nil {
		return testSetup{}, err
	}

	New(manager, ConfigMapReconcilerConfig{
		BaseDir: workDir,
		NewClient: func(*rest.Config) (client.Client, error) {
			return cl, nil
		}})

	channel := make(chan struct{})

	go func() {
		manager.Start(channel)
	}()

	return testSetup{
		workDir: workDir,
		client:  cl,
		channel: channel,
	}, nil
}

func teardown(s testSetup) error {
	close(s.channel)
	return os.RemoveAll(s.workDir)
}

type testSetup struct {
	workDir string
	client  client.Client
	channel chan struct{}
}
