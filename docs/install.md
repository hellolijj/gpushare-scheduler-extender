# gputopology 安装文档

<a name="JhL1G"></a>
# 1. 安装 gputopology-device-plugin

<a name="eA1nV"></a>
## 1.1 配置容器支持 nvidia-docker

给每个 node 节点上安装 nividia-docker 安装教程参见[官方文档](https://github.com/NVIDIA/nvidia-docker/wiki/Installation-(version-2.0))。编辑 `/etc/docker/daemon.json` 文件为如下内容：

```json
{
   "default-runtime": "nvidia",
   "runtimes": {
        "nvidia": {
            "path": "/usr/bin/nvidia-container-runtime",
            "runtimeArgs": []
        }
    },
   "registry-mirrors": ["https://cagz8nbe.mirror.aliyuncs.com"]
}
```

> 建议也配置阿里云镜像下载加速器。


重新加载配置文件
```bash
# systemctl daemon-reload
# systemctl restart docker
```

检查 docker runtime<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1562773092661-01701200-756b-425a-8940-8b26fe72db40.png#align=left&display=inline&height=93&name=image.png&originHeight=186&originWidth=954&size=59640&status=done&width=477)

<a name="O3Zai"></a>
## [](https://www.yuque.com/lijunjun-bjkm9/rwyxlc/quqwog#MKZ6D)1.2 部署 device-plugin 

执行如下命令 部署 device-plugin

```bash
# kubectl apply -f https://github.com/hellolijj/k8s-device-plugin/raw/master/deploy/gputopology-device-plugin.yaml
```

> ⚠️如果节点上已经安装了 nvidia-plugin 需要先将其删掉。如果是 static pod 部署模式，需要从 /etc/kubernetes/manifest 目录删除部署文件。

<a name="sp8Mo"></a>
## 1.3 给节点打标签使支持gpu topology

```yaml
# kubectl label node <target_node> gputopology=true
```

<a name="sa3Zz"></a>
## 1.4 验证 

```bash
[root@iZ8vbazwei4j05nbediqaeZ lijj]# kubectl get pods --all-namespaces -o wide -w | grep device
kube-system    gputopology-device-plugin-ds-c652q                     1/1     Running     0          47h     192.168.0.112   cn-zhangjiakou.192.168.0.112   <none>
kube-system    gputopology-device-plugin-ds-mh7k9                     1/1     Running     0          47h     192.168.0.113   cn-zhangjiakou.192.168.0.113   <none>
kube-system    nvidia-device-plugin-cn-zhangjiakou.192.168.0.112      1/1     Running     6          15d     192.168.0.112   cn-zhangjiakou.192.168.0.112   <none>
kube-system    nvidia-device-plugin-cn-zhangjiakou.192.168.0.113      1/1     Running     4          19d     192.168.0.113   cn-zhangjiakou.192.168.0.113   <none>
```

执行以上命令，出现如下情况，则部署成功<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1564472587692-311f29d0-d0dd-4a1c-b43b-30b90ef5b347.png#align=left&display=inline&height=96&name=image.png&originHeight=192&originWidth=2336&size=156400&status=done&width=1168)
<a name="DRRfR"></a>
###
<a name="SCz2t"></a>
# 2. 安装 gputopology-scheduele-extender 

<a name="ZgJR2"></a>
## 2.1 部署 static-policy 策略

```json
{
  "ecs.sccgn6.24xlarge": {
    "1": [
      [0], [1], [2], [3], [4], [5], [6], [7]
    ],
    "2": [
      [0, 2],
      [1, 3],
      [4, 6],
      [5, 7]
    ],
    "4": [
      [0, 1, 2, 3],
      [4, 5, 6, 7]
    ],
    "8": [
      [0, 1, 2, 3, 4, 5, 6, 7]
    ]
  }
}
```

```bash
# curl -o /etc/kubernetes/static-node.json https://raw.githubusercontent.com/hellolijj/gputopology-scheduler-extender/master/config/static-node.json
```

<a name="2qrue"></a>
## 2.1 部署 gputopology-scheduler-extender

执行以下命令，部署 gputopology-scheduler-extender

```bash
# kubectl apply -f https://raw.githubusercontent.com/hellolijj/gputopology-scheduler-extender/master/config/gputopology-schd-extender.yaml
```

> TODO: extender 类型 deployment 改为 daemonset。或者 replicas 改为 master节点个数。


<a name="rWjpe"></a>
## [](https://www.yuque.com/lijunjun-bjkm9/rwyxlc/quqwog#EJleN)2.2 master 节点上配置 scheduler extender 策略
编辑 scheduler-extender 配置文件于 /etc/kubernetes/gputopology-scheduler-policy-config.json
```json
{
  "kind": "Policy",
  "apiVersion": "v1",
  "priorities": [
    {"name": "LeastRequestedPriority", "weight": 1},
    {"name": "BalancedResourceAllocation", "weight": 1},
    {"name": "ServiceSpreadingPriority", "weight": 1},
    {"name": "EqualPriority", "weight": 1}
  ],
  "extenders": [
    {
      "urlPrefix": "http://127.0.0.1:32743/gputopology-scheduler",
      "PrioritizeVerb": "sort",
      "weight": 4,
      "bindVerb":   "bind",
      "enableHttps": false,
      "nodeCacheCapable": true,
      "managedResources": [
        {
          "name": "aliyun.com/gpu",
          "ignoredByScheduler": false
        }
      ],
      "ignorable": false
    }
  ]
}
```
或者执行以下命令
```bash
# curl -o /etc/kubernetes/gputopology-scheduler-policy-config.json https://raw.githubusercontent.com/hellolijj/gputopology-scheduler-extender/master/config/gputopology-scheduler-policy-config.json
```


<a name="8suz0"></a>
## 2.3 配置 scheduler 参数，重启

编辑 scheuler static pod yaml文件， 重新启动使得 配置文件生效。

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ""
  creationTimestamp: null
  labels:
    component: kube-scheduler
    tier: control-plane
  name: kube-scheduler
  namespace: kube-system
spec:
  containers:
  - command:
    - kube-scheduler
    - --address=127.0.0.1
    - --kubeconfig=/etc/kubernetes/scheduler.conf
    - --leader-elect=true
    - --policy-config-file=/etc/kubernetes/gputopology-scheduler-policy-config.json
    - -v=4
    image: registry-vpc.cn-shanghai.aliyuncs.com/acs/kube-scheduler:v1.12.6-aliyun.1
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 8
      httpGet:
        host: 127.0.0.1
        path: /healthz
        port: 10251
        scheme: HTTP
      initialDelaySeconds: 15
      timeoutSeconds: 15
    name: kube-scheduler
    resources:
      requests:
        cpu: 100m
    volumeMounts:
    - mountPath: /etc/kubernetes/scheduler.conf
      name: kubeconfig
      readOnly: true
    - mountPath: /etc/localtime
      name: localtime
      readOnly: true
    - mountPath: /etc/kubernetes/gputopology-scheduler-policy-config.json
      name: scheduler-policy-config
      readOnly: true
  hostNetwork: true
  priorityClassName: system-cluster-critical
  volumes:
  - hostPath:
      path: /etc/localtime
      type: ""
    name: localtime
  - hostPath:
      path: /etc/kubernetes/scheduler.conf
      type: FileOrCreate
    name: kubeconfig
  - hostPath:
      path: /etc/kubernetes/gputopology-scheduler-policy-config.json
      type: FileOrCreate
    name: scheduler-policy-config
status: {}
```

> 通过 vim 编辑时，不可在 /etc/kuberntes/manifest 下编辑。可在其他目录下编辑 mv 至此文件夹。


也可以通过以下命令安装。
```bash
# curl -o /etc/kubernetes/manifests/kube-scheduler.yaml https://raw.githubusercontent.com/hellolijj/gputopology-scheduler-extender/master/config/kube-scheduler.yaml
```

> 重启 scheduler 的过程，及 配置 scheudler 策略文件需要在每一个 master 节点上执行。
> ⚠️要保证 每一个master 都部署了 extender 并重启了 scheduler。

<a name="3qith"></a>
#
<a name="JTJAy"></a>
## 2.4 验证

```bash
$ kubectl get pods --all-namespaces | grep scheduler
```

执行以上命令，出现如下情况，则部署成功。<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1562762437400-9817d044-90ef-445d-8709-e2638b63f1fa.png#align=left&display=inline&height=209&name=image.png&originHeight=418&originWidth=1536&size=254619&status=done&width=768)

