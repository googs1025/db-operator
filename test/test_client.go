package test

import (
	"context"
	"fmt"
	clientv1 "github.com/myoperator/dbcore/pkg/client/clientset/versioned/typed/dbconfig/v1"
	"github.com/myoperator/dbcore/pkg/k8sconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

// 练习调用code-generator生成出来的client
func main() {
	k8sConfig := k8sconfig.K8sRestConfig()

	client, err := clientv1.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatal(err)
	}
	dcList, _ := client.DbConfigs("default").List(context.Background(), metav1.ListOptions{})
	fmt.Println(dcList)
}