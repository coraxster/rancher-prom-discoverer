package main

import (
	"encoding/json"
	"log"
	"os"
)

type TargetWriter struct {
	filepath string
}

func NewTargetWriter(filepath string) *TargetWriter {
	return &TargetWriter{filepath: filepath}
}

func (tw *TargetWriter) Write(rancherTargets []*RancherTarget) error {
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
	log.Println("[INFO] file written")
	return nil
}

type PromTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func rancher2PromTarget(rt *RancherTarget) *PromTarget {
	rt.Labels["hostname"] = rt.Host
	rt.Labels["stack"] = rt.Stack
	rt.Labels["service"] = rt.Service
	return &PromTarget{
		Targets: []string{rt.Target},
		Labels:  rt.Labels,
	}
}
