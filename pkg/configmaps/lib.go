package configmaps

import (
	"context"
	"crypto/md5"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/operator-framework/operator-sdk/pkg/predicate"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("configmaps")

// ConfigMapReconcilerConfig is the configuration of the reconciler
type ConfigMapReconcilerConfig struct {
	BaseDir         string
	Labels          string
	Namespace       string
	OnReconcileDone func() error
}

type configMapReconciler struct {
	client    client.Client
	config    ConfigMapReconcilerConfig
	selector  labels.Selector
	baseDir   string
	namespace string
	fileCache map[types.NamespacedName]configFiles
}

// configFiles is a map where keys are the names of the files and values are digests of their content
type configFiles = map[string][16]byte

// New creates a config map reconciler with given configuration and configures a controller for it
func New(mgr manager.Manager, config ConfigMapReconcilerConfig) error {
	lbls, err := labels.ConvertSelectorToLabelsMap(config.Labels)
	if err != nil {
		return err
	}

	r := &configMapReconciler{
		client:    mgr.GetClient(),
		config:    config,
		selector:  lbls.AsSelector(),
		fileCache: make(map[types.NamespacedName]configFiles),
	}

	// register the controller with the manager
	bld := builder.ControllerManagedBy(mgr)
	bld.Named("config-bump")
	bld.WithEventFilter(predicate.ResourceFilterPredicate{Selector: r.selector})
	if err = bld.Complete(r); err != nil {
		return err
	}

	r.Initialize()

	return nil
}

// Initialize performs an initial sync of the local set of files with the configured config maps
// It also initializes the watches to initialize the reconciliation loop
func (c *configMapReconciler) Initialize() error {
	list := &corev1.ConfigMapList{}
	opts := []client.ListOption{
		client.InNamespace(c.config.Namespace),
		client.MatchingLabelsSelector{Selector: c.selector},
	}

	if err := c.client.List(context.TODO(), list, opts...); err != nil {
		return err
	}

	c.fileCache = make(map[types.NamespacedName]configFiles)
	processedFiles := make([]string, 0, 8)

	for _, cm := range list.Items {
		for name, data := range cm.Data {
			path := filepath.Join(c.config.BaseDir, name)
			doWrite := false
			if _, err := os.Stat(path); err == nil || os.IsExist(err) {
				// if the file exists
				if content, err := ioutil.ReadFile(path); err != nil {
					log.Error(err, "Failed to open the config file to see if it changed", "file", path)
				} else {
					dataHash := md5.Sum([]byte(data))
					contentHash := md5.Sum([]byte(content))

					doWrite = dataHash != contentHash
				}
			} else {
				// the file doesn't exist
				doWrite = true
			}

			if doWrite {
				if f, err := os.Create(path); err != nil {
					log.Error(err, "Failed to create a file for the configmap",
						"file", path, "namespace", cm.GetObjectMeta().GetNamespace(), "name", cm.GetObjectMeta().GetName())
				} else {
					defer f.Close()
					f.Write([]byte(data))
				}
			}

			processedFiles = append(processedFiles, path)
		}
	}

	// now go through all the existing files and delete those we have not processed while reading the config maps
	files, err := ioutil.ReadDir(c.config.BaseDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			path := filepath.Join(c.config.BaseDir, f.Name())
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}

	return nil
}

// Reconcile handles the changes in the configured config maps
func (c *configMapReconciler) Reconcile(r reconcile.Request) (reconcile.Result, error) {
	cm := &corev1.ConfigMap{}
	err := c.client.Get(context.TODO(), r.NamespacedName, cm)
	if err != nil && !kerrors.IsNotFound(err) {
		return reconcile.Result{}, errors.New("error while retrieving object")
	}

	// map of file names and their content as found in the config map
	toCreate := make(map[string][]byte)
	toUpdate := make(map[string][]byte)
	toDelete := make([]string, 0, 8)

	cachedFiles, cached := c.fileCache[r.NamespacedName]
	if err == nil {
		if cached {
			for name, data := range cm.Data {
				if existingHash, ok := cachedFiles[name]; ok {
					// we already know about this file - let's see if it changed
					hash := md5.Sum([]byte(data))

					if hash != existingHash {
						toUpdate[name] = []byte(data)
					}
				} else {
					// our cache doesn't contain this file...
					toCreate[name] = []byte(data)
				}
			}
		} else {
			for name, data := range cm.Data {
				toCreate[name] = []byte(data)
			}
		}
	} else {
		// not found - one of the config maps has been deleted
		if cached {
			for name := range cachedFiles {
				toDelete = append(toDelete, name)
			}
		} else {
			// we got report about a deletion of a config map that was not in our cache.
			// This calls for a full scan
			if err := c.Initialize(); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	for name, data := range toCreate {
		f, err := os.Create(filepath.Join(c.config.BaseDir, name))
		if err != nil {
			return reconcile.Result{}, err
		}

		defer f.Close()

		if _, err = f.Write(data); err != nil {
			return reconcile.Result{}, err
		}
	}

	for name, data := range toUpdate {
		f, err := os.Create(filepath.Join(c.config.BaseDir, name))
		if err != nil {
			return reconcile.Result{}, err
		}

		defer f.Close()

		if _, err = f.Write(data); err != nil {
			return reconcile.Result{}, err
		}
	}

	for _, name := range toDelete {
		err := os.Remove(filepath.Join(c.config.BaseDir, name))
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
