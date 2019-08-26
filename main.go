package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
)

func main() {
	var (
		serviceYaml []byte
		serviceJson []byte
	)
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// ================1.create namespace================
	namespace, err := clientset.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hyperchain",
		},
	})
	fmt.Println(namespace,err)

	// ================2.create deployment================
	deploymentsClient := clientset.AppsV1().Deployments("hyperchain")
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hyperchain-demo",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "hyperchain-demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "hyperchain-demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "hyperchain-demo",
							Image: "nginx:1.7.9",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	// ================3.create service================
	services := clientset.CoreV1().Services("hyperchain")
	service := v1.Service{}
	// 读取YAML
	if serviceYaml, err = ioutil.ReadFile("./nginx-app-service.yaml"); err != nil {
		fmt.Println(err)
	}

	// YAML转JSON
	if serviceJson, err = yaml2.ToJSON(serviceYaml); err != nil {
		fmt.Println(err)
	}

	// JSON转struct
	if err = json.Unmarshal(serviceJson, &service); err != nil {
		fmt.Println(err)
	}
	create, err := services.Create(&service)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(create)

	// ================4.create ingress proxy================
	configMaps := clientset.CoreV1().ConfigMaps("ingress-nginx")
	configMap, err := configMaps.Get("tcp-services", metav1.GetOptions{})
	configMap.Data["30004"] = "hyperchain/hyperchain-demo-service:98"

	update, err := configMaps.Update(configMap)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(update)

	fmt.Println("================5.rollback all resource================")
	prompt()
	// ================5.rollback all resource================
	fmt.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete("hyperchain-demo", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
	prompt()
	fmt.Println("Deleting service...")
	if err := services.Delete("hyperchain-demo-service", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted service...")
	prompt()
	fmt.Println("rollback configMap...")
	update.Data["30004"] = ""
	rollback, err := configMaps.Update(configMap)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("rollback: ",rollback)



	//for {
	//	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	for  p,v :=range pods.Items  {
	//		fmt.Printf("=====================pod %v====================\n",p+1)
	//		fmt.Println(v)
	//	}
	//	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	//
	//	// Examples for error handling:
	//	// - Use helper functions like e.g. errors.IsNotFound()
	//	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	//	namespace := "default"
	//	pod := "tomcat-demo-6bc7d5b6f4-p27g6"
	//	_, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
	//	if errors.IsNotFound(err) {
	//		fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	//	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	//		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
	//			pod, namespace, statusError.ErrStatus.Message)
	//	} else if err != nil {
	//		panic(err.Error())
	//	} else {
	//		fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	//	}
	//
	//	time.Sleep(10 * time.Second)
	//}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
func int32Ptr(i int32) *int32 { return &i }

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}


