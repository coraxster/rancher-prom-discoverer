package main

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
)

type TargetWriter struct {
	filepath    string
	lastWritten []*RancherTarget
}

func NewTargetWriter(filepath string) *TargetWriter {
	return &TargetWriter{filepath: filepath}
}

func (tw *TargetWriter) Write(rancherTargets []*RancherTarget) error {
	if reflect.DeepEqual(rancherTargets, tw.lastWritten) {
		log.Println("[INFO] not changed")
		return nil
	}
	promoTargets := make([]*PromTarget, 0, len(rancherTargets))
	for _, rt := range rancherTargets {
		promoTargets = append(promoTargets, rancher2PromTarget(rt))
	}
	f, err := os.Create(tw.filepath)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(promoTargets)
	if err != nil {
		return err
	}
	tw.lastWritten = rancherTargets
	log.Println("[INFO] file written")
	return nil
}

type PromTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func rancher2PromTarget(rt *RancherTarget) *PromTarget {
	promLabels := make(map[string]string)
	for k, v := range rt.Labels {
		promLabels[k] = v
	}
	promLabels["hostname"] = rt.Host
	promLabels["stack"] = rt.Stack
	promLabels["service"] = rt.Service
	return &PromTarget{
		Targets: []string{rt.Target},
		Labels:  promLabels,
	}
}
