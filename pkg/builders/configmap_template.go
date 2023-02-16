package builders

// configmap对应模板
// 因为都是固定不变的，可以不用放入模版中。
//const ConfigmapTemplate =`
//apiVersion: v1
//kind: ConfigMap
//metadata:
// name: dbcore-{{ .Name }}
// namespace: {{ .Namespace }}
//data:
// app.yml: |
//  dbConfig:
//   dsn: ""
//   maxOpenConn: 20
//   maxLifeTime: 1800
//   maxIdleConn: 5
//  appConfig:
//   rpcPort: 8081
//   httpPort: 8090
//
//  apis:
//   - name: test
//     sql: "select * from test"
//`


// configmap对应模板
const ConfigmapTemplate = `
  dbConfig:
   dsn: "{{ .Dsn }}"
   maxOpenConn: {{ .MaxOpenConn }}
   maxLifeTime: {{ .MaxLifeTime }}
   maxIdleConn: {{ .MaxIdleConn }}
  appConfig:
   rpcPort: 8081
   httpPort: 8090
  apis:
   - name: test
     sql: "select * from test"
`