package builders

// 模版内容
const DeploymentTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dbcore-{{ .Name }}
  namespace: {{ .Namespace }}
spec:
  selector:
    matchLabels:
      app: dbcore-{{ .Namespace}}-{{ .Name }}
  replicas: 1
  template:
    metadata:
      labels:
        app: dbcore-{{ .Namespace}}-{{ .Name }}
        version: v1
    spec:
      initContainers:
        - name: init-test
          image: busybox:1.28
          command: ['sh', '-c', 'echo sleeping && sleep 15']
      containers:
        - name: dbcore-{{ .Namespace}}-{{ .Name }}-container
          image: docker.io/shenyisyn/dbcore:v1
          imagePullPolicy: IfNotPresent
          ports:
             - containerPort: 8081
             - containerPort: 8090

`
