package utils

const (
	ResourceName = "aliyun.com/gpu-count"

	EnvNVGPU              = "NVIDIA_VISIBLE_DEVICES"
	EnvResourceIndex      = "ALIYUN_COM_GPU_IDX"       // 在 annotation 标记使用哪些gpuid 格式 1,2,4 or 2
	EnvResourceByPod      = "ALIYUN_COM_GPU_COUNT_POD" // 在 annotation 标记这个pod 使用了几个 gpu  格式: 1 or 0 or 4
	EnvAssignedFlag       = "ALIYUN_COM_GPU_ASSIGNED"
	EnvResourceAssumeTime = "ALIYUN_COM_GPU_ASSUME_TIME"
)
