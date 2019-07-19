package utils

const (
	ResourceName = "aliyun.com/gpu"
	
	EnvResourceIndex      = "ALIYUN_COM_GPU_GROUP" // 在 annotation 标记使用哪些gpuid 格式 1,2,4 or 2
	EnvAssignedFlag       = "ALIYUN_COM_GPU_ASSIGNED"
	EnvResourceAssumeTime = "ALIYUN_COM_GPU_ASSUME_TIME"
	EnvGPUAnnotation      = "GPU_TOPOLOGY"
	
	GPU_PRIFX   = "GPU_"
)
