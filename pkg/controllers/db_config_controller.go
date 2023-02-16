package controllers

import (
	"context"
	"fmt"
	v1 "github.com/myoperator/dbcore/pkg/apis/dbconfig/v1"
	"github.com/myoperator/dbcore/pkg/builders"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DbConfigController struct {
	client.Client
}

func NewDbConfigController() *DbConfigController {
	return &DbConfigController{}
}


func(r *DbConfigController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	config := &v1.DbConfig{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		return reconcile.Result{}, err
	}
	fmt.Println(config)

	// configmap 构建
	cmBuilder, err := builders.NewConfigMapBuilder(config, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	cm, err := cmBuilder.Build(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	// deployment 构建
	depBuilder, err := builders.NewDeploymentBuilder(config, r.Client, cmBuilder)
	if err != nil {
		return reconcile.Result{}, err
	}

	dep, err := depBuilder.Build(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}



	fmt.Println(dep)
	fmt.Println(cm)

	return reconcile.Result{}, err
}

// InjectClient 必须使用inject
func (r *DbConfigController) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

// OnDelete
func (r *DbConfigController) OnDelete(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface){
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == v1.KindForUse && ref.APIVersion == v1.ApiVersion {
			// 重新入列
			klog.Info("被删除的对象名称是 name: ", event.Object.GetName(),"kind: ",  event.Object.GetObjectKind())
			limitingInterface.Add(
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:ref.Name,
						Namespace: event.Object.GetNamespace(),
					},
				})
		}
	}
}

// OnUpdate
func (r *DbConfigController) OnUpdate(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface){
	for _, ref := range event.ObjectNew.GetOwnerReferences() {
		if ref.Kind == v1.KindForUse && ref.APIVersion == v1.ApiVersion {
			// 重新入列
			klog.Info("更新的新对象名称是 name: ", event.ObjectNew.GetName(),"kind: ",  event.ObjectNew.GetObjectKind())
			limitingInterface.Add(
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:ref.Name,
						Namespace: event.ObjectNew.GetNamespace(),
					},
				})
		}
	}
}
