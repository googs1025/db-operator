package builders

import (
	"bytes"
	"context"
	v12 "github.com/myoperator/dbcore/pkg/apis/dbconfig/v1"
	"html/template"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"fmt"
)

type DeploymentBuilder struct {
	Dep 	    *appsv1.Deployment
	CmBuilder 	*ConfigMapBuilder  // 关联对象
	Client      client.Client
	YamlConfig  *v12.DbConfig
}

func name(name string) string {
	return "dbcore-" + name
}

func NewDeploymentBuilder(config *v12.DbConfig, client client.Client, cmBuilder *ConfigMapBuilder) (*DeploymentBuilder, error) {

	dep := &appsv1.Deployment{}
	dep.Name = config.Name
	dep.Namespace = config.Namespace

	key := types.NamespacedName{
		Name: name(config.Name), // 模版内的name 不一样
		Namespace: config.Namespace,
	}

	err := client.Get(context.Background(), key, dep)
	// 如果没取到，需要创建
	if err != nil {

		tem, err := template.New("deployment").Parse(DeploymentTemplate)
		if err != nil {
			klog.Errorf("template parse error:", err)
			return nil, err
		}
		var t bytes.Buffer
		err = tem.Execute(&t, dep)
		if err != nil {
			klog.Errorf("template execute error:", err)
			return nil, err
		}

		err = yaml.Unmarshal(t.Bytes(), dep)
		if err != nil {
			klog.Errorf("deployment yaml unmarshal error:", err)
			return nil, err
		}

	}



	return &DeploymentBuilder{Dep: dep, Client: client, YamlConfig: config, CmBuilder: cmBuilder}, nil
}


func (d *DeploymentBuilder) SetReplica(replica int) *DeploymentBuilder {
	*d.Dep.Spec.Replicas = int32(replica)
	return d
}

// 更新deployment中的字段
func (d *DeploymentBuilder) apply() *DeploymentBuilder {
	*d.Dep.Spec.Replicas = int32(d.YamlConfig.Spec.Replicas)

	return d
}

// 需要级连删除
func(d *DeploymentBuilder) setOwner() *DeploymentBuilder{
	d.Dep.OwnerReferences = append(d.Dep.OwnerReferences,
		v1.OwnerReference {
			APIVersion: d.YamlConfig.APIVersion,
			Kind: d.YamlConfig.Kind,
			Name: d.YamlConfig.Name,
			UID: d.YamlConfig.UID,
		})
	return d
}

// 配合滚动deployment使用
const CMAnnotation = "dbcore.config/md5"

func(d *DeploymentBuilder) setCMAnnotation(configStr string) {
	// 需要放在Template中的Annotations字段，因为才能成功触发pod的滚动更新
	d.Dep.Spec.Template.Annotations[CMAnnotation] = configStr
}

// Build 构建出deployment对象
func (d *DeploymentBuilder) Build(ctx context.Context) (*appsv1.Deployment, error) {

	// 创建
	if d.Dep.CreationTimestamp.IsZero() {
		d.apply().setOwner() // 更新deployment对象的字段，ex: replicas，且创建时需要设置OwnerReferences

		// 设置 config md5
		d.setCMAnnotation(d.CmBuilder.DataKey)

		err := d.Client.Create(ctx, d.Dep)
		if err != nil {
			klog.Error("create deployment err: ", err)
			return nil, err
		}
	// 更新
	} else {

		// 更新:法一 update方式
		d.apply() // 更新deployment对象的字段，ex: replicas
		// 设置 config md5
		d.setCMAnnotation(d.CmBuilder.DataKey)
		err := d.Client.Update(ctx, d.Dep)
		if err != nil {
			klog.Error("update deployment err: ", err)
			return nil, err
		}
		// 更新: 法二 patch方式
		//patch := client.MergeFrom(d.Dep.DeepCopy()) // 需要先拷贝一份旧的。
		//d.apply() // 更新deployment对象的字段，ex: replicas
		//err = d.Client.Patch(ctx, d.Dep, patch) // 使用patch方法
		//if err != nil {
		//	return nil, err
		//}

		// deployment 查看状态
		replicas := d.Dep.Status.ReadyReplicas // 获取当前deployment的ready状态副本数
		d.YamlConfig.Status.Ready = fmt.Sprintf("%d/%d", replicas, d.YamlConfig.Spec.Replicas)
		d.YamlConfig.Status.Replicas = replicas
		err = d.Client.Status().Update(ctx, d.YamlConfig)
		if err != nil {
			klog.Error("update deployment status err: ", err)
			return nil, err
		}

	}
	return d.Dep, nil
}


