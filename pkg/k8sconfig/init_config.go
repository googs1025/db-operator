package k8sconfig

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)



// K8sRestConfigInPod 集群内部署
func K8sRestConfigInPod() *rest.Config{
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	return config
}


// K8sRestConfig 获取config对象
func K8sRestConfig() *rest.Config {
	//自定义环境
	if os.Getenv("release") == "1" {
		log.Println("run in cluster")
		return K8sRestConfigInPod()
	}
	log.Println("run outside cluster")
	// 集群外部署
	config, err := clientcmd.BuildConfigFromFlags("","./resources/config" )
	config.Insecure = true
	if err != nil {
	   log.Fatal(err)
	}

	return config
}
