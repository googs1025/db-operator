package builders

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	configv1 "github.com/myoperator/dbcore/pkg/apis/dbconfig/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

type ConfigMapBuilder struct {
	CM  		*corev1.ConfigMap
	YamlConfig  *configv1.DbConfig
	Client  	client.Client
	DataKey string // 把cm中的app.yml的数据  进行md5更新
}

func NewConfigMapBuilder(yamlConfig *configv1.DbConfig, client client.Client) (*ConfigMapBuilder, error) {

	cm := &corev1.ConfigMap{}

	err := client.Get(context.Background(), types.NamespacedName{
		Name: name(yamlConfig.Name),
		Namespace: yamlConfig.Namespace,
	}, cm)
	if err != nil {

		cm.Name = name(yamlConfig.Name)
		cm.Namespace = yamlConfig.Namespace
		cm.Data = make(map[string]string)

		// FIXME: 不再使用，模版由configmap配置文件改为只有Data字段，所以下列代码废除
		//cm.Name = yamlConfig.Name
		//cm.Namespace = yamlConfig.Namespace
		//tpl, err := template.New("configmap").Parse(ConfigmapTemplate)
		//if err != nil {
		//	klog.Errorf("template parse error:", err)
		//	return nil, err
		//}
		//
		//var t bytes.Buffer
		//err = tpl.Execute(&t, cm)
		//if err != nil {
		//	klog.Errorf("template execute error:", err)
		//	return nil, err
		//}
		//err = yaml.Unmarshal(t.Bytes(), cm)
		//if err != nil {
		//	klog.Errorf("configmap yaml unmarshal error:", err)
		//	return nil, err
		//}

	}

	return &ConfigMapBuilder{YamlConfig: yamlConfig, Client: client, CM: cm}, nil
}

// 需要级连删除
func(c *ConfigMapBuilder) setOwner() *ConfigMapBuilder{
	c.CM.OwnerReferences = append(c.CM.OwnerReferences,
		v1.OwnerReference {
			APIVersion: c.YamlConfig.APIVersion,
			Kind: c.YamlConfig.Kind,
			Name: c.YamlConfig.Name,
			UID: c.YamlConfig.UID,
		})
	return c
}

const configMapKey = "app.yml"

// 更新与渲染configmap中的字段
func (c *ConfigMapBuilder) apply() *ConfigMapBuilder {

	tpl, err := template.New("configmap").Parse(ConfigmapTemplate)
	if err != nil {
		klog.Errorf("template parse error:", err)
		return c
	}

	var t bytes.Buffer
	err = tpl.Execute(&t, c.YamlConfig.Spec)
	if err != nil {
		klog.Errorf("template execute error:", err)
		return c
	}

	c.CM.Data[configMapKey] = t.String()

	return c
}

// parseKey 将configmap里面的 key=app.yml的内容 取出变成md5
func(c *ConfigMapBuilder) parseKey() *ConfigMapBuilder{

	if appData, ok := c.CM.Data[configMapKey]; ok{
		c.DataKey = Md5(appData)
		return c
	}
	c.DataKey = ""
	return  c
}

func Md5(str string) string  {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// Build 构建出configmap对象
func (c *ConfigMapBuilder) Build(ctx context.Context) (*corev1.ConfigMap, error) {


	if c.CM.CreationTimestamp.IsZero() {
		c.apply().setOwner().parseKey() // 更新cm对象的字段，且设置ownerReferences
		err := c.Client.Create(ctx, c.CM)
		if err != nil {
			klog.Error("create configmap err: ", err)
			return nil, err
		}
	} else {

		// 更新:法一 update方式
		c.apply().parseKey()
		err := c.Client.Update(ctx, c.CM)
		if err != nil {
			klog.Error("update configmap err: ", err)
			return nil, err
		}
	}
	return c.CM, nil
}

