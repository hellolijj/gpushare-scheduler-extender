package utils

import "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

// 根据 gpu 的描述及缩写 返回 gpu 的id

var (
	abbrGpu = map[string]nvml.P2PLinkType{
		"PSB": 1,
		"PIX": 2,
		"PXB": 3,
		"PHB": 4,
		"NODE": 5,
		"SYS": 6,
		"NV1": 7,
		"NV2": 8,
		"NV3": 9,
		"NV4": 10,
		"NV5": 11,
		"NV6": 12,
	}
)

func GetGPULinkFromDescAndAbbr(abbr string) nvml.P2PLinkType {
	abbrKey, ok := abbrGpu[abbr]
	if !ok {
		return 0
	}
	return abbrKey
}
