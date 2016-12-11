package main

import (
	"flag"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/schema"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	ex_v1 "github.com/shmurata/tpr-sample/apis/example.com/v1"
)

var (
	config *rest.Config
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// initialize third party resource if it does not exist
	tpr, err := clientset.Extensions().ThirdPartyResources().Get("hello-world.example.com")
	if err != nil {
		if errors.IsNotFound(err) {
			tpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: v1.ObjectMeta{
					Name: "hello-world.example.com",
				},
				Versions: []v1beta1.APIVersion{
					{Name: "v1"},
				},
				Description: "Hello World Object",
			}

			result, err := clientset.Extensions().ThirdPartyResources().Create(tpr)
			if err != nil {
				panic(err)
			}
			fmt.Printf("CREATED: %#v\nFROM: %#v\n", result, tpr)
		} else {
			panic(err)
		}
	} else {
		fmt.Printf("SKIPPING: already exists %#v\n", tpr)
	}

	// make a new config for our extension's API group, using the first config as a baseline
	var tprconfig *rest.Config
	tprconfig = config
	configureClient(tprconfig)

	tprclient, err := rest.RESTClientFor(tprconfig)
	if err != nil {
		panic(err)
	}

	var hello ex_v1.HelloWorld

	err = tprclient.Get().
		Resource("helloworlds").
		Namespace(api.NamespaceDefault).
		Name("example1").
		Do().Into(&hello)

	if err != nil {
		if errors.IsNotFound(err) {
			// Create an instance of our TPR
			hello := &ex_v1.HelloWorld{
				Metadata: api.ObjectMeta{
					Name: "example1",
				},
				Spec: ex_v1.HelloWorldSpec{
					Foo: "hello",
					Bar: true,
				},
			}

			var result ex_v1.HelloWorld
			err = tprclient.Post().
				Resource("helloworlds").
				Namespace(api.NamespaceDefault).
				Body(hello).
				Do().Into(&result)

			if err != nil {
				panic(err)
			}
			fmt.Printf("CREATED: %#v\n", result)
		} else {
			panic(err)
		}
	} else {
		fmt.Printf("GET: %#v\n", hello)
	}

	// Fetch a list of our TPRs
	helloList := ex_v1.HelloWorldList{}
	err = tprclient.Get().Resource("helloworlds").Do().Into(&helloList)
	if err != nil {
		panic(err)
	}
	fmt.Printf("LIST: %#v\n", helloList)
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func configureClient(config *rest.Config) {
	groupversion := schema.GroupVersion{
		Group:   "example.com",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupversion,
				&ex_v1.HelloWorld{},
				&ex_v1.HelloWorldList{},
				&api.ListOptions{},
				&api.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(api.Scheme)
}
