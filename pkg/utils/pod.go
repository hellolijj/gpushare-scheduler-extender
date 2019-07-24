package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
)

// AssignedNonTerminatedPod selects pods that are assigned and non-terminal (scheduled and running).
func AssignedNonTerminatedPod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return false
	}

	if len(pod.Spec.NodeName) == 0 {
		return false
	}
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return false
	}
	return true
}

// IsCompletePod determines if the pod is complete
func IsCompletePod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return true
	}

	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return true
	}
	return false
}

// IsGSoCPod determines if it's the pod for gsoc scheduler pod
func IsGPUPod(pod *v1.Pod) bool {
	return GetGPUCountFromPodResource(pod) > 0
}

// GetGPUIDFromAnnotation gets GPU ID from Annotation
func GetGPUIDFromAnnotation(pod *v1.Pod) []int {
	ids := []int{}
	if len(pod.ObjectMeta.Annotations) > 0 {
		value, found := pod.ObjectMeta.Annotations[EnvResourceIndex]
		if found {
			idList := strings.Split(value, ",")
			for _, idStr := range idList {
				id, err := strconv.Atoi(idStr)
				if err != nil {
					return []int{}
				}
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// GetGPUCountFromPodAnnotation gets the GPU Count of the pod, choose the larger one between gpu memory and gpu init container memory
func GetGPUCountFromPodAnnotation(pod *v1.Pod) (gpuCount int) {
	gpuCount = len(GetGPUIDFromAnnotation(pod))
	log.Printf("debug: pod %s in ns %s with status %v has GPU Count %d",
		pod.Name,
		pod.Namespace,
		pod.Status.Phase,
		gpuCount)
	return gpuCount
}

func GetGPUCountFromPodResource(pod *v1.Pod) int {
	var total int
	containers := pod.Spec.Containers
	for _, container := range containers {
		if val, ok := container.Resources.Limits[ResourceName]; ok {
			total += int(val.Value())
		}
	}
	return total
}

// GetUpdatedPodAnnotationSpec updates pod env with devIds
func GetUpdatedPodAnnotationSpec(oldPod *v1.Pod, devIds []uint) (newPod *v1.Pod) {
	newPod = oldPod.DeepCopy()
	if len(newPod.ObjectMeta.Annotations) == 0 {
		newPod.ObjectMeta.Annotations = map[string]string{}
	}
	var devs string
	for i, devId := range devIds {
		if i == 0 {
			devs += fmt.Sprintf("%d", devId)
		} else {
			devs += fmt.Sprintf(",%d", devId)
		}
	}

	now := time.Now()
	newPod.ObjectMeta.Annotations[EnvResourceIndex] = devs
	newPod.ObjectMeta.Annotations[EnvAssignedFlag] = "false"
	newPod.ObjectMeta.Annotations[EnvResourceAssumeTime] = fmt.Sprintf("%d", now.UnixNano())
	return newPod
}
