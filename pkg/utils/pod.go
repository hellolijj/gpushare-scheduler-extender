package utils

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	"strings"
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
func IsGSoCPod(pod *v1.Pod) bool {
	return true
}


// GetGPUIDFromAnnotation gets GPU ID from Annotation
func GetGPUIDFromAnnotation(pod *v1.Pod) []int {
	ids := []int{}
	if len(pod.ObjectMeta.Annotations) > 0 {
		value, found := pod.ObjectMeta.Annotations[EnvResourceIndex]
		if found {
			idList := strings.Split(value, ",")
			for id := range idList {
				ids = append(ids, int(id))
			}
		}
	}
	return ids
}

// GetGPUIDFromEnv gets GPU ID from Env
func GetGPUIDFromEnv(pod *v1.Pod) int {
	id := -1
	for _, container := range pod.Spec.Containers {
		id = getGPUIDFromContainer(container)
		if id >= 0 {
			return id
		}
	}

	return id
}

func getGPUIDFromContainer(container v1.Container) (devIdx int) {
	devIdx = -1
	var err error
loop:
	for _, env := range container.Env {
		if env.Name == EnvResourceIndex {
			devIdx, err = strconv.Atoi(env.Value)
			if err != nil {
				log.Printf("warn: Failed due to %v for %s", err, container.Name)
				devIdx = -1
			}
			break loop
		}
	}

	return devIdx
}

// GetGPUCountFromPodAnnotation gets the GPU Count of the pod, choose the larger one between gpu memory and gpu init container memory
func GetGPUCountFromPodAnnotation(pod *v1.Pod) (gpuCount int) {
	if len(pod.ObjectMeta.Annotations) > 0 {
		value, found := pod.ObjectMeta.Annotations[EnvResourceByPod]
		if found {
			s, _ := strconv.Atoi(value)
			if s < 0 {
				s = 0
			}
			
			gpuCount += s
		}
	}

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
			devs += fmt.Sprintf("_%d", devId)
		}
	}

	now := time.Now()
	newPod.ObjectMeta.Annotations[EnvResourceIndex] = devs
	newPod.ObjectMeta.Annotations[EnvAssignedFlag] = "false"
	newPod.ObjectMeta.Annotations[EnvResourceAssumeTime] = fmt.Sprintf("%d", now.UnixNano())
	return newPod
}
