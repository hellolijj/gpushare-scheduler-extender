# gputopology 使用指导

<a name="h8rFZ"></a>
# 测试系统
安装完成后，使用如下命令训练一个 gpu 任务

```bash
kubectl apply -f https://raw.githubusercontent.com/hellolijj/gpushare-scheduler-extender/gsoc/samples/test.yaml
```

通过查看 log 日志。如果出现如下结果则说明任务训练完成。<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1564294376195-f6917e1b-9798-4c88-b525-9e8c35efb73a.png#align=left&display=inline&height=531&name=image.png&originHeight=1062&originWidth=1052&size=332388&status=done&width=526)
<a name="mMBcX"></a>
##
<a name="K7SVr"></a>
# 问题排查
<a name="W1HXj"></a>
## 检查device-plugin是否上传topo信息

node节点的topology:
```bash
[root@iz8vbffcyrsv82qww1fty8z k8s-device-plugin]# nvidia-smi topo -m
	GPU0	GPU1	GPU2	GPU3	GPU4	GPU5	GPU6	GPU7	CPU Affinity
GPU0	 X 	NV1	NV1	NV2	SYS	SYS	NV2	SYS	0-95
GPU1	NV1	 X 	NV2	NV1	SYS	SYS	SYS	NV2	0-95
GPU2	NV1	NV2	 X 	NV2	NV1	SYS	SYS	SYS	0-95
GPU3	NV2	NV1	NV2	 X 	SYS	NV1	SYS	SYS	0-95
GPU4	SYS	SYS	NV1	SYS	 X 	NV2	NV1	NV2	0-95
GPU5	SYS	SYS	SYS	NV1	NV2	 X 	NV2	NV1	0-95
GPU6	NV2	SYS	SYS	SYS	NV1	NV2	 X 	NV1	0-95
GPU7	SYS	NV2	SYS	SYS	NV2	NV1	NV1	 X 	0-95

Legend:

  X    = Self
  SYS  = Connection traversing PCIe as well as the SMP interconnect between NUMA nodes (e.g., QPI/UPI)
  NODE = Connection traversing PCIe as well as the interconnect between PCIe Host Bridges within a NUMA node
  PHB  = Connection traversing PCIe as well as a PCIe Host Bridge (typically the CPU)
  PXB  = Connection traversing multiple PCIe switches (without traversing the PCIe Host Bridge)
  PIX  = Connection traversing a single PCIe switch
  NV#  = Connection traversing a bonded set of # NVLinks
```

![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1563023135963-7a3e6238-54cb-420b-9368-3ff84f67e99a.png#align=left&display=inline&height=373&name=image.png&originHeight=746&originWidth=1594&size=263427&status=done&width=797)

上传node 的 annotation：
```yaml
[root@iZ8vbazwei4j05nbediqaeZ lijj]# kubectl get node cn-zhangjiakou.192.168.0.113 -o yaml | head -n 20
apiVersion: v1
kind: Node
metadata:
  annotations:
    GPU_TOPOLOGY: '{"GPU_NV1_0_1":"Single NVLink","GPU_NV1_0_2":"Single NVLink","GPU_NV1_1_3":"Single
      NVLink","GPU_NV1_2_4":"Single NVLink","GPU_NV1_3_5":"Single NVLink","GPU_NV1_4_6":"Single
      NVLink","GPU_NV1_5_7":"Single NVLink","GPU_NV1_6_7":"Single NVLink","GPU_NV2_0_3":"Two
      NVLinks","GPU_NV2_0_6":"Two NVLinks","GPU_NV2_1_2":"Two NVLinks","GPU_NV2_1_7":"Two
      NVLinks","GPU_NV2_2_3":"Two NVLinks","GPU_NV2_4_5":"Two NVLinks","GPU_NV2_4_7":"Two
      NVLinks","GPU_NV2_5_6":"Two NVLinks","GPU_SYS_0_4":"Cross CPU socket","GPU_SYS_0_5":"Cross
      CPU socket","GPU_SYS_0_7":"Cross CPU socket","GPU_SYS_1_4":"Cross CPU socket","GPU_SYS_1_5":"Cross
      CPU socket","GPU_SYS_1_6":"Cross CPU socket","GPU_SYS_2_5":"Cross CPU socket","GPU_SYS_2_6":"Cross
      CPU socket","GPU_SYS_2_7":"Cross CPU socket","GPU_SYS_3_4":"Cross CPU socket","GPU_SYS_3_6":"Cross
      CPU socket","GPU_SYS_3_7":"Cross CPU socket"}'
    NODE_TYPE: ecs.sccgn6.24xlarge
    flannel.alpha.coreos.com/backend-data: "null"
    flannel.alpha.coreos.com/backend-type: ""
    flannel.alpha.coreos.com/kube-subnet-manager: "true"
    flannel.alpha.coreos.com/public-ip: 192.168.0.113
    kubeadm.alpha.kubernetes.io/cri-socket: /var/run/dockershim.sock
    node.alpha.kubernetes.io/ttl: "0"
  creationTimestamp: 2019-05-31T07:54:08Z
  labels:
[root@iZ8vbazwei4j05nbediqaeZ lijj]#
```

<a name="uA25C"></a>
## 检查是否经过 topology 调度
通过查看 node 的 annotation 字段。annotation 出现如下 ALIYUN_COM_GPU_ASSIGNED: true 则经过 topology 调度<br />


请求 aliyun.com/gpu = 1 的调度<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1562637864784-1d36b6b0-c2a8-4b24-83c6-93df9bcd0003.png#align=left&display=inline&height=534&name=image.png&originHeight=1122&originWidth=1524&size=306023&status=done&width=725)

请求 aliyun.com/gpu = 2 的调度<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1562772853441-8d960763-f7e8-4bc0-8f4a-d051ebf44dda.png#align=left&display=inline&height=571&name=image.png&originHeight=1142&originWidth=2142&size=388103&status=done&width=1071)


如果没有这个字段则没有经过 topology 调度

【普通调度】

![image.png](https://cdn.nlark.com/yuque/0/2019/png/394957/1562727793947-2637e856-b0b2-492a-8d1b-5dbb77203a98.png#align=left&display=inline&height=417&name=image.png&originHeight=834&originWidth=1416&size=218765&status=done&width=708)
