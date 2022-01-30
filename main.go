package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "Absolute Path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "Absolute path to kubeconfig file")
	}

	flag.Parse()
	// fmt.Println(*kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	secret, err := clientset.CoreV1().Secrets("namespace").List(context.TODO(), metav1.ListOptions{})
	secrets := ParseSecrets(secret.Items)
	for range secret.Items{
		fmt.Println(<-secrets)
	}

}

func ParseSecrets(secrets []apiv1.Secret) <- chan string {
	c := make(chan string)

	for _, secret := range secrets{
		go func (secret apiv1.Secret)  {
			c <- string(secret.Data["token"])
		}(secret)
	}
	return c
}