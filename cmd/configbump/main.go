package main

import (
	"context"
	"os"

	arg "github.com/alexflint/go-arg"
	"github.com/che-incubator/configbump/pkg/configmaps"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	// controller-runtime imports
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Opts represents the commandline arguments to the executable
type opts struct {
	Dir string `arg:"-d,required,env:CONFIG_BUMP_DIR" help:"The directory to which persist the files retrieved from config maps. Can also be specified using env var: CONFIG_BUMP_DIR"`

	// TODO not supported yet
	//TLSVerify bool `arg:"--tls-verify,-t,env:CONFIG_BUMP_TLS_VERIFY" default:"true" help:"Whether to require valid certificate chain. Can also be specified using env var: CONFIG_BUMP_TLS_VERIFY"`

	Labels    string `arg:"-l,required,env:CONFIG_BUMP_LABELS" help:"An expression to match the labels against. Consult the Kubernetes documentation for the syntax required. Can also be specified using env var: CONFIG_BUMP_LABELS"`
	Namespace string `arg:"-n,env:CONFIG_BUMP_NAMESPACE" help:"The namespace in which to look for the config maps to persist. Can also be specified using env var: CONFIG_BUMP_NAMESPACE. If not specified, it is autodetected."`

	// TODO the whole process bumping not implemented yet.
	//ProcessCommand       string `arg:"--process-command,-c,env:CONFIG_BUMP_PROCESS_COMMAND" help:"The commandline by which to identify the process to send the signal to. This can be a regular expression. Ignored if process pid is specified. Can also be specified using env var: CONFIG_BUMP_PROCESS_COMMAND"`
	//ProcessPid           int32  `arg:"--process-pid,-p,env:CONFIG_BUMP_PROCESS_PID" help:"The PID of the process to send the signal to, if known. Otherwise process detection can be used. Can also be specified using env var: CONFIG_BUMP_PROCESS_PID"`
	//ProcessParentCommand string `arg:"--process-parent-command,-a,env:CONFIG_BUMP_PARENT_PROCESS_COMMAND" help:"The commandline by which to identify the parent process of the process to send signal to. This can be a regular expression. Ignored if parent process pid is specified. Can also be specified using env var: CONFIG_BUMP_PARENT_PROCESS_COMMAND"`
	//ProcessParentPid     int32  `arg:"--process-parent-pid,-i,env:CONFIG_BUMP_PARENT_PROCESS_PID" help:"The PID of the parent process of the process to send the signal to, if known. Otherwise process detection can be used. Can also be specified using env var: CONFIG_BUMP_PARENT_PROCESS_PID"`
	//Signal               string `arg:"-s,env:CONFIG_BUMP_SIGNAL" help:"The name of the signal to send to the process on the configuration files change. Use 'kill -l' to get a list of possible signals. Can also be specified using env var: CONFIG_BUMP_SIGNAL"`
}

// Version returns the version of the program
func (opts) Version() string {
	return "config-bump 7.117.0-next"
}

const controllerName = "config-bump"

var log = logf.Log.WithName(controllerName)

func main() {
	// Setup signal handler to gracefully shut down the manager
	ctx := signals.SetupSignalHandler()

	zap, err := zap.NewProduction()
	if err != nil {
		println("Failed to initialize a zap logger.")
		os.Exit(1)
	}
	logf.SetLogger(zapr.NewLogger(zap))
	var opts opts
	arg.MustParse(&opts)

	// process signalling not implemented yet.
	// d, err := bumper.DetectCommand(opts.ProcessCommand)
	// if err != nil {
	// }
	// var ds []bumper.Detection = []bumper.Detection{d}
	// b := bumper.New(opts.Signal, ds)

	// once process signalling is implemented, we can call:
	// initializeConfigMapController(opts.Labels, opts.Dir, b.Bump)
	if err := initializeConfigMapController(ctx, opts.Labels, opts.Dir, opts.Namespace, func() error { return nil }); err != nil {
		log.Error(err, "Could not initialize the config map sync controller")
		os.Exit(1)
	}
}

func initializeConfigMapController(ctx context.Context, labels string, baseDir string, namespace string, onReconcileDone func() error) error {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Could not get config from API server")
		return err
	}

	if namespace == "" {
		log.Error(err, "Namespace was not provided via commandline arguments")
		return err
	}

	mgr, err := manager.New(cfg, manager.Options{
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{namespace: cache.Config{}},
		},
	})
	if err != nil {
		log.Error(err, "Could not create a manager for creating controllers")
		return err
	}

	configmapReconciler, err := configmaps.New(mgr, configmaps.ConfigMapReconcilerConfig{
		BaseDir:         baseDir,
		Labels:          labels,
		OnReconcileDone: onReconcileDone,
		Namespace:       namespace,
	})
	if err != nil {
		log.Error(err, "Could not initialize configmaps reconciler")
		return err
	}

	if err = configmapReconciler.SetupWithManager(mgr); err != nil {
		log.Error(err, "Could not setup the manager with configmaps reconciler")
		return err
	}
	return mgr.Start(ctx)
}
