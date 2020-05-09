package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"log"
	"net/http"
	"strings"
)

type Rancher struct {
	host    string
	token   string
	project string
}

func NewRancher(host, token, project string) *Rancher {
	return &Rancher{host: host, token: token, project: project}
}

type RancherTarget struct {
	Host    string
	Stack   string
	Service string
	Target  string
	Labels  map[string]string
}

func (r *Rancher) ListAutoPromServices() ([]*RancherTarget, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	projectId, err := r.getProjectId()
	if err != nil {
		return nil, err
	}
	return r.getRancherTargets(projectId)
}

const rancherFindProjectIdQuery = "https://%s/v2-beta/projects?name=%s"
const autoPromEndpointLabel = "prometheus.endpoint"
const autoPromLabelsPrefix = "prometheus.labels."
const promUri = "http://%v:%v%v"

func formatPromUrl(ip string, port int, endpoint string) string {
	return fmt.Sprintf(promUri, ip, port, endpoint)
}

func (r *Rancher) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(r.token)))
	return http.DefaultClient.Do(req)
}

func (r *Rancher) getProjectId() (string, error) {
	resp, err := r.doRequest(fmt.Sprintf(rancherFindProjectIdQuery, r.host, r.project))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	projectResp := struct {
		Type    string
		Message string
		Data    []struct {
			Id string
		} `json:"data"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&projectResp); err != nil {
		return "", err
	}
	if projectResp.Type == "error" {
		return "", errors.New(projectResp.Message)
	}
	if len(projectResp.Data) == 0 {
		return "", errors.New("not found project " + r.project)
	}
	if len(projectResp.Data) > 1 {
		return "", errors.New("few projects with name " + r.project)
	}
	return projectResp.Data[0].Id, nil
}

func (r *Rancher) getRancherTargets(projectId string) ([]*RancherTarget, error) {
	resp, err := r.doRequest(fmt.Sprintf("https://%s/v2-beta/projects/%s/services", r.host, projectId))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	servicesResp := struct {
		Type    string
		Message string
		Data    []struct {
			Name            string
			Links           struct{ Stack string }
			LaunchConfig    struct{ Labels map[string]string } `json:"launchConfig"`
			PublicEndpoints []struct {
				IpAddress string `json:"ipAddress"`
				Port      int    `json:"port"`
				HostId    string `json:"hostId"`
			} `json:"publicEndpoints"`
		}
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&servicesResp); err != nil {
		return nil, err
	}
	if servicesResp.Type == "error" {
		return nil, errors.New(servicesResp.Message)
	}
	var found []*RancherTarget
	for _, d := range servicesResp.Data {
		promEndpoint, ok := d.LaunchConfig.Labels[autoPromEndpointLabel]
		if !ok {
			continue
		}
		for _, publicEndpoint := range d.PublicEndpoints {
			promUrl := formatPromUrl(publicEndpoint.IpAddress, publicEndpoint.Port, promEndpoint)
			if err := checkPromUrl(promUrl); err != nil {
				log.Println("[WARN] error with " + d.Name + " " + promEndpoint + ": " + err.Error())
				continue
			}
			stackName, err := r.getStackName(d.Links.Stack)
			if err != nil {
				log.Println("[WARN] error with fetch stack name " + d.Name + " : " + err.Error())
				continue
			}
			found = append(found, &RancherTarget{
				Stack:   stackName,
				Service: d.Name,
				Target:  fmt.Sprintf("%v:%v", publicEndpoint.IpAddress, publicEndpoint.Port),
				Host:    publicEndpoint.HostId,
				Labels:  parseLabels(d.LaunchConfig.Labels),
			})
			log.Println(fmt.Sprintf("[INFO] got service %s/%s on %s. %s", stackName, d.Name, publicEndpoint.HostId, promUrl))
		}
	}
	return found, nil
}

func (r *Rancher) getStackName(stackLink string) (string, error) {
	resp, err := r.doRequest(stackLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	stackResp := struct {
		Name string
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&stackResp); err != nil {
		return "", err
	}
	return stackResp.Name, nil
}

func checkPromUrl(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("not 200 code")
	}
	return nil
}

func parseLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range labels {
		if !strings.HasPrefix(k, autoPromLabelsPrefix) {
			continue
		}
		k = strings.TrimPrefix(k, autoPromLabelsPrefix)
		if k == "" {
			continue
		}
		result[k] = v
	}
	return result
}
