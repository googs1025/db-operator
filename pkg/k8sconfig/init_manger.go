package k8sconfig

import (
	v1 "github.com/myoperator/dbcore/pkg/apis/dbconfig/v1"
	"github.com/myoperator/dbcore/pkg/controllers"
	appsv1 "k8s.io/api/apps/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/source"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// InitManager 初始化管理器
func InitManager() {

	logf.SetLogger(zap.New())
	// 1. 初始化管理器
	mgr, err := manager.New(K8sRestConfig(), manager.Options{
		Logger:  logf.Log.WithName("dbcore"),
	})
	if err != nil {
		mgr.GetLogger().Error(err, "unable to set up manager")
		os.Exit(1)
	}

	// 2. 注册Scheme
	if err = v1.SchemeBuilder.AddToScheme(mgr.GetScheme());err!=nil{
		mgr.GetLogger().Error(err, "unable add scheme")
		os.Exit(1)
	}

	// 3. 控制器相关
	dbConfigController := controllers.NewDbConfigController()
	if err = builder.ControllerManagedBy(mgr).
		For(&v1.DbConfig{}). // 监听主资源
		// 监听子资源，为了在误删除子资源时，会重新创建。
		Watches(&source.Kind{Type: &appsv1.Deployment{}},
			handler.Funcs{
				DeleteFunc: dbConfigController.OnDelete,
				UpdateFunc: dbConfigController.OnUpdate,
			},
		).
		// 这里的传入对象，需要实现Reconcile。
		Complete(controllers.NewDbConfigController());err != nil {
		mgr.GetLogger().Error(err, "unable to create manager")
		os.Exit(1)
	}
	// 4. 启动管理器
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		mgr.GetLogger().Error(err, "unable to start manager")
		os.Exit(1)
	}

}
