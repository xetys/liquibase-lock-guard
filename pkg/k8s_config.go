package pkg

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

var cachedConfig *rest.Config

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func K8SConfig() (*rest.Config, error) {
	if cachedConfig != nil {
		return cachedConfig, nil
	}

	var kubeconfig *string

	config, err := rest.InClusterConfig()

	if err != nil {
		klog.Infoln("in cluster config failed, trying from local")
		kubeconfigEnvVar := os.Getenv("KUBECONFIG")
		if len(kubeconfigEnvVar) > 0 {
			klog.Infoln("found higher priority KUBECONFIG in %s\n", kubeconfigEnvVar)
			kubeconfig = flag.String("kubeconfig", kubeconfigEnvVar, "absolute path to the kubeconfig file")
		} else {
			if home := homeDir(); home != "" {
				kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
			} else {
				kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
			}
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	cachedConfig = config

	return config, nil
}

func ClientSet() (*kubernetes.Clientset, error) {

	// creates the in-cluster config
	config, err := K8SConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, err
	}

	return clientset, nil
}
