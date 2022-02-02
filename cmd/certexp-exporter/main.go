package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Secret struct {
	Name           string `json:"name"`
	ServiceAccount string `json:"serviceaccount"`
	Token          string `json:"token"`
}

type ListOfSecrets struct {
	Namespace string `json:"namespace"`
	Secrets   []Secret
}

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "Absolute Path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "Absolute path to kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting Secrets Exporter...")
	 
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c := make(chan string)
		go Data(clientset, c)
		fmt.Fprintf(w, <-c)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Data(clientset *kubernetes.Clientset, c chan string)  {	
	values := GetSecrets(clientset)
	a := []ListOfSecrets{}
	for range GetNamespaces(clientset) {
		a = append(a, <-values)
	}

	json, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	c <- string(json)
}

func GetNamespaces(clientset *kubernetes.Clientset) []string {

	namespace, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{LabelSelector: "scope=client"})

	if err != nil {
		log.Fatal(err)
	}

	var namespaces []string

	for _, i := range namespace.Items {
		namespaces = append(namespaces, i.Name)
	}
	return namespaces
}

func GetSecrets(clientset *kubernetes.Clientset) <-chan ListOfSecrets {

	c := make(chan ListOfSecrets)
	var sec []Secret

	for _, namespace := range GetNamespaces(clientset) {

		go func(namespace string) {

			secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})

			if err != nil {
				log.Fatal(err)
			}
			for _, secret := range secrets.Items {
				if secret.Type == "kubernetes.io/service-account-token" {
					sec = append(sec, Secret{
						Name:           secret.Name,
						ServiceAccount: secret.Annotations["kubernetes.io/service-account.name"],
						Token:          string(secret.Data["token"]),
					})
				}
			}

			c <- ListOfSecrets{
				Namespace: namespace,
				Secrets:   sec,
			}

		}(namespace)
	}
	return c
}
