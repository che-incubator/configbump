package main

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/che-incubator/configbump/pkg/bumper"
)

// Opts represents the commandline arguments to the executable
type opts struct {
	Dir                  string `arg:"-d,required,env:CONFIG_BUMP_DIR" help:"The directory to which persist the files retrieved from config maps. Can also be specified using env var: CONFIG_BUMP_DIR"`
	Namespace            string `arg:"-n,required,env:CONFIG_BUMP_NAMESPACE" help:"The namespace in which to look for the config maps to persist. Can also be specified using env var: CONFIG_BUMP_NAMESPACE"`
	TLSVerify            bool   `arg:"--tls-verify,-t,env:CONFIG_BUMP_TLS_VERIFY" default:"true" help:"Whether to require valid certificate chain. Can also be specified using env var: CONFIG_BUMP_TLS_VERIFY"`
	Labels               string `arg:"-l,env:CONFIG_BUMP_LABELS" help:"An expression to match the labels against. Consult the Kubernetes documentation for the syntax required. Can also be specified using env var: CONFIG_BUMP_LABELS"`
	ProcessCommand       string `arg:"--process-command,-c,env:CONFIG_BUMP_PROCESS_COMMAND" help:"The commandline by which to identify the process to send the signal to. This can be a regular expression. Ignored if process pid is specified. Can also be specified using env var: CONFIG_BUMP_PROCESS_COMMAND"`
	ProcessPid           int32  `arg:"--process-pid,-p,env:CONFIG_BUMP_PROCESS_PID" help:"The PID of the process to send the signal to, if known. Otherwise process detection can be used. Can also be specified using env var: CONFIG_BUMP_PROCESS_PID"`
	ProcessParentCommand string `arg:"--process-parent-command,-a,env:CONFIG_BUMP_PARENT_PROCESS_COMMAND" help:"The commandline by which to identify the parent process of the process to send signal to. This can be a regular expression. Ignored if parent process pid is specified. Can also be specified using env var: CONFIG_BUMP_PARENT_PROCESS_COMMAND"`
	ProcessParentPid     int32  `arg:"--process-parent-pid,-i,env:CONFIG_BUMP_PARENT_PROCESS_PID" help:"The PID of the parent process of the process to send the signal to, if known. Otherwise process detection can be used. Can also be specified using env var: CONFIG_BUMP_PARENT_PROCESS_PID"`
	Signal               string `arg:"-s,env:CONFIG_BUMP_SIGNAL" help:"The name of the signal to send to the process on the configuration files change. Use 'kill -l' to get a list of possible signals. Can also be specified using env var: CONFIG_BUMP_SIGNAL"`
}

// Version returns the version of the program
func (opts) Version() string {
	return "config-bump 0.1.0"
}

func main() {
	var opts opts
	arg.MustParse(&opts)

	fmt.Printf("dir: %s", opts.Dir)
	d, err := bumper.DetectCommand("ahoj")
	if err != nil {

	}
	var ds []bumper.Detection = []bumper.Detection{d}

	// TODO start the reconciler and call the below once the reconciler
	// persisted all the config map changes to the filesystem
	b := bumper.New(opts.Signal, ds)
	if err := b.Bump(); err != nil {
		//nazdar
	}
}
