package main

import (
	"flag"
	"fmt"
	exv1beta "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)
type ing struct {
	ingress v1beta1.IngressInterface
}
func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//var i ing
	//i.ingress = clientset.ExtensionsV1beta1().Ingresses(apiv1.NamespaceDefault)
	//i.create_ingress()
	//i.delete_ingress()
	//i.list_ingress()
	//i.watch_ingress()
	//i.update_ingress()


	//创建ConfigMap
	//configMap, err := clientset.CoreV1().ConfigMaps("ingress-nginx").Create(&apiv1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "tcp-nginx", Namespace: "ingress-nginx"}, Data: map[string]string{"30001": "default/nginx:98"}})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("configMap %s is created successful", configMap.Name)
	err = clientset.CoreV1().ConfigMaps("ingress-nginx").Delete("tcp-nginx", &metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("configMap %s is delete successful", "")

}
func (i *ing)create_ingress()  {
	ingress_yaml := &exv1beta.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx",
		},
		Spec: exv1beta.IngressSpec{
			Rules: []exv1beta.IngressRule{
				exv1beta.IngressRule{
					Host: "nginx.k8s.local",
					IngressRuleValue: exv1beta.IngressRuleValue{
						HTTP: &exv1beta.HTTPIngressRuleValue{
							Paths: []exv1beta.HTTPIngressPath{
								exv1beta.HTTPIngressPath{
									Backend: exv1beta.IngressBackend{
										ServiceName: "nginx",
										ServicePort: intstr.IntOrString{
											Type: intstr.Int,
											IntVal: 98,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ingress, err := i.ingress.Create(ingress_yaml)
	if err != nil {
		panic(err)
	}
	fmt.Printf("ingress %s is created successful", ingress.Name)
}

func (i *ing) delete_ingress ()  {
	err := i.ingress.Delete("nginx",&metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}else {
		fmt.Printf("the ingress delete successful")
	}
}

func (i *ing) list_ingress()  {
	ingress_list, err := i.ingress.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, i1 := range ingress_list.Items {
		fmt.Println(i1.Name)
	}
}

func (i *ing) watch_ingress ()  {
	watch_ingress, err := i.ingress.Watch(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	select {
	case e := <-watch_ingress.ResultChan():
		fmt.Println(e.Type)
	}
}

func (i *ing) update_ingress()  {
	ingress_yaml, err := i.ingress.Get("nginx",metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	ingress_yaml.Spec.Rules = []exv1beta.IngressRule{
		exv1beta.IngressRule{
			Host: "nginx-example.local.com",
		},
	}

	ingress, err := i.ingress.Update(ingress_yaml)
	if err != nil {
		panic(err)
	}
	fmt.Printf("the %s update successful",ingress.Name)
}
