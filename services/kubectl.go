package services

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	log "github.com/sirupsen/logrus"
)

func BuildClient(ctx string) *kubernetes.Clientset {
	var path *string
	if home := homedir.HomeDir(); home != "" {
		path = flag.String("config", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		path = flag.String("config", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	switchContext(ctx, *path)

	config, err := clientcmd.BuildConfigFromFlags("", *path)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func switchContext(ctx, kubeconfigPath string) (err error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("error getting RawConfig: %w", err)
	}

	if config.Contexts[ctx] == nil {
		return fmt.Errorf("context %s doesn't exists", ctx)
	}

	config.CurrentContext = ctx
	err = clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), config, true)
	if err != nil {
		return fmt.Errorf("error ModifyConfig: %w", err)
	}

	log.Info("Switched to context \"%s\"", ctx)
	return nil
}
