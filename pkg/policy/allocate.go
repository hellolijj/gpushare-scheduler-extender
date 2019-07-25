package policy

import (
	"fmt"
	"log"
	
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func (p Policy) Allocate(clientset *kubernetes.Clientset, pod *v1.Pod, n *utils.NodeInfo) (err error) {
	var newPod *v1.Pod
	p.rwmu.Lock()
	defer p.rwmu.Unlock()
	log.Printf("debug: Allocate() ----Begin to allocate GPU for gpu topology for pod %s in ns %s----", pod.Name, pod.Namespace)
	// 1. Update the pod spec
	devIds, found := p.allocateGPUID(pod, n)
	if found {
		log.Printf("debug: Allocate() 1. Allocate GPU ID %v to pod %s in ns %s.----", devIds, pod.Name, pod.Namespace)
		
		newPod = utils.GetUpdatedPodAnnotationSpec(pod, devIds)
		_, err = clientset.CoreV1().Pods(newPod.Namespace).Update(newPod)
		if err != nil {
			// the object has been modified; please apply your changes to the latest version and try again
			if err.Error() == utils.OptimisticLockErrorMsg {
				// retry
				pod, err = clientset.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				// newPod = utils.GetUpdatedPodEnvSpec(pod, devId, nodeInfo.GetTotalGPUMemory()/nodeInfo.GetGPUCount())
				newPod = utils.GetUpdatedPodAnnotationSpec(pod, devIds)
				_, err = clientset.CoreV1().Pods(newPod.Namespace).Update(newPod)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	} else {
		err = fmt.Errorf("the node %s can't place the pod %s in ns %s", pod.Spec.NodeName, pod.Name, pod.Namespace)
	}

	// 2. Bind the pod to the node
	if err == nil {
		binding := &v1.Binding{
			ObjectMeta: metav1.ObjectMeta{Name: pod.Name, UID: pod.UID},
			Target:     v1.ObjectReference{Kind: "Node", Name: n.GetName()},
		}
		log.Printf("debug: Allocate() 2. Try to bind pod %s in %s namespace to node %s with %v",
			pod.Name,
			pod.Namespace,
			pod.Spec.NodeName,
			binding)
		err = clientset.CoreV1().Pods(pod.Namespace).Bind(binding)
		if err != nil {
			log.Printf("warn: Failed to bind the pod %s in ns %s due to %v", pod.Name, pod.Namespace, err)
			return err
		}
	}

	// 3. update the device info if the pod is update successfully
	if err == nil {
		log.Printf("debug: Allocate() 3. Try to add pod %s in ns %s to dev %v",
			pod.Name,
			pod.Namespace,
			devIds)
		devs := n.GetDevs()
		for _, devId := range devIds {
			dev, found := devs[int(devId)]
			if !found {
				log.Printf("warn: Pod %s in ns %s failed to find the GPU ID %d in node %s", pod.Name, pod.Namespace, devId, n.GetName())
			} else {
				dev.AddPod(newPod)
			}
		}
	}
	log.Printf("debug: Allocate() ----End to allocate GPU for gpu mem for pod %s in ns %s----", pod.Name, pod.Namespace)
	return err

}


func (policy Policy) allocateGPUID(pod *v1.Pod, n *utils.NodeInfo) (candidateDevID []uint, found bool) {
	reqGPU := 0
	found = false
	availableGPUs := n.GetAvailableGPUs()
	reqGPU = utils.GetGPUCountFromPodResource(pod)
	log.Printf("debug: reqGPU for pod %s in ns %s: %d", pod.Name, pod.Namespace, reqGPU)
	log.Printf("debug: AvailableGPUs: %v in node %s", availableGPUs, n.GetName())
	
	if reqGPU > 0 {
		if availableGPUs > 0 && availableGPUs-reqGPU >= 0 {
			ids, err := policy.Run.Allocate(n, reqGPU)
			if err != nil {
				log.Printf("allocate gpu to node failed, resaon: %v", err)
				return
			}
			for _, id := range ids {
				candidateDevID = append(candidateDevID, uint(id))
			}
			found = true
		}
		if found {
			log.Printf("debug: Find candidate dev id %d for pod %s in ns %s successfully.",
				candidateDevID,
				pod.Name,
				pod.Namespace)
		} else {
			log.Printf("warn: Failed to find available GPUs %d for the pod %s in the namespace %s",
				reqGPU,
				pod.Name,
				pod.Namespace)
		}
	}
	
	return candidateDevID, found
}
