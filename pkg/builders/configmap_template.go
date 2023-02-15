package builders

// configmap对应模板
const ConfigmapTemplate =`
apiVersion: v1
kind: ConfigMap
metadata:
 name: dbcore-{{ .Name }}
 namespace: {{ .Namespace }}
data:
 app.yaml: |
  dbConfig:
   dsn: ""
   maxOpenConn: 20
   maxLifeTime: 1800
   maxIdleConn: 5
  appConfig:
   rpcPort: 8081
   httpPort: 8090

  apis:
   - name: test
     sql: "select * from test"
`