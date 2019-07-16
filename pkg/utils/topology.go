package utils

import "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

// 根据 gpu 的描述及缩写 返回 gpu 的id

var (
	desc_gpu = map[string]nvml.P2PLinkType{
		"Cross CPU socket": 1,
		"Same CPU socket":  2,
		"Host PCI bridge": 3,
		"Multiple PCI switches": 4,
		"Single PCI switch": 5,
		"Same board": 6,
		"Single NVLink": 7,
		"Two NVLinks": 8,
		"Three NVLinks": 9,
		"Four NVLinks": 10,
		"Five NVLinks": 11,
		"Six NVLinks": 12,
	}
	
	abbr_gpu = map[string]nvml.P2PLinkType{
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



func GetGPULinkFromDescAndAbbr(desc, abbr string) nvml.P2PLinkType {
	descKey, ok := desc_gpu[desc]
	if !ok {
		return 0
	}
	abbrKey, ok := abbr_gpu[abbr]
	if !ok {
		return 0
	}
	if descKey != abbrKey {
		return 0
	}
	return descKey
}
