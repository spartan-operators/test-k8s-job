package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig string
	var imageName string
	var purge bool

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	}

	flag.BoolVar(&purge, "purge", false, "(optional), remove created jobs from cluster")
	flag.StringVar(&imageName, "image", "luksa/batch-job", "(optional) custom job image")

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	if purge {
		removeJob(clientSet)
	} else {
		spawnJob(clientSet, &imageName)
	}
}

func spawnJob(client *kubernetes.Clientset, imageName *string) {
	metadata := &metav1.ObjectMeta{
		Name:        "batch-job",
		Namespace:   "default",
		Labels: 	 map[string]string{},
		Annotations: map[string]string{},
	}

	podSpec := &apiv1.PodSpec{
		Containers: 	[]apiv1.Container {
			{
				Name:                     "batch-job",
				Image:                    *imageName,
				Command:                  []string{},
				Args:                     []string{},
			},
		},
		RestartPolicy:	apiv1.RestartPolicyOnFailure,
	}

	podTemplate := &apiv1.PodTemplateSpec{
		ObjectMeta: *metadata,
		Spec:       *podSpec,
	}

	jobSpec := &batchv1.JobSpec{
		Template:  *podTemplate,
	}

	job := &batchv1.Job{
		ObjectMeta: *metadata,
		Spec:       *jobSpec,
	}

	if res, err := client.BatchV1().Jobs("default").Create(context.TODO(), job, metav1.CreateOptions{}); err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("API Job Response: %v\n", res)
	}
}

func removeJob(client *kubernetes.Clientset) {
	if err := client.BatchV1().Jobs("default").Delete(context.TODO(), "batch-job", metav1.DeleteOptions{}); err != nil {
		panic(err.Error())
	}
}
