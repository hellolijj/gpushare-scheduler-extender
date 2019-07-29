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
	gpuAbbr = map[nvml.P2PLinkType]string{
		1: "PSB",
		2: "PIX",
		3: "PXB",
		4: "PHB" ,
		5: "NODE",
		6: "SYS",
		7: "NV1",
		8: "NV2",
		9: "NV3",
		10: "NV4",
		11: "NV5",
		12: "NV6",
		
	}
)

func GetGPULinkFromDescAndAbbr(abbr string) nvml.P2PLinkType {
	abbrKey, ok := abbrGpu[abbr]
	if !ok {
		return 0
	}
	return abbrKey
}

func GetGPUAbbr(linkType nvml.P2PLinkType) string {
	abbr, ok := gpuAbbr[linkType]
	if !ok {
		return ""
	}
	return abbr
}