// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	vdc *valueDescriptorCache
)

type ValueDescriptorCache interface {
	ForName(name string) (contract.ValueDescriptor, bool)
	All() []contract.ValueDescriptor
	Add(descriptor contract.ValueDescriptor) error
	Update(descriptor contract.ValueDescriptor) error
	Remove(id string) error
	RemoveByName(name string) error
}

type valueDescriptorCache struct {
	vdMap   map[string]contract.ValueDescriptor // key is ValueDescriptor name
	nameMap map[string]string                   // key is id, and value is ValueDescriptor name
	mutex   sync.Mutex
}

func (v *valueDescriptorCache) ForName(name string) (contract.ValueDescriptor, bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	vd, ok := v.vdMap[name]
	return vd, ok
}

func (v *valueDescriptorCache) All() []contract.ValueDescriptor {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	vds := make([]contract.ValueDescriptor, len(v.vdMap))
	i := 0
	for _, vd := range v.vdMap {
		vds[i] = vd
		i++
	}
	return vds
}

func (v *valueDescriptorCache) Add(descriptor contract.ValueDescriptor) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.add(descriptor)
}

func (v *valueDescriptorCache) add(descriptor contract.ValueDescriptor) error {
	_, ok := v.vdMap[descriptor.Name]
	if ok {
		return fmt.Errorf("value descriptor %s has already existed in cache", descriptor.Name)
	}
	v.vdMap[descriptor.Name] = descriptor
	v.nameMap[descriptor.Id] = descriptor.Name
	return nil
}

func (v *valueDescriptorCache) Update(descriptor contract.ValueDescriptor) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if err := v.remove(descriptor.Id); err != nil {
		return err
	}
	return v.add(descriptor)
}

func (v *valueDescriptorCache) Remove(id string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.remove(id)
}

func (v *valueDescriptorCache) remove(id string) error {
	name, ok := v.nameMap[id]
	if !ok {
		return fmt.Errorf("value descriptor %s does not exist in cache", id)
	}

	return v.removeByName(name)
}

func (v *valueDescriptorCache) RemoveByName(name string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.removeByName(name)
}

func (v *valueDescriptorCache) removeByName(name string) error {
	vd, ok := v.vdMap[name]
	if !ok {
		return fmt.Errorf("value descriptor %s does not exist in cache", name)
	}
	delete(v.nameMap, vd.Id)
	delete(v.vdMap, name)
	return nil
}

func newValueDescriptorCache(descriptors []contract.ValueDescriptor) ValueDescriptorCache {
	defaultSize := len(descriptors) * 2
	vdMap := make(map[string]contract.ValueDescriptor, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	for _, vd := range descriptors {
		vdMap[vd.Name] = vd
		nameMap[vd.Id] = vd.Name
	}
	vdc = &valueDescriptorCache{vdMap: vdMap, nameMap: nameMap}
	return vdc
}

func ValueDescriptors() ValueDescriptorCache {
	return vdc
}
