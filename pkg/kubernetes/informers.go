/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"sync"
	"time"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/sample-controller/pkg/signals"
)

var (
	onceForInformerManager  sync.Once
	informerManagerInstance *InformerManager
)

func noResyncPeriodFunc() time.Duration {
	return 0
}

// NewInformer creates a new K8S client based on a service account
func NewInformer(client clientset.Interface) *InformerManager {
	onceForInformerManager.Do(func() {
		informerManagerInstance = &InformerManager{
			client:          client,
			stopCh:          signals.SetupSignalHandler(),
			informerFactory: informers.NewSharedInformerFactory(client, noResyncPeriodFunc()),
		}
	})
	return informerManagerInstance
}

// AddNodeListener hooks up add, update, delete callbacks
func (im *InformerManager) AddNodeListener(add func(obj interface{}), update func(oldObj, newObj interface{}), remove func(obj interface{})) {
	if im.nodeInformer == nil {
		im.nodeInformer = im.informerFactory.Core().V1().Nodes().Informer()
	}

	im.nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    add,
		UpdateFunc: update,
		DeleteFunc: remove,
	})
}

// AddPVCListener hooks up add, update, delete callbacks
func (im *InformerManager) AddPVCListener(add func(obj interface{}), update func(oldObj, newObj interface{}), remove func(obj interface{})) {
	if im.pvcInformer == nil {
		im.pvcInformer = im.informerFactory.Core().V1().PersistentVolumeClaims().Informer()
	}
	im.pvcSynced = im.pvcInformer.HasSynced

	im.pvcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    add,
		UpdateFunc: update,
		DeleteFunc: remove,
	})
}

// AddPVListener hooks up add, update, delete callbacks
func (im *InformerManager) AddPVListener(add func(obj interface{}), update func(oldObj, newObj interface{}), remove func(obj interface{})) {
	if im.pvInformer == nil {
		im.pvInformer = im.informerFactory.Core().V1().PersistentVolumes().Informer()
	}
	im.pvSynced = im.pvInformer.HasSynced

	im.pvInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    add,
		UpdateFunc: update,
		DeleteFunc: remove,
	})
}

// AddConfigMapListener hooks up add, update, delete callbacks
func (im *InformerManager) AddConfigMapListener(add func(obj interface{}), update func(oldObj, newObj interface{}), remove func(obj interface{})) {
	if im.configMapInformer == nil {
		im.configMapInformer = im.informerFactory.Core().V1().ConfigMaps().Informer()
	}
	im.configMapSynced = im.configMapInformer.HasSynced

	im.configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    add,
		UpdateFunc: update,
		DeleteFunc: remove,
	})
}

// AddPodListener hooks up add, update, delete callbacks
func (im *InformerManager) AddPodListener(add func(obj interface{}), update func(oldObj, newObj interface{}), remove func(obj interface{})) {
	if im.podInformer == nil {
		im.podInformer = im.informerFactory.Core().V1().Pods().Informer()
	}
	im.podSynced = im.podInformer.HasSynced

	im.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    add,
		UpdateFunc: update,
		DeleteFunc: remove,
	})
}

// GetPVLister returns Persistent Volume Lister for the calling informer manager
func (im *InformerManager) GetPVLister() corelisters.PersistentVolumeLister {
	return im.informerFactory.Core().V1().PersistentVolumes().Lister()
}

// GetPVCLister returns PVC Lister for the calling informer manager
func (im *InformerManager) GetPVCLister() corelisters.PersistentVolumeClaimLister {
	return im.informerFactory.Core().V1().PersistentVolumeClaims().Lister()
}

// GetConfigMapLister returns ConfigMap Lister for the calling informer manager
func (im *InformerManager) GetConfigMapLister() corelisters.ConfigMapLister {
	return im.informerFactory.Core().V1().ConfigMaps().Lister()
}

// GetPodLister returns Pod Lister for the calling informer manager
func (im *InformerManager) GetPodLister() corelisters.PodLister {
	return im.informerFactory.Core().V1().Pods().Lister()
}

// Listen starts the Informers
func (im *InformerManager) Listen() (stopCh <-chan struct{}) {
	go im.informerFactory.Start(im.stopCh)
	if im.pvSynced != nil && im.pvcSynced != nil && im.podSynced != nil && im.configMapSynced != nil {
		if !cache.WaitForCacheSync(im.stopCh, im.pvSynced, im.pvcSynced, im.podSynced, im.configMapSynced) {
			return
		}

	}
	return im.stopCh
}
